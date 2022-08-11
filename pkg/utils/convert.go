package utils

import (
	"bytes"
	"golang.org/x/text/encoding/charmap"
	"io/ioutil"
)

func StringFromWindows1251(data string) (string, error) {
	b, err := ioutil.ReadAll(charmap.Windows1251.NewDecoder().Reader(bytes.NewReader([]byte(data))))
	if err != nil {
		return "", err
	}
	return string(b), nil
}
