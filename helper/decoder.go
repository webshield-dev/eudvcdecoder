package helper

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"fmt"
	"github.com/webshield-dev/eudvcdecoder/datamodel"
	"image/png"
	"io"
	"os"

	"github.com/fxamacker/cbor/v2"

	"github.com/dasio/base45"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

//
// Decode a EU Digital COVID Certificate
//

/*
The decoding steps are as follows
1. Read the QR code .png containing the Digital Certificate to get a base45 encoded certificate
2. Decode the base45 certificate to get a compressed certificate
3. ZLIB inflate the compressed certificate to get a CBOR Web Token
4. CBOR decode the CBOR Web Token to get the protected header, unprotected header, payload, and signature
5. CBOR decode the protected header to get the Signing Algorithm and KeyID
6. CBOR decode the payload to get the issuer, iat, exp, subject information, and vaccination information
7. NOT implemented check the COSE signature by getting signing key from issuing State and using it to check the CBOR signature.

*/

//NewDecoder make a decoder
func NewDecoder(debug bool, maxDebug bool) Decoder {
	return &decoderImpl{debug: debug, maxDebug: maxDebug}
}

//Decoder methods to decode a EU covid certificate
type Decoder interface {

	//FromFileQRCodePNG decode starting with a QR code PNG stored in filename
	//if an error returns what it has processed so far
	FromFileQRCodePNG(filename string) (*Output, error)
}

//Output the results of decoding
type Output struct {
	DecodedQRCode []byte
	Base45Decoded []byte
	Inflated      []byte

	//COSeCBORTag	 credential a CBOR tagged message currently can only handle COSE_Sign1 which is 18
	// see https://datatracker.ietf.org/doc/html/rfc8152#section-2
	COSeCBORTag uint64

	CBORUnmarshalledI       interface{}
	CBORUnmarshalledPayload []byte //cbor encoded payload
	PayloadI                interface{}
	ProtectedHeader         map[int]interface{} // from spec
	UnProtectedHeader       *COSEHeader
	COSESignature           []byte
	CommonPayload           *datamodel.DGCCommonPayload
	DiagnoseLines           []string //if trying to learn display here
}

//COSEHeader only contains what is specified in the vaccine credential
//https://ec.europa.eu/health/sites/default/files/ehealth/docs/digital-green-certificates_v3_en.pdf
// see CBOR https://datatracker.ietf.org/doc/html/rfc8152#section-3.1 for where 1 and 4 come from
//  Generic_Headers = (
//       ? 1 => int / tstr,  ; algorithm identifier
//       ? 2 => [+label],    ; criticality
//       ? 3 => tstr / int,  ; content type
//       ? 4 => bstr,        ; key identifier
//       ? 5 => bstr,        ; IV
//       ? 6 => bstr,        ; Partial IV
//       ? 7 => COSE_Signature / [+COSE_Signature] ; Counter signature
//
type COSEHeader struct {
	//Alg can be an int or a utf-8 string so make an interface
	Alg interface{} `cbor:"1,keyasint,omitempty"` // this mapping is incorrect as can be utf-8 string or an int
	Kid []byte      `cbor:"4,keyasint,omitempty"`
}

type decoderImpl struct {
	debug    bool
	maxDebug bool
}

func (di *decoderImpl) FromFileQRCodePNG(filename string) (*Output, error) {

	output := &Output{
		DiagnoseLines: make([]string, 0),
	}

	//
	//1. Read QR code image
	//

	decodedQRCode, err := di.readQRCode(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading QR code file=%s err=%s", filename, err)
	}
	output.DecodedQRCode = decodedQRCode

	//
	//2. Base64 Decode
	//

	//remove the HCx: prefix
	base45B := decodedQRCode[4:]
	base45Decoded, err := base45.DecodeString(string(base45B))
	if err != nil {
		return output, err
	}
	output.Base45Decoded = base45Decoded

	//
	//3. Inflate
	//

	reader := bytes.NewReader(base45Decoded)
	zlibReader, err := zlib.NewReader(reader)
	if err != nil {
		return output, err
	}

	inflated := new(bytes.Buffer)
	_, err = io.Copy(inflated, zlibReader)
	if err != nil {
		return output, err
	}
	output.Inflated = inflated.Bytes()

	//
	//4. CBOR decode the CBOR Web Token to get the protected header, unprotected header, payload, and signature
	//
	//
	if err := di.cborUnMarshall(inflated.Bytes(), output); err != nil {
		return output, err
	}

	return output, nil

}

