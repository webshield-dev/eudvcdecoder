package helper

import (
	"fmt"
	"io/ioutil"
	"os"
)

//ReadData read a file and return bytes
func ReadData(path string) (data []byte, err error) {

	var f *os.File
	f, err = os.Open(os.ExpandEnv(path))
	if err != nil {
		return nil, fmt.Errorf("error reading %s err=%s", os.ExpandEnv(path), err)
	}

	data, err = ioutil.ReadAll(f)
	if err != nil {
		_ = f.Close()
		return nil, err
	}

	return data, f.Close()
}
