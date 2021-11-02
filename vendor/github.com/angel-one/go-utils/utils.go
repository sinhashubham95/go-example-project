package utils

import (
	"encoding/xml"
	"github.com/angel-one/go-utils/log"
	jsoniter "github.com/json-iterator/go"
	"io"
	"io/ioutil"
)

// GetDataAsBytes is used to get the data as bytes.
func GetDataAsBytes(data io.ReadCloser) ([]byte, error) {
	bytes, err := ioutil.ReadAll(data)
	defer closeData(data)
	return bytes, err
}

// GetDataAsString is used to get the data as string.
func GetDataAsString(data io.ReadCloser) (string, error) {
	bytes, err := GetDataAsBytes(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// GetJSONData is used to get the JSON data parsed into a struct.
// Make sure you pass the struct by reference.
func GetJSONData(data io.ReadCloser, val interface{}) error {
	bytes, err := GetDataAsBytes(data)
	if err != nil {
		return err
	}
	return jsoniter.Unmarshal(bytes, val)
}

// GetXMLData is used to get the XML data parsed into a struct.
// Make sure you pass the struct by reference.
func GetXMLData(data io.ReadCloser, val interface{}) error {
	bytes, err := GetDataAsBytes(data)
	if err != nil {
		return err
	}
	return xml.Unmarshal(bytes, val)
}

func closeData(data io.ReadCloser) {
	err := data.Close()
	if err != nil {
		log.Error(nil).Err(err).Msg("error closing data")
	}
}
