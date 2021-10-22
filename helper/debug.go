package helper

import (
	"fmt"
	"github.com/fxamacker/cbor/v2"
)

//debug routines when get cast error

//DebugCBORCommonPayload display type and values
func DebugCBORCommonPayload(payload []byte) []string {

	rl := make([]string, 0)

	rl = append(rl, "ERROR cbor unmarshalling CommonPayload HCERT diagnosing")

	//if an error using the known types then use an interface for HCERT so can process in debug
	type resilientCommonPayloadCBORMapping struct {
		ISS   string      `cbor:"1,keyasint,omitempty"`
		EXP   uint64      `cbor:"4,keyasint,omitempty"`
		IAT   uint64      `cbor:"6,keyasint,omitempty"`
		HCERT interface{} `cbor:"-260,keyasint,omitempty"`
	}

	var cp resilientCommonPayloadCBORMapping
	if err := cbor.Unmarshal(payload, &cp); err != nil {
		return append(rl, fmt.Sprintf("error debugging cbor payload unmarshall error err=%s", err))
	}

	switch cp.HCERT.(type) {

	case map[interface{}]interface{}:
		{
			hcertM := cp.HCERT.(map[interface{}]interface{})
			for k, v := range hcertM {
				switch kt := k.(type) {
				case uint64:
					{
						ki := kt
						if ki == 1 {
							//can process
							arls := AnalyseMap(v, "  ")
							rl = append(rl, arls...)

						} else {
							rl = append(rl, fmt.Sprintf("ERROR HCERT.map[key] expected=1 got=%d", ki))
						}

					}

				default:
					{

					}
					rl = append(rl, fmt.Sprintf("ERROR HCERT.map[key] expected=uint64 got=%T", k))
				}
			}
		}

	default:
		{
			rl = append(rl, fmt.Sprintf("HCERT expected map[interface{}]interface{} got=%T", cp.HCERT))
		}

	}

	return rl

}

//AnalyseMap kept finding that different issues use different types so created this routine to scan
//whole structure and print out key, key.(type), value and value.(type) to help diagnose. Not can always
//see content if use -verbose 1 as does not try to JSON unmarshall
func AnalyseMap(mapI interface{}, indent string) []string {

	rl := make([]string, 0)

	switch mapI.(type) {

	case map[interface{}]interface{}:
		{
			//can process
			rl = append(rl, fmt.Sprintf("%sADD HCERT processing", indent))
			m1 := mapI.(map[interface{}]interface{})
			for k1, v1 := range m1 {
				rl = append(rl, fmt.Sprintf("%skey=%v %T v=%v %T", indent, k1, k1, v1, v1))

				switch v1t := v1.(type) {
				case map[interface{}]interface{}, map[string]interface{}:
					{
						newrls := AnalyseMap(v1, indent+"  ")
						rl = append(rl, newrls...)
					}
				case []interface{}:
					{
						a := v1t
						for _, entry := range a {
							newrls := AnalyseMap(entry, indent+"  ")
							rl = append(rl, newrls...)
						}

					}
				default:
					{
						//no more work
					}
				}
			}

		}

	default:
		rl = append(rl, fmt.Sprintf(
			"ERROR expected map[interface{}]interface{} got=%T", mapI))

	}

	return rl

}
