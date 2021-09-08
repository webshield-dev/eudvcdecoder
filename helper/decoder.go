package helper

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"fmt"
	"github.com/webshield-dev/eudvcdecoder/datamodel"
	"image"
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
	//DOES not verify
	FromFileQRCodePNG(filename string) (*Output, error)

	//IsDGCFromQRCodeContents returns true if the card is a digital green card, does no processing
	//looks for HCI code
	IsDGCFromQRCodeContents(qrCodeContents []byte) bool

	//FromQRCodeContents decode from the QR code contents, this starts with HC1
	//Does not verify
	FromQRCodeContents(qrCodeContents []byte) (*Output, error)
}

//Output the results of decoding
type Output struct {

	//DecodedQRCode the result of reading the QR code
	DecodedQRCode []byte

	//Base45Decoded the result of base45 decoding the decoded QR code
	Base45Decoded []byte

	//Inflated the result of inflating the base45 decoded qr code
	Inflated []byte

	//COSeCBORTag the message is encoded as a CBOR Tagged Message, this is the TAG from the message.
	//currently only handle COSE_Sign1 which is tag 18 see https://datatracker.ietf.org/doc/html/rfc8152#section-2
	COSeCBORTag uint64

	CBORUnmarshalledI       interface{}
	CBORUnmarshalledPayload []byte //cbor encoded payload
	PayloadI                interface{}
	ProtectedHeader         map[int]interface{} // did not make a COSEHeader as wanted to see what else is inside
	UnProtectedHeader       *datamodel.COSEHeader
	COSESignature           []byte
	CommonPayload           *datamodel.DGCCommonPayload
	DiagnoseLines           []string //if trying to learn display here
}


//DCC return the (Digital Covid Certificate) inside the record, if none returns nil
func (o *Output) DCC() *datamodel.DCC {
	if o.CommonPayload != nil {
		return o.CommonPayload.HCERT.DCC()
	}

	return nil
}

type decoderImpl struct {
	debug    bool
	maxDebug bool
}

//FromFileQRCodePNG reads file that contains the QR code PNG and decodes to get the output,
//DOES NOT verify the signature
func (di *decoderImpl) FromFileQRCodePNG(filename string) (*Output, error) {

	//
	//1. Read QR code image
	//
	qrCodeContents, err := di.readQRCode(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading QR code file=%s err=%s", filename, err)
	}

	return di.FromQRCodeContents(qrCodeContents)
}

//IsDGCFromQRCodeContents returns true if the card is a digital green card, does no processing
//looks for HC1 code
func (di *decoderImpl) IsDGCFromQRCodeContents(qrCodeContents []byte) bool {

	prefix := qrCodeContents[0:3]
	return string(prefix) == datamodel.QRCodePrefix
}

//FromQRCodeContents see interface
func (di *decoderImpl) FromQRCodeContents(qrCodeContents []byte) (*Output, error) {

	output := &Output{
		DiagnoseLines: make([]string, 0),
	}

	output.DecodedQRCode = qrCodeContents

	//
	//2. Base64 Decode
	//

	//remove the HCx: prefix
	base45B := qrCodeContents[4:]
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
	/* #nosec G110 */ //ok as not passed from outside
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

	var sCWT datamodel.SignedCWT
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
		var failProtected datamodel.COSEHeader
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

	var p datamodel.DGCPayloadCBORMapping
	if err := cbor.Unmarshal(sCWT.Payload, &p); err != nil {
		//debug process to understand more
		outputToPopulate.DiagnoseLines = DebugCBORCommonPayload(sCWT.Payload)

		return fmt.Errorf("error cbor unmarshalling common payload run with verbose to see more err=%s", err)
	}

	//create the datamodel version of common payload
	outputToPopulate.CommonPayload = &datamodel.DGCCommonPayload{}
	outputToPopulate.CommonPayload.Populate(&p)

	//
	// Add Signature not used for now
	//
	outputToPopulate.COSESignature = sCWT.Signature

	return nil

}

func (di *decoderImpl) readQRCode(filename string) (decodedQRCode []byte, err error) {

	// open and decode image file
	var file *os.File
	file, err = os.Open(os.ExpandEnv(filename))
	if err != nil {
		return nil, err
	}
	defer func() {
		err1 := file.Close()
		if err == nil {
			err = err1
		}
	}()

	var img image.Image
	img, err = png.Decode(file)
	if err != nil {
		return nil, err
	}

	// prepare BinaryBitmap
	var bmp *gozxing.BinaryBitmap
	bmp, err = gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil, err
	}

	// decode image
	var result *gozxing.Result
	qrReader := qrcode.NewQRCodeReader()
	result, err = qrReader.Decode(bmp, nil)
	if err != nil {
		return nil, err
	}

	return []byte(result.String()), err

}
