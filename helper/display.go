package helper

import (
	"bytes"
	"encoding/json"
    "fmt"
    "github.com/webshield-dev/eudvcdecoder/datamodel"
)

// display helpers

//PettyIdentCommonPayload pretty ident
func PettyIdentCommonPayload(dcp *datamodel.DGCCommonPayload) (string, error) {

    //
    // for now can only pretty print Vaccine certs
    //
    cert := dcp.HCERT[datamodel.HCERTMapKeyOne]
    if len(cert.Vaccine) == 0 {
        return "No Vaccine Results may be a Test or a Recovery credential", nil
    }

    result,  err := PrettyIdent(dcp)
    if err != nil {
        cert := dcp.HCERT[datamodel.HCERTMapKeyOne]
        return "", fmt.Errorf("error pretty indenting common payload err=%s  payload=%+v cert=%+v",
            err, *dcp, cert)
    }

    return result, nil
}

//PrettyIdent pretty ident some json
func PrettyIdent(i interface{}) (string, error) {

	b, err := json.Marshal(i)
	if err != nil {
		return "", err
	}

	dst := &bytes.Buffer{}
	_ = json.Indent(dst, b, "", "  ")

	return dst.String(), nil
}
