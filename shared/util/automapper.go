package util

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

func Automapper(objOrigin interface{}, objDestination interface{}) error {
	var (
		err error
	)
	jsonOrigin := StructToJson(objOrigin)
	err = json.Unmarshal([]byte(jsonOrigin), &objDestination)

	return err
}

func MapToStruct(data map[string]interface{}, result interface{}) error {
	val := reflect.ValueOf(result).Elem()
	typ := val.Type()

	// Iterate over struct fields
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Get JSON tag
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" {
			// If no JSON tag, use field name as default
			jsonTag = fieldType.Name
		}

		// Unconditionally remove ",omitempty" from the tag, if it exists
		jsonTag = strings.Replace(jsonTag, ",omitempty", "", -1)

		// Check if map contains key with JSON tag
		if value, exists := data[jsonTag]; exists {
			if !field.CanSet() {
				fmt.Printf("Cannot set field: %s\n", jsonTag)
				continue
			}

			// Assign value based on field type
			switch field.Kind() {
			case reflect.Ptr:
				if valueStr, ok := value.(string); ok {
					ptr := reflect.New(field.Type().Elem())
					ptr.Elem().SetString(valueStr)
					fmt.Println("TAG valueStr ", valueStr, " - ", value, " => ", field, " - ", ptr)
					field.Set(ptr)
				}

			case reflect.Struct:
				// Handling time.Time struct
				if field.Type() == reflect.TypeOf(time.Time{}) {
					if valueStr, ok := value.(string); ok {
						t, err := time.Parse(time.RFC3339, valueStr)
						if err != nil {
							return err
						}
						field.Set(reflect.ValueOf(t))
					}
				}

			case reflect.String:
				// Handle string fields
				if valueStr, ok := value.(string); ok {
					field.SetString(valueStr)
				} else {
					fmt.Printf("Value is not a string: %v\n", value)
				}

			case reflect.Int, reflect.Int32, reflect.Int64:
				// Handle integer fields
				if valueFloat, ok := value.(float64); ok {
					field.SetInt(int64(valueFloat))
				} else {
					fmt.Printf("Value is not an integer: %v\n", value)
				}

			case reflect.Float32, reflect.Float64:
				// Handle float fields
				if valueFloat, ok := value.(float64); ok {
					field.SetFloat(valueFloat)
				} else {
					fmt.Printf("Value is not a float: %v\n", value)
				}

			case reflect.Bool:
				// Handle boolean fields
				if valueBool, ok := value.(bool); ok {
					field.SetBool(valueBool)
				} else {
					fmt.Printf("Value is not a boolean: %v\n", value)
				}
			case reflect.Slice:
				// Handle slice fields
				sliceValue := reflect.MakeSlice(field.Type(), 0, 0)
				if values, ok := value.([]interface{}); ok {
					for _, v := range values {
						elem := reflect.New(field.Type().Elem()).Elem()
						switch field.Type().Elem().Kind() {
						case reflect.String:
							if str, ok := v.(string); ok {
								elem.SetString(str)
							}
						case reflect.Int, reflect.Int32, reflect.Int64:
							if num, ok := v.(float64); ok {
								elem.SetInt(int64(num))
							}
						case reflect.Float32, reflect.Float64:
							if fnum, ok := v.(float64); ok {
								elem.SetFloat(fnum)
							}
						case reflect.Bool:
							if b, ok := v.(bool); ok {
								elem.SetBool(b)
							}
						default:
							fmt.Printf("Unsupported slice element type: %s\n", field.Type().Elem().Kind())
						}
						sliceValue = reflect.Append(sliceValue, elem)
					}
					field.Set(sliceValue)
				} else {
					fmt.Printf("Value is not a slice: %v\n", value)
				}

			default:
				fmt.Printf("Unsupported field type: %s\n", field.Kind())
			}
		} else {
			fmt.Printf("Key not found for JSON tag: %s\n", jsonTag)
		}
	}
	return nil
}
