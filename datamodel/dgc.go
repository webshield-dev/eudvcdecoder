package datamodel


//QRCodePrefix the DGC is prefixed with this before converting into a QR code PNG
const QRCodePrefix = "HC1"

//DCC (Digital Covid Certificate) JSON schema
// see
//  - https://github.com/ehn-dcc-development/ehn-dcc-schema
//  - https://ec.europa.eu/health/sites/default/files/ehealth/docs/covid-certificate_json_specification_en.pdf
type DCC struct {
	Version  string      `json:"ver"`
	DOB      string      `json:"dob"`
	Name     Name        `json:"nam,omitempty"`
	Vaccine  []Vaccine   `json:"v,omitempty"`
	Test     interface{} `json:"t,omitempty"`
	Recovery interface{} `json:"r,omitempty"`
}

//Name as defined in https://github.com/ehn-dcc-development/ehn-dcc-schema
type Name struct {

	//FN Surname(s), such as family name(s), of the holder.
	//Exactly 1 (one) non-empty field MUST be provided, with
	//all surnames included in it. In case of multiple surnames,
	//these MUST be separated by a space. Combination names
	//including hyphens or similar characters must however
	//stay the same.
	FN string `json:"fn,omitempty"`

	//FNT Surname(s) of the holder transliterated using the same
	//convention as the one used in the holder’s machine
	//readable travel documents (such as the rules defined in
	//ICAO Doc 9303 Part 3).
	//Exactly 1 (one) non-empty field MUST be provided, only
	//including characters A-Z and <. Maximum length: 80
	//characters (as per ICAO 9303 specification).
	FNT string `json:"fnt,omitempty"`

	//GN Forename(s), such as given name(s), of the holder.
	//If the holder has no forenames, the field MUST be
	//skipped.
	//In all other cases, exactly 1 (one) non-empty field MUST
	//be provided, with all forenames included in it. In case of
	//multiple forenames, these MUST be separated by a space.
	GN string `json:"gn,omitempty"`

	//GNT Forename(s) of the holder transliterated using the same
	//convention as the one used in the holder’s machine
	//readable travel documents (such as the rules defined in
	//ICAO Doc 9303 Part 3).
	//If the holder has no forenames, the field MUST be
	//skipped.In all other cases, exactly 1 (one) non-empty field MUST
	//be provided, only including characters A-Z and <.
	//Maximum length: 80 characters.
	GNT string `json:"gnt,omitempty"`
}

//Vaccine Vaccination group, if present, MUST contain exactly 1 (one) entry describing exactly one vaccination
//event. All elements of the vaccination group are mandatory, empty values are not supported.
type Vaccine struct {

	//TGA coded value from the value set disease-agent-targeted.json.
	TG string `json:"tg,omitempty"`

	//VP Type of the vaccine or prophylaxis used.
	VP string `json:"vp,omitempty"`

	//MP Medicinal product used for this specific dose of vaccination. A coded value
	//from the value set vaccine-medicinal-product.json.
	MP string `json:"mp,omitempty"`

	//MA Marketing authorisation holder or manufacturer, if no marketing authorization
	//holder is present. A coded value from the value set vaccine-mah-manf.json.
	MA string `json:"ma,omitempty"`

	//DN Sequence number (positive integer) of the dose given during this vaccination
	//event. 1 for the first dose, 2 for the second dose etc.
	//"dn": "1" (first dose in a series)
	//"dn": "2" (second dose in a series)
	//"dn": "3" (third dose in a series, such as in case of a booster)
	//
	// Although spec says int have found some issuers use float and some use int so use float as will
	// work for both
	DN float64 `json:"dn,omitempty"`

	//SD Total number of doses (positive integer) in a complete vaccination series
	//according to the used vaccination protocol. The protocol is not in all cases
	//directly defined by the vaccine product, as in some countries only one dose of
	//normally two-dose vaccines is delivered to people recovered from COVID19. In these cases,
	//the value of the field should be set to 1.
	//"sd": "1" (for all 1-dose vaccination schedules)
	//"sd": "2" (for 2-dose vaccination schedules)
	//"sd": "3" (in case of a booster)
	//
	// Although spec says int have found some issuers use float and some use int so use float as will
	// work for both
	SD float64 `json:"sd,omitempty"`

	//DT The date when the described dose was received, in the format YYYY-MM-DD
	//(full date without time). Other formats are not supported
	DT string `json:"dt,omitempty"`

	//CO Country expressed as a 2-letter ISO3166 code (RECOMMENDED) or a
	//reference to an international organisation responsible for the vaccination event
	//(such as UNHCR or WHO). A coded value from the value set country-2-codes.json.
	CO string `json:"co,omitempty"`

	//IS Name of the organisation that issued the certificate. Identifiers are allowed as
	//part of the name, but not recommended to be used individually without the
	//name as a text. Max 80 UTF-8 characters.
	//Exactly 1 (one) non-empty field MUST be provided. Example:
	//"is": "Ministry of Health of the Czech Republic"
	//"is": "Vaccination Centre South District 3"
	IS string `json:"is,omitempty"`

	//CI Unique certificate identifier (UVCI) as specified in the vaccinationproof_interoperability-guidelines_en.pdf (europa.eu)
	//The inclusion of the checksum is optional. The prefix "URN:UVCI:" may be
	//added.
	CI string `json:"ci,omitempty"`
}

//HCERTMap looking a unmarshalled CBOR this is a map with one key "1" that is the DCC
//see https://ec.europa.eu/health/sites/default/files/ehealth/docs/digital-green-certificates_v3_en.pdf
type HCERTMap map[uint64]*DCC


//DCC return the DCC in the map
func (h HCERTMap) DCC() *DCC{
	return h[HCERTMapKeyOne]
}

//HCERTMapKeyOne not sure if there are planned extensions with other keys so did not want to collapse out
//of the model for now
const HCERTMapKeyOne uint64 = 1

//DGCCommonPayload the common payload defined in
// https://ec.europa.eu/health/sites/default/files/ehealth/docs/digital-green-certificates_v3_en.pdf
type DGCCommonPayload struct {
	//ISS Issuer of the DGC
	ISS string `json:"iss"`

	//IAT Issuing Date of the DGC
	IAT uint64 `json:"iat"`

	//EXP Expiring Date of the DGC
	EXP uint64 `json:"exp"`

	//HCERT Payload of the DGC can be a vaccine, test, or recovery
	HCERT HCERTMap `json:"hcert"`
}

//Populate for a cbor mapped source
func (dcp *DGCCommonPayload) Populate(source *DGCPayloadCBORMapping) {
	dcp.ISS = source.ISS
	dcp.IAT = source.IAT
	dcp.EXP = source.EXP
	dcp.HCERT = source.HCERT

}

//DGCPayloadCBORMapping extract the CBOR mapping from the JSON mapping just in case later need
//to treat differently
// CBOR unmarshall the Payload into the common payload CBOR mapping as defined on section 2.6.3 in
// https://ec.europa.eu/health/sites/default/files/ehealth/docs/digital-green-certificates_v3_en.pdf
// also see CWT for CBOR mapping of iss, exp, iat
// https://datatracker.ietf.org/doc/html/rfc8392#section-4
type DGCPayloadCBORMapping struct {
	ISS   string   `cbor:"1,keyasint,omitempty"`
	EXP   uint64   `cbor:"4,keyasint,omitempty"`
	IAT   uint64   `cbor:"6,keyasint,omitempty"`
	HCERT HCERTMap `cbor:"-260,keyasint,omitempty"`
}
