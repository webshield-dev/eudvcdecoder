package helper_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"github.com/webshield-dev/eudvcdecoder/datamodel"
	"github.com/webshield-dev/eudvcdecoder/helper"
	"testing"
)

type dccTestData struct {
	JSON   *datamodel.DCC
	Prefix string `json:"PREFIX"`
}

func Test_Decode(t *testing.T) {

	type testCase struct {
		name string

		qrCodePath string
		jsonPath   string
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
			jsonPath:   "../testfiles/dcc-testdata/IE/2DCode/Raw/1.json",
		},
		{
			name:       "should support greece test qr code png",
			qrCodePath: "../testfiles/dcc-testdata/GR/2DCode/png/1.png",
			jsonPath:   "../testfiles/dcc-testdata/GR/2DCode/raw/1.json",
		},
		{
			name:       "should support NL vaccine qr code png",
			qrCodePath: "../testfiles/dcc-testdata/NL/png/041-NL-vaccination.png",
			jsonPath:   "../testfiles/dcc-testdata/NL/2DCode/raw/041-NL-vaccination.json",
		},
		{
			name:       "should support German Vaccine qr code png",
			qrCodePath: "../testfiles/dcc-testdata/DE/2DCode/png/1.png",
			jsonPath:   "../testfiles/dcc-testdata/DE/2DCode/raw/1.json",
		},
		{
			name:       "should support austria vaccine qr code png",
			qrCodePath: "../testfiles/dcc-testdata/AT/png/1.png",
			jsonPath:   "../testfiles/dcc-testdata/AT/2DCode/raw/1.json",
		},
	}

	vcDecoder := helper.NewDecoder(true, true)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			decodeOutput, err := vcDecoder.FromFileQRCodePNG(tc.qrCodePath)
			require.NoError(t, err)
			require.NotNil(t, decodeOutput)
			require.True(t, decodeOutput.Decoded, "should have successfully decoded data")
			require.NotEmpty(t, decodeOutput.DecodedQRCode, "should have decoded QR code")

			require.True(t, vcDecoder.IsDGCFromQRCodeContents(decodeOutput.DecodedQRCode), "should be a DGC")

			if tc.jsonPath != "" {
				jsonB, err := helper.ReadData(tc.jsonPath)
				require.NoError(t, err)
				var testData dccTestData
				err = json.Unmarshal(jsonB, &testData)
				require.NoError(t, err)
				dcc := decodeOutput.DCC()
				require.Equal(t, *testData.JSON, *dcc)

				if testData.Prefix != "" {
                    require.Equal(t, testData.Prefix, string(decodeOutput.DecodedQRCode), "base45 decoded should match")

                }
			}

		})
	}
}
