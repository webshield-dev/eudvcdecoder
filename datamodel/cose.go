package datamodel

//COSEHeader only contains what is specified in the vaccine credential
//https://ec.europa.eu/health/sites/default/files/ehealth/docs/digital-green-certificates_v3_en.pdf
// see CBOR https://datatracker.ietf.org/doc/html/rfc8152#section-3.1 for where 1 and 4 come from
//  Generic_Headers = (
//       ? 1 => int / tstr,  ; algorithm identifier
//       ? 2 => [+label],    ; criticality
//       ? 3 => tstr / int,  ; content type
//       ? 4 => bstr,        ; key identifier
//       ? 5 => bstr,        ; IV
//       ? 6 => bstr,        ; Partial IV
//       ? 7 => COSE_Signature / [+COSE_Signature] ; Counter signature
//
type COSEHeader struct {
	Alg int     `cbor:"1,keyasint,omitempty"`

	//Kid has a []byte but ran into issue with unmarshalled so changed to a unit8
	Kid []uint8 `cbor:"4,keyasint,omitempty"`
}

//SignedCWT the CBOR web token (CWT) see https://datatracker.ietf.org/doc/html/rfc8392
type SignedCWT struct {
    _ struct{} `cbor:",toarray"`
    // this seems to be cbor encoded if make a coseHeader then fails with
    // cannot unmarshal byte string into Go struct field encoding_test.signedCWT.Protected of type encoding_test.coseHeader
    // when cbor.Unmarshal the whole web token
    // The set of protected headers wrapped in a byte string - see https://datatracker.ietf.org/doc/html/rfc8152#section-2
    // needs to be CBOR decoded
    //CBOR encoding of the map of protected headers, that is wrapped in a byte string
    //see https://datatracker.ietf.org/doc/html/rfc8152#section-3 and section-2
    Protected []byte // this seems to be cbor encoded

    //  Set of unprotected header  parameters as a map
    // see https://datatracker.ietf.org/doc/html/rfc8152#section-3
    Unprotected COSEHeader

    //The CBOR encoded content as a byte string, needs to be CBOR decoded
    Payload []byte

    //The COSE signature - is a singe signer
    Signature []byte
}
