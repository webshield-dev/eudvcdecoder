package helper

import (
	"bytes"
	"compress/zlib"
	"fmt"
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
	DecodedQRCode     []byte
	Base45Decoded     []byte
	Inflated          []byte
	CBORUnmarshalledI interface{}
	PayloadI          interface{}
	ProtectedHeader   *COSEHeader
	UnProtectedHeader *COSEHeader
	COSESignature     []byte
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
	Alg int    `cbor:"1,keyasint,omitempty"`
	Kid []byte `cbor:"4,keyasint,omitempty"`
}

type decoderImpl struct {
	debug    bool
	maxDebug bool
}

func (di *decoderImpl) FromFileQRCodePNG(filename string) (*Output, error) {

	output := &Output{}

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
	// CBOR the whole thing to see what is there
	// how to read
	// The
	//   Number: 18  - CBOR Tag this means COSE Signle Signer Data Object -
	//         - see https://datatracker.ietf.org/doc/html/draft-bormann-cbor-notable-tags-01
	//   Contents - is a CBOR array https://datatracker.ietf.org/doc/html/rfc8152#section-2
	//		- The set of protected header parameters wrapped in a bstr.
	//          - This bucket is encoded in the message as a binary object.  This value
	//            is obtained by CBOR encoding the protected map and wrapping it in
	//            a bstr object.
	//          - see https://datatracker.ietf.org/doc/html/rfc8152#section-3
	//      - The set of unprotected header parameters as a map.
	//          - see https://datatracker.ietf.org/doc/html/rfc8152#section-3
	//      - The content of the message.  The content is either the plaintext
	//       or the ciphertext as appropriate.  The content may be detached,
	//       but the location is still used.  The content is wrapped in a bstr
	//       when present and is a nil value when detached
	var dgcI interface{}
	if err := cbor.Unmarshal(inflated, &dgcI); err != nil {
		return err
	}

	outputToPopulate.CBORUnmarshalledI = dgcI

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
		return err
	}

	// Add the unprotected header was a map that did not need more decoding
	outputToPopulate.UnProtectedHeader = &sCWT.Unprotected

	//
	// CBOR decode the protected header
	//
	if len(sCWT.Protected) != 0 {
		var protectedH COSEHeader
		if err := cbor.Unmarshal(sCWT.Protected, &protectedH); err != nil {
			return err
		}
		outputToPopulate.ProtectedHeader = &protectedH
	}

	//
	//CBOR decode the payload into a generic interface that needs to be processed
	//
	var payloadI interface{}
	if err := cbor.Unmarshal(sCWT.Payload, &payloadI); err != nil {
		return err
	}
	outputToPopulate.PayloadI = payloadI

	//
	// Add Signature not used for now
	//
	outputToPopulate.COSESignature = sCWT.Signature

	return nil

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