func (di *decoderImpl) cborUnMarshall(inflated []byte, outputToPopulate *Output) error {

	//
	// Is a CBOR tagged message that has a tag to define what type of message,
	// see https://datatracker.ietf.org/doc/html/rfc8152#section-2 for the COSE structure
	// held in th Conents
	//
	//
	// The
	//   Number: 18  - CBOR Tag this means COSE Single Signer Data Object -
	//         - see https://datatracker.ietf.org/doc/html/draft-bormann-cbor-notable-tags-01
	//   Contents - The COSE object structure is a CBOR array https://datatracker.ietf.org/doc/html/rfc8152#section-2
	//   elements are
	//		[0] The set of protected header parameters wrapped in a bstr.
	//          - This bucket is encoded in the message as a binary object.  This value
	//            is obtained by CBOR encoding the protected map and wrapping it in
	//            a bstr object.
	//          - see https://datatracker.ietf.org/doc/html/rfc8152#section-3
	//      [1] The set of unprotected header parameters as a map.
	//          - see https://datatracker.ietf.org/doc/html/rfc8152#section-3
	//      [2] The content of the message.  The content is either the plaintext
	//       or the ciphertext as appropriate.  The content may be detached,
	//       but the location is still used.  The content is wrapped in a bstr
	//       when present and is a nil value when detached
	//

	var taggedMessage cbor.Tag
	if err := cbor.Unmarshal(inflated, &taggedMessage); err != nil {
		return fmt.Errorf("error unmarshalling inflated CWT into an interface{} err=%s", err)
	}
	outputToPopulate.COSeCBORTag = taggedMessage.Number
	outputToPopulate.CBORUnmarshalledI = taggedMessage

	//must be a COSE_Sign1 otherwise cannot read signature
	if taggedMessage.Number != 18 {
		return fmt.Errorf("error CBOR tagged message number must be 18 got=%d", taggedMessage.Number)
	}

	//
	// Unpack using some structs
	//

	type signedCWT struct {
		_ struct{} `cbor:",toarray"`
		// this seems to be cbor encoded if make a coseHeader then fails with
		// cannot unmarshal byte string into Go struct field encoding_test.signedCWT.Protected of type encoding_test.coseHeader
		// when cbor.Unmarshal the whole web token
		// The set of protected headers wrapped in a byte string - see https://datatracker.ietf.org/doc/html/rfc8152#section-2
		// needs to be CBOR decoded
		//CBOR encoding of the map of protected headers, that is wrapped in a byte string
		//see https://datatracker.ietf.org/doc/html/rfc8152#section-3 and section-2
		Protected []byte // this seems to be cbor encoded

		//  Set of unprotected header  parameters as a map
		// see https://datatracker.ietf.org/doc/html/rfc8152#section-3
		Unprotected COSEHeader

		//The CBOR encoded content as a byte string, needs to be CBOR decoded
		Payload []byte

		//The COSE signature - is a singe signer
		Signature []byte
	}

	var sCWT signedCWT
	if err := cbor.Unmarshal(inflated, &sCWT); err != nil {
		return fmt.Errorf("error unmarshalling inflated CWT into an CWT struct err=%s", err)
	}

	// Add the unprotected header was a map that did not need more decoding
	outputToPopulate.UnProtectedHeader = &sCWT.Unprotected

	//
	// CBOR decode the protected header
	//
	if len(sCWT.Protected) != 0 {
		var protectedI map[int]interface{}
		if err := cbor.Unmarshal(sCWT.Protected, &protectedI); err != nil {
			return fmt.Errorf("error cbor.Unmarshal protected header hex=%s err=%s",
				hex.EncodeToString(sCWT.Protected), err)
		}
		outputToPopulate.ProtectedHeader = protectedI

		/*
		//
		// DEBUG dump the protected header types
		//
		for k, v := range protectedI {
			fmt.Printf("**** DEBUG PROTECTED HEADER k=%d v=%v t=%T\n", k, v, v)
		}*/

		//added this as sometimes found issues and this is a way to further check
		//fixme why not set protected header to this type?
		var failProtected COSEHeader
		if err := cbor.Unmarshal(sCWT.Protected, &failProtected); err != nil {
			return fmt.Errorf("error cbor.Unmarshal protected header hex=%s err=%s",
				hex.EncodeToString(sCWT.Protected), err)
		}

	}

	//
	//CBOR decode the payload into a generic interface that needs to be processed
	//
	outputToPopulate.CBORUnmarshalledPayload = sCWT.Payload
	var payloadI interface{}
	if err := cbor.Unmarshal(sCWT.Payload, &payloadI); err != nil {
		return err
	}
	outputToPopulate.PayloadI = payloadI

	//
	// CBOR unmarshall the Payload into the common payload CBOR mapping as defined on section 2.6.3 in
	// https://ec.europa.eu/health/sites/default/files/ehealth/docs/digital-green-certificates_v3_en.pdf
	// also see CWT for CBOR mapping of iss, exp, iat
	// https://datatracker.ietf.org/doc/html/rfc8392#section-4
	type commonPayloadCBORMapping struct {
		ISS   string             `cbor:"1,keyasint,omitempty"`
		EXP   uint64             `cbor:"4,keyasint,omitempty"`
		IAT   uint64             `cbor:"6,keyasint,omitempty"`
		HCERT datamodel.HCERTMap `cbor:"-260,keyasint,omitempty"`
	}

	var p commonPayloadCBORMapping
	if err := cbor.Unmarshal(sCWT.Payload, &p); err != nil {
		//debug process to understand more
		outputToPopulate.DiagnoseLines = DebugCBORCommonPayload(sCWT.Payload)

		return fmt.Errorf("error cbor unmarshalling common payload run with verbose to see more err=%s", err)
	}

	//create the datamodel version of common payload, just did not want to expose CBOR mapping outside of here
	outputToPopulate.CommonPayload = &datamodel.DGCCommonPayload{
		ISS:   p.ISS,
		IAT:   p.IAT,
		EXP:   p.EXP,
		HCERT: p.HCERT}

	//
	// Add Signature not used for now
	//
	outputToPopulate.COSESignature = sCWT.Signature

	return nil

}

