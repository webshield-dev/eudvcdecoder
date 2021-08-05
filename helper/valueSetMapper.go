package helper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/webshield-dev/eudvcdecoder/datamodel"
)

// Routines to get value sets that are loaded from files that are located  at
//  VS_DATA_BASE/valuesetdata/files

const (
	maFileName string = "/vaccine-mah-manf.json"
	mpFileName string = "/vaccine-medicinal-product.json"
	vpFileName string = "/vaccine-prophylaxis.json"
)

//NewValueSetMapper create and initialize all its internal data
func NewValueSetMapper(vsDataPath string) (*ValueSetMapper, error) {

	vsm := &ValueSetMapper{
		vsDataPath: vsDataPath,
	}
	if err := vsm.init(); err != nil {
		return nil, err
	}

	return vsm, nil
}

//ValueSetMapper maps codes to metadata value sets
type ValueSetMapper struct {
	vsDataPath string
	maCodes    *datamodel.ValueSet
	mpCodes    *datamodel.ValueSet
	vpCodes    *datamodel.ValueSet
}

//DecodeMA decode the Marketing authorisation holder or manufacturer, a coded value
//from the value set vaccine-mah-manf.json
func (vsm *ValueSetMapper) DecodeMA(code string) *datamodel.ValueSetValue {
	result := vsm.maCodes.ValueSetValues[code]
	return &result
}

//DecodeMP decode the vaccine product name using A coded value
//from the value set vaccine-medicinal-product.json
func (vsm *ValueSetMapper) DecodeMP(code string) *datamodel.ValueSetValue {
	result := vsm.mpCodes.ValueSetValues[code]
	return &result
}

//DecodeVP decode the Type of the vaccine or prophylaxis used
//from the value set vaccine-prophylaxis.json
func (vsm *ValueSetMapper) DecodeVP(code string) *datamodel.ValueSetValue {
	result := vsm.vpCodes.ValueSetValues[code]
	return &result
}

func (vsm *ValueSetMapper) init() error {

	//setup ma
	data, err := readData(vsm.vsDataPath + maFileName)
	if err != nil {
		return err
	}
	var maCodes datamodel.ValueSet
	if err = json.Unmarshal(data, &maCodes); err != nil {
		return err
	}
	vsm.maCodes = &maCodes

	//setup mp
	data, err = readData(vsm.vsDataPath + mpFileName)
	if err != nil {
		return err
	}
	var mpCodes datamodel.ValueSet
	if err = json.Unmarshal(data, &mpCodes); err != nil {
		return err
	}
	vsm.mpCodes = &mpCodes

	//setup vp
	data, err = readData(vsm.vsDataPath + vpFileName)
	if err != nil {
		return err
	}
	var vpCodes datamodel.ValueSet
	if err = json.Unmarshal(data, &vpCodes); err != nil {
		return err
	}
	vsm.vpCodes = &vpCodes

	return nil

}

func readData(path string) ([]byte, error) {

	f, err := os.Open(os.ExpandEnv(path))
	if err != nil {
		return nil, fmt.Errorf("error reading %s err=%s", path, err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return data, nil
}
