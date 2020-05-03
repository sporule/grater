package utility

import "reflect"

//IsNil checks if all items are empty, it will return true if it is not nil
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
		if !result {
			//return the result directly if the result is true
			return result
		}
	}
	return result
}
