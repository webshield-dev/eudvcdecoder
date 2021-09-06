package verifier

import (
    "fmt"
    "github.com/webshield-dev/eudvcdecoder/helper"
)

//
// Verifier verifies a EU Digital Health Credential
//  - decodes it
//  - if possible verifies the signature
//
type Verifier interface {


    //FromFileQRCodePNG verifies a EUDC from its QR code PNG stored in filename
    //if an error returns what it has processed so far, incase want to display
    FromFileQRCodePNG(filename string) (*Output, error)

}

//Output the result of a verifier
type Output struct {
    //DecodeOutput the output from decoding
    DecodeOutput *helper.Output

    //VerifiedSignature true if the signature has been verified
    VerifiedSignature bool
}

//NewVerifier make a verifier
func NewVerifier(debug bool, maxDebug bool) Verifier {

    decoder := helper.NewDecoder(debug, maxDebug)

    return &verifierImpl{debug: debug, maxDebug: maxDebug, decoder: decoder}
}



type verifierImpl struct {
    debug    bool
    maxDebug bool
    decoder helper.Decoder
}


func (v *verifierImpl) FromFileQRCodePNG(filename string) (*Output, error) {

    verifyOutput := &Output{}

    //first decode
    decodeOutput, err := v.decoder.FromFileQRCodePNG(filename)
    if err != nil {
        verifyOutput.DecodeOutput = decodeOutput //some decode stages may have passed
        return verifyOutput, fmt.Errorf("error decoding the digital credential err=%s", err)
    }
    verifyOutput.DecodeOutput = decodeOutput

    //
    // fixme add code to verify the signature
    //
    verifyOutput.VerifiedSignature = false

    return verifyOutput, nil
}


