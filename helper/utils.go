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
        return nil, fmt.Errorf("error reading %s err=%s", path, err)
    }

    defer func() {
        err1 := f.Close()
        if err == nil {
            err = err1
        }
    }()

    data, err = ioutil.ReadAll(f)
    if err != nil {
        return nil, err
    }

    return data, err
}
