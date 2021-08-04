package helper_test

import (
    "github.com/stretchr/testify/require"
    "github.com/webshield-dev/eudvcdecoder/helper"
    "testing"
)

func Test_Decode(t *testing.T) {

    type testCase struct {
        name string

        qrCodePath string

        //  "CBOR": **CBOR (hex encoded)**,
        cbor string

        // "COSE": **COSE (hex encoded)**,
        cose string

        // **COMPRESSED (hex encoded)**,
        compressed string

        //**BASE45 Encoded compressed COSE**,
        base45 string
    }

    //
    //test data https://github.com/eu-digital-green-certificates/dgc-testdata
    //
    testCases := []testCase{
        {
            name:       "ireland test qr code",
            qrCodePath: "../testfiles/ireland_1_qr.png",
            base45: "HC1:NCFE70X90T9WTWGVLKX49LDA:4NX35 CPX*42BB3XK2F3U7PF9I2F3Z:N3 Q6JC X8Y50.FK6ZK7:EDOLFVC*70B$D% D3IA4W5646946846.966KCN9E%961A6DL6FA7D46XJCCWENF6OF63W5NW6C46WJCT3E$B9WJC0FDTA6AIA%G7X+AQB9746QG7$X8SW6/TC4VCHA7LB7$471S6N-COA7X577:6 47F-CZIC6UCF%6AK4.JCP9EJY8L/5M/5546.96VF6%JCJQEK69WY8KQEPD09WEQDD+Q6TW6FA7C46TPCBEC8ZKW.C8WE7H801AY09ZJC2/D*H8Y3EN3DMPCG/DOUCNB8WY8I3DOUCCECZ CO/EZKEZ964461S6GVC*JC1A6$473W59%6D4627BPFL .4/FQQRJ/2519D+9D831UT8D4KB82JP63-G$C4/1B2SMHXDW2V:CSU6NJIO4U0-T6573C+DM-FARF9.3KMF+PVCBD$%K-4PKOE",

        },
    }

    vcDecoder := helper.NewDecoder(true, true)

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {

            decodeOutput, err := vcDecoder.FromFileQRCodePNG(tc.qrCodePath)
            require.NoError(t, err)
            require.Equal(t, tc.base45, string(decodeOutput.DecodedQRCode), "should match")

        })
    }
}
