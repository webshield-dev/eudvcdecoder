package datamodel

//ValueSet holds lookup data
//see https://ec.europa.eu/health/sites/default/files/ehealth/docs/digital-green-value-sets_en.pdf
//example https://github.com/ehn-dcc-development/ehn-dcc-schema/blob/release/1.3.0/valuesets/vaccine-medicinal-product.json
type ValueSet struct {

    //ValueSetID its id
    ValueSetID string `json:"valueSetID,omitempty"`

    //ValueSetDate value set date
    ValueSetDate string `json:"valueSetDate,omitempty"`

    //ValueSetValues the values
    ValueSetValues map[string]ValueSetValue  `json:"valueSetValues,omitempty"`

}

//ValueSetValue as it says
type ValueSetValue struct {

    Display string `json:"display,omitempty"`

    Lang string `json:"lang,omitempty"`

    Active bool `json:"active,omitempty"`

    System string `json:"system,omitempty"`

    Version string `json:"version,omitempty"`
}



