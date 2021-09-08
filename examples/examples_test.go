package examples_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/webshield-dev/eudvcdecoder/examples"
	euDgcVerifier "github.com/webshield-dev/eudvcdecoder/verifier"
)

func Test_Examples_Can_Be_Verified(t *testing.T) {

	type testCase struct {
		name   string
		qrCode []byte // can pass the numeric code to test
		opts   *euDgcVerifier.VerifyOptions
	}

	testCases := []testCase{
		{
			name:   "Ireland example 1",
			qrCode: examples.GetQRCodeIE1(),
		},
	}

	ctx := context.TODO()

	verifier, err := euDgcVerifier.NewVerifier(true, true)
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

		    //just make sure verify does not fail and so code is good, as other areas test in detail
			verifyOutput, err := verifier.FromQRCodeContents(ctx, tc.qrCode, tc.opts)
			require.NoError(t, err)
            dcc := verifyOutput.DCC()
            require.NotNil(t, dcc, "should contain a card")
		})
	}

}