func (di *decoderImpl) debugCommonPayload(payload []byte) []string {

	rl := make([]string, 0)

	rl = append(rl, fmt.Sprintf("ERROR cbor unmarshalling CommonPayload HCERT diagnosing"))

	//if an error using the known types then use an interface for HCERT so can process in debug
	type resilientCommonPayloadCBORMapping struct {
		ISS   string      `cbor:"1,keyasint,omitempty"`
		EXP   uint64      `cbor:"4,keyasint,omitempty"`
		IAT   uint64      `cbor:"6,keyasint,omitempty"`
		HCERT interface{} `cbor:"-260,keyasint,omitempty"`
	}

	var cp resilientCommonPayloadCBORMapping
	if err := cbor.Unmarshal(payload, &cp); err != nil {
		return append(rl, fmt.Sprintf("error debugging cbor payload unmarshall error err=%s", err))
	}

	switch cp.HCERT.(type) {

	case map[interface{}]interface{}:
		{
			hcertM := cp.HCERT.(map[interface{}]interface{})
			for k, v := range hcertM {
				switch k.(type) {
				case uint64:
					{
						ki := k.(uint64)
						if ki == 1 {
							//can process
							arls := AnalyseMap(v, "  ")
							for _, arl := range arls {
								rl = append(rl, arl)
							}

						} else {
							rl = append(rl, fmt.Sprintf("ERROR HCERT.map[key] expected=1 got=%d", ki))
						}

					}

				default:
					{

					}
					rl = append(rl, fmt.Sprintf("ERROR HCERT.map[key] expected=uint64 got=%T", k))
				}
			}
		}

	default:
		{
			rl = append(rl, fmt.Sprintf("HCERT expected map[interface{}]interface{} got=%T", cp.HCERT))
		}

	}

	return rl

}

func (di *decoderImpl) readQRCode(filename string) ([]byte, error) {

	// open and decode image file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, err
	}

	// prepare BinaryBitmap
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil, err
	}

	// decode image
	qrReader := qrcode.NewQRCodeReader()
	result, err := qrReader.Decode(bmp, nil)
	if err != nil {
		return nil, err
	}

	return []byte(result.String()), nil

}
