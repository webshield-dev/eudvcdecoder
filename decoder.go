package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/webshield-dev/eudvcdecoder/datamodel"

	"github.com/webshield-dev/eudvcdecoder/helper"
)

/*

Decoder will decode and display the contents of a EU COVID-19 Digital Certificate, starting with teh QR code .png


The CLI flags are
1. -qrc_file <value> file containing the qr code png
2. -verbose <level> where level is 0 -> 9, default is zero

Example running with no verbose
- `go run . -qrfile ./testfiles/at_1.png`
- `go run . -qrfile ./testfiles/ie_1_qr.png`

Example running with verbose

    `go run . -qrfile ./testfiles/ie_1_qr.png -verbose 1`

*/

const (
	cliVerboseFlag    = "verbose"
	cliQRFilenameFlag = "qrfile"
)

var (
	cliVerbose    string
	cliQRFilename string
)

// makeFlagSet return flag set needed to start
func makeFlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ExitOnError)

	fs.StringVar(&cliVerbose, cliVerboseFlag, "0", "level of verbose")
	fs.StringVar(&cliQRFilename, cliQRFilenameFlag, "", "qr code (.png) file name")

	return fs
}

func main() {

	fs := makeFlagSet()

	err := fs.Parse(os.Args[1:])
	if err != nil {
		fmt.Printf("error parsing command line flags err=%s\n", err)
		fs.PrintDefaults()
		os.Exit(1)
	}

	verbose, err := strconv.Atoi(cliVerbose)
	if err != nil {
		fmt.Printf("error parsing verbose flag err=%s\n", err)
		fs.PrintDefaults()
		os.Exit(1)

	}

	maxVerbose := verbose > 1
	lowVerbose := verbose == 1

	//set up value set data
	vsDataPath := os.Getenv("VS_DATA_PATH")
	if vsDataPath == "" {
		vsDataPath = "./valuesetdata"
	}
	vsMapper, err := helper.NewValueSetMapper(vsDataPath)
	if err != nil {
		fmt.Printf("error setting up value set mapper err=%s", err)
		os.Exit(1)
	}

	dc := helper.NewDecoder(true, true)

	fmt.Printf("Decoding EU Covid-19 Certificate\n")
	fmt.Printf("  qrCodefile=%s  ValueSetPath=%s  verbose=%d\n", cliQRFilename, vsDataPath, verbose)



	decodeOutput, err := dc.FromFileQRCode(cliQRFilename)
	if err != nil {
		_ = displayResults(vsMapper, decodeOutput, lowVerbose, maxVerbose)
		fmt.Printf("ERROR processing certficate err=%s\n", err)
		os.Exit(1)
	}

	if err := displayResults(vsMapper, decodeOutput, lowVerbose, maxVerbose); err != nil {
		fmt.Printf("error displaying successful decode err=%s\n", err)
		os.Exit(1)
	}

}

