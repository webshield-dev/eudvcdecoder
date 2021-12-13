package verifier

import (
	"context"
	"fmt"
	dhcPdm "github.com/webshield-dev/dhc-common/pdm"
	"github.com/webshield-dev/dhc-common/vaccinemd"
	"github.com/webshield-dev/dhc-common/verification"
	eudvcdatamodel "github.com/webshield-dev/eudvcdecoder/datamodel"
	"github.com/webshield-dev/eudvcdecoder/helper"
)

//
// Verifier verifies a EU Digital Health Credential
//  - decodes it
//  - if possible verifies the signature
//
type Verifier interface {

	//FromFileQRCode verifies a EUDC QR code stored in a .png or .jpg file.
	//if an error returns what it has processed so far, incase want to display
	FromFileQRCode(ctx context.Context, filename string, opts *VerifyOptions) (*Output, error)

	//FromQRCodePNGBytes decode starting with a QR code PNG represented as bytes
	//first makes a local PNG image and the decodes to get the HC1: representation
	//if an error returns what it has processed so far
	FromQRCodePNGBytes(ctx context.Context, pngB []byte, opts *VerifyOptions) (*Output, error)

	//IsDGCFromQRCodeContents returns true if the card is a digital green card, does no processing
	//looks for HCI code
	IsDGCFromQRCodeContents(qrCodeContents []byte) bool

	//FromQRCodeContents decode from the QR code contents, this starts with HC1
	FromQRCodeContents(ctx context.Context, qrCodeContents []byte, opts *VerifyOptions) (*Output, error)
}

//Output the result of a verifier
type Output struct {
	//DecodeOutput the output from decoding
	DecodeOutput *helper.Output

	//Results captures all the verifications that occurred
	Results *verification.CardVerificationResults
}

//DCC return the (Digital Covid Certificate) inside the record, if none returns nil
func (o *Output) DCC() *eudvcdatamodel.DCC {
	if o.DecodeOutput != nil {
		return o.DecodeOutput.DCC()
	}

	return nil
}

//VerifyOptions options to verify
type VerifyOptions struct {
	//UnSafe if set does not verify the signature, somehow the client knows valid.
	UnSafe bool

	//FakeVerificationResultValid if passed in the card verification results will be fake values
	//required for some strange demo situation, do not reuce
	FakeVerificationResultValid bool
}

//NewVerifier make a verifier
func NewVerifier(debug bool, maxDebug bool) (Verifier, error) {

	decoder := helper.NewDecoder(debug, maxDebug)

	return &verifierImpl{debug: debug, maxDebug: maxDebug, decoder: decoder}, nil
}

type verifierImpl struct {
	debug    bool
	maxDebug bool
	decoder  helper.Decoder
}

func (v *verifierImpl) FromFileQRCode(ctx context.Context, filename string, opts *VerifyOptions) (*Output, error) {

	verifyOutput := &Output{}

	//first decode
	decodeOutput, err := v.decoder.FromFileQRCode(filename)
	if err != nil {
		verifyOutput.DecodeOutput = decodeOutput //some decode stages may have passed
		return verifyOutput, fmt.Errorf("error decoding the digital credential err=%s", err)
	}
	verifyOutput.DecodeOutput = decodeOutput
	if !decodeOutput.Decoded {
		//if did not manage to decode then no point in trying to verify
		return verifyOutput, nil
	}

	//verify signature
	if err = v.verify(verifyOutput, opts); err != nil {
		return verifyOutput, fmt.Errorf("error verifying the digital credential err=%s", err)

	}

	return verifyOutput, nil
}

func (v *verifierImpl) FromQRCodePNGBytes(ctx context.Context, pngB []byte, opts *VerifyOptions) (*Output, error) {

	verifyOutput := &Output{}

	//first decode
	decodeOutput, err := v.decoder.FromQRCodePNGBytes(pngB)
	if err != nil {
		verifyOutput.DecodeOutput = decodeOutput //some decode stages may have passed
		return verifyOutput, fmt.Errorf("error decoding the digital credential err=%s", err)
	}
	verifyOutput.DecodeOutput = decodeOutput

	if err = v.verify(verifyOutput, opts); err != nil {
		return verifyOutput, fmt.Errorf("error verifying the digital credential err=%s", err)
	}

	return verifyOutput, nil
}

func (v *verifierImpl) IsDGCFromQRCodeContents(qrCodeContents []byte) bool {
	return v.decoder.IsDGCFromQRCodeContents(qrCodeContents)
}

func (v *verifierImpl) FromQRCodeContents(ctx context.Context, qrCodeContents []byte, opts *VerifyOptions) (*Output, error) {

	verifyOutput := &Output{}

	//first decode
	decodeOutput, err := v.decoder.FromQRCodeContents(qrCodeContents)
	if err != nil {
		verifyOutput.DecodeOutput = decodeOutput //some decode stages may have passed
		return verifyOutput, fmt.Errorf("error decoding the digital credential err=%s", err)
	}
	verifyOutput.DecodeOutput = decodeOutput

	if err = v.verify(verifyOutput, opts); err != nil {
		return verifyOutput, fmt.Errorf("error verifying the digital credential err=%s", err)
	}

	return verifyOutput, nil
}

func (v *verifierImpl) verify(verifyOutput *Output, opts *VerifyOptions) error {

	//
	// Fixme no verifications are currently performed so this state will always been unknown.
	// for product demos (as still in dev) to prospects want to show a valid so allow client
	// to override to fake verification results

	vp := verification.NewProcessor()

	results := vp.GetVerificationResults()

	if opts != nil && opts.FakeVerificationResultValid && results.State != verification.CardVerificationStateValid {
		//fixme comeback and revist, see issue, for now added as upper layers needed to demo code
		//to prospect and need to be able to show as valid for now.
		vp.SetFetchedKey()
		vp.SetSignatureChecked()
		vp.SetSignatureValid()
		vp.SetIssuerTrusted()

		//fake immunization, so does not matter what use
		doses := []*dhcPdm.Dose{
			{
				Coding: vaccinemd.Coding{
					System: vaccinemd.CVXSystem,
					Code:   "207", //moderna
				},

				OccurrenceDateTime: "2021-03-16",
			},
			{
				Coding: vaccinemd.Coding{
					System: vaccinemd.CVXSystem,
					Code:   "207", //moderna
				},

				OccurrenceDateTime: "2021-04-06",
			},
		}

		_, _ = vp.VerifyImmunization(vaccinemd.RegionUSA, doses)

		results = vp.GetVerificationResults()

	}

	verifyOutput.Results = results
	return nil
}
