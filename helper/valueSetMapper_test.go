package helper_test

import (
    "github.com/stretchr/testify/require"
    "github.com/webshield-dev/eudvcdecoder/helper"
    "testing"
)

const vsDataPath string = "../valuesetdata"

func Test_MAValueSet_Mapper(t *testing.T) {

	type testCase struct {
		name string
		code string
		expectedDisplayName string
	}

	testCases := []testCase{
		{
			name: "should find Moderna",
			code: "ORG-100031184",
			expectedDisplayName: "Moderna Biotech Spain S.L.",
		},
	}

	vsMapper, err := helper.NewValueSetMapper(vsDataPath)
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			require.Equal(t, tc.expectedDisplayName, vsMapper.DecodeMA(tc.code).Display, "should find code")

		})
	}
}

func Test_MPValueSet_Mapper(t *testing.T) {

	type testCase struct {
		name string
		code string
		expectedDisplayName string
	}

	testCases := []testCase{
		{
			name: "should find Moderna",
			code: "EU/1/20/1507",
			expectedDisplayName: "COVID-19 Vaccine Moderna",
		},
	}

	vsMapper, err := helper.NewValueSetMapper(vsDataPath)
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

		    require.Equal(t, tc.expectedDisplayName, vsMapper.DecodeMP(tc.code).Display, "should find code")
		    
        })
	}
}

func Test_VPValueSet_Mapper(t *testing.T) {

	type testCase struct {
		name string
		code string
		expectedDisplayName string
	}

	testCases := []testCase{
		{
			name: "should find SARS-CoV-2",
			code: "1119305005",
			expectedDisplayName: "SARS-CoV-2 antigen vaccine",
		},
	}

	vsMapper, err := helper.NewValueSetMapper(vsDataPath)
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			require.Equal(t, tc.expectedDisplayName, vsMapper.DecodeVP(tc.code).Display, "should find code")

		})
	}
}
