package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/webshield-dev/eudvcdecoder/helper"
)

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

	dc := helper.NewDecoder(true, true)
	decodeOutput, err := dc.FromFileQRCodePNG(cliQRFilename)
	if err != nil {
		_ = displayResults(decodeOutput, verbose != 0)
		fmt.Printf("ERROR processing certficate err=%s\n", err)
		os.Exit(1)
	}

	if err := displayResults(decodeOutput, verbose != 0); err != nil {
		fmt.Printf("error displaying successful decode err=%s\n", err)
		os.Exit(1)
	}

}

func displayResults(output *helper.Output, verbose bool) error {
	if output == nil {
		return nil
	}

	fmt.Printf("Decoding EU Covid Certificate\n")


	if len(output.DecodedQRCode) != 0 {
		fmt.Printf("  Read QR Code PNG %s Successfully...\n", cliQRFilename)
		if verbose {
			fmt.Printf("    value=%s\n", string(output.DecodedQRCode))
		}
	}

	if len(output.Base45Decoded) != 0 {
		fmt.Printf("  Base45 Decoded Successfully...\n")
		if verbose {
			fmt.Printf("    hex(value)=%s\n", hex.EncodeToString(output.Base45Decoded))
		}
	}

	if len(output.Inflated) != 0 {
		fmt.Printf("  ZLIB Inflated Successfully...\n")
		if verbose {
			fmt.Printf("    hex(value)=%s\n", hex.EncodeToString(output.Inflated))
		}
	}

	if output.CBORUnmarshalledI != nil {
		fmt.Printf("  CBOR UnMarshalled CBOR Web Token (CWT) Successfully...\n")
		if verbose {
			fmt.Printf("    value=%+v\n", output.CBORUnmarshalledI)
		}
	}

	if output.ProtectedHeader != nil {
		fmt.Printf("    CWT CBOR UnMarshalled ProtectedHeader Successfully...\n")
		if verbose {
			fmt.Printf("      value=%+v\n", *output.ProtectedHeader)
		}
	}

	if output.UnProtectedHeader != nil {
		fmt.Printf("    CWT Read UnProtectedHeader Successfully...\n")
		if verbose {
			fmt.Printf("      value=%+v\n", *output.UnProtectedHeader)
		}
	}

	if output.PayloadI != nil {
		fmt.Printf("    CWT CBOR UnMarshalled Payload Successfully...\n")
		if verbose {
			fmt.Printf("      value=%+v\n", output.PayloadI)
		}
	}

	if len(output.COSESignature) != 0 {
		fmt.Printf("    CWT Read COSE Signature (single signer) Successfully...\n")
		if verbose {
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
		//Lets display all the important parts
		//
		fmt.Printf("Decoded EU Covid Certificate\n")


		//
		//Protected header
		//
		prettyResult, err := prettyIdent(output.ProtectedHeader)
		if err != nil {
			return err
		}
		fmt.Printf("protectedHeader=%s\n", prettyResult)

		//
		// Common payload
		//

		prettyResult, err = prettyIdent(output.CommonPayload)
		if err != nil {
			return err
		}
		fmt.Printf("commonPayload=%s\n", prettyResult)

		//
		// Signature in hex
		//
		fmt.Printf("hex(signature)=%s\n", hex.EncodeToString(output.COSESignature))
	}


	return nil

}

func prettyIdent(i interface{}) (string, error) {

	b, err := json.Marshal(i)
	if err != nil {
		return "", err
	}

	dst := &bytes.Buffer{}
	_ = json.Indent(dst, b, "", "  ")

	return dst.String(), nil
}
