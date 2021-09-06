package verifier_test

import (
	"github.com/stretchr/testify/require"
	"github.com/webshield-dev/eudvcdecoder/verifier"
	"testing"
)

//Test_Verifier - cursory tests and main tests are in the decoder
func Test_Verifier(t *testing.T) {

	type testCase struct {
		name string

		qrCodePath string
	}

	//
	//test data https://github.com/eu-digital-green-certificates/dgc-testdata
	//
	testCases := []testCase{
		{
			name:       "should decode a WebShield generated file",
			qrCodePath: "../testfiles/vaccine/ws_generate_qrcode.png",
		},
		{
			name:       "should support ireland vaccine qr code",
			qrCodePath: "../testfiles/dcc-testdata/IE/png/1_qr.png",
		},
		{
			name:       "should support greece test qr code png",
			qrCodePath: "../testfiles/dcc-testdata/GR/2DCode/png/3.png",
		},
		{
			name:       "should support NL vaccine qr code png",
			qrCodePath: "../testfiles/dcc-testdata/NL/png/072-NL-vaccination.png",
		},
		{
			name:       "should support German Vaccine qr code png",
			qrCodePath: "../testfiles/dcc-testdata/DE/2DCode/png/1.png",
		},
		{
			name:       "should support austria vaccine qr code png",
			qrCodePath: "../testfiles/dcc-testdata/AT/png/1.png",
		},
	}

	dgVerifier, err := verifier.NewVerifier(true, true)
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			decodeOutput, err := dgVerifier.FromFileQRCodePNG(tc.qrCodePath)
			require.NoError(t, err)
			require.NotNil(t, decodeOutput)
		})
	}
}