func displayResults(vsMapper *helper.ValueSetMapper, output *helper.Output,
	lowVerbose bool, maxVerbose bool) error {
	if output == nil {
		return nil
	}

	if len(output.DecodedQRCode) != 0 {
		fmt.Printf("  Step 1 - Read QR Code PNG %s Successfully...\n", cliQRFilename)
		if maxVerbose {
			fmt.Printf("    value=%s\n", string(output.DecodedQRCode))
		}
	}

	if len(output.Base45Decoded) != 0 {
		fmt.Printf("  Step 2 - Base45 Decoded Successfully...\n")
		if maxVerbose {
			fmt.Printf("    hex(value)=%s\n", hex.EncodeToString(output.Base45Decoded))
		}
	}

	if len(output.Inflated) != 0 {
		fmt.Printf("  Step 3 - ZLIB Inflated Successfully...\n")
		if maxVerbose {
			fmt.Printf("    hex(value)=%s\n", hex.EncodeToString(output.Inflated))
		}
	}

	if output.CBORUnmarshalledI != nil {
		fmt.Printf("  Step 4 - CBOR UnMarshalled CBOR Web Token (CWT) using COSE tagged message COSE Number=%d Successfully...\n",
			output.COSeCBORTag)
		if maxVerbose {
			fmt.Printf("    value=%+v\n", output.CBORUnmarshalledI)
		}
	}

	if output.ProtectedHeader != nil {
		fmt.Printf("    CWT CBOR UnMarshalled the Protected Header Successfully...\n")
		if maxVerbose {
			fmt.Printf("      value=%+v\n", output.ProtectedHeader)
		}
	}

	if output.UnProtectedHeader != nil {
		fmt.Printf("    CWT Read the UnProtected Header Map Successfully...\n")
		if maxVerbose {
			fmt.Printf("      value=%+v\n", *output.UnProtectedHeader)
		}
	}

	if output.PayloadI != nil {
		fmt.Printf("    CWT CBOR UnMarshalled the Payload Successfully...\n")
		if maxVerbose {
			fmt.Printf("      value=%+v\n", output.PayloadI)
		}
	}

	if len(output.COSESignature) != 0 {
		fmt.Printf("    CWT Read the COSE Signature (single signer) Successfully...\n")
		if maxVerbose {
			fmt.Printf("      hex(value)=%s\n", hex.EncodeToString(output.COSESignature))
		}
	}

	if len(output.DiagnoseLines) != 0 {
		for _, line := range output.DiagnoseLines {
			fmt.Printf("%s\n", line)
		}
	}

	if output.CommonPayload != nil {

		//
		// Display details
		//

		//
		//Lets display all the important parts
		//
		fmt.Printf("Successfully Decoded EU Covid-19 Certificate\n")

		if lowVerbose || maxVerbose {
			fmt.Printf("\n**** EU Covid-19 Certificate Details **** \n")

			//
			//Protected header
			//
			prettyResult, err := helper.PrettyIdent(output.ProtectedHeader)
			if err != nil {
				return err
			}
			fmt.Printf("Protected Header=%s\n", prettyResult)

			//
			// Common payload
			//
			prettyPayload, err := helper.PettyIdentCommonPayload(output.CommonPayload)
			if err != nil {
				fmt.Printf("Error pretty printing payload raw=%+v\n", output.PayloadI)
				return err
			}
			fmt.Printf("Common Payload=%s\n", prettyPayload)

			//
			// Signature in hex
			//
			fmt.Printf("hex(signature)=%s\n", hex.EncodeToString(output.COSESignature))
		}

		//
		// Always Display Summary
		//
		displaySummary(vsMapper, output)
	}

	return nil
}

func displaySummary(vsMapper *helper.ValueSetMapper, output *helper.Output) {

	cert := output.CommonPayload.HCERT[datamodel.HCERTMapKeyOne]
	if cert == nil {
		return
	}

	fmt.Printf("\n**** EU Covid-19 Certificate Summary **** \n")

	fmt.Printf("")

	fullName := cert.Name.FullName()

	fmt.Printf("Name:%s\n", fullName)
	fmt.Printf("DOB :%s\n", cert.DOB)

	fmt.Printf("Vaccine Details\n")
	for _, vaccine := range cert.Vaccine {

		//display MP - Medicinal product used for this specific dose of vaccination. A
		maVS := vsMapper.DecodeMA(vaccine.MA)
		mpVS := vsMapper.DecodeMP(vaccine.MP)
		vpVS := vsMapper.DecodeVP(vaccine.VP)

		//convert dosage infomation to ints
		sdI := int64(vaccine.SD)
		dnI := int64(vaccine.DN)

		fmt.Printf("  Doses Administered: %d\n", dnI)
		fmt.Printf("  Doses Required:     %d\n", sdI)
		fmt.Printf("  When:               %s\n", vaccine.DT)
		if mpVS != nil {
			fmt.Printf("  Vaccine Product:    %s\n", mpVS.Display)
		}
		if vpVS != nil {
			fmt.Printf("  Vaccine Type:       %s\n", vpVS.Display)
		}
		if maVS != nil {
			fmt.Printf("  Vaccine Maker:      %s\n", maVS.Display)
		}
		fmt.Printf("  Issuer:             %s\n", vaccine.IS)
		fmt.Printf("  ID:                 %s\n", vaccine.CI)

	}
}
