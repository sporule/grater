package utility

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
)

//IsNil checks if all items are empty, it will return true if it is nil
func IsNil(items ...interface{}) (result bool) {
	result = false
	for _, item := range items {
		if item == nil {
			result = true
			return result
		}
		value := reflect.ValueOf(item)
		//return true if the value is not valid
		if !value.IsValid() {
			result = true
			return result
		}
		testKind := value.Kind()
		switch value.Kind() {
		case reflect.Slice, reflect.String, reflect.Map, reflect.Array:
			result = value.Len() <= 0
			if testKind == 0 {
				testKind = 12
			}
		default:
			//return false by default
			result = false
		}
		if result {
			//return the result directly if the result is true
			return result
		}
	}
	return result
}

//GetError returns error message if error is not nil
func GetError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

//Result is the result struct for c.json
type Result struct {
	Code int
	Obj  interface{}
}

//Expand expands the result object to fit c.json
func (result *Result) Expand() (int, interface{}) {
	return result.Code, result.Obj
}

//Error is the struct for c.json
type Error struct {
	Error string `json:"error,omitempty"`
}

//Config returns the global config
var Config map[string]string

//LoadConfiguration loads json configuration
func LoadConfiguration(filepath string) (err error) {
	content, err := ioutil.ReadFile(filepath)
	if err == nil {
		err = json.Unmarshal(content, &Config)
		return err
	}
	return err
}
