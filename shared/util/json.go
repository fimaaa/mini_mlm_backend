package util

import "encoding/json"

func StructToJson(obj interface{}) string {
	bytes, _ := json.Marshal(obj)
	return string(bytes)
}

func JsonToStruct(jsonStr string) (interface{}, error) {
	var data interface{}

	err := json.Unmarshal([]byte(jsonStr), &data)

	return data, err
}

func JsonToObj(jsonOrigin string, objDestination interface{}) error {
	var (
		err error
	)
	err = json.Unmarshal([]byte(jsonOrigin), &objDestination)

	return err
}

// StructToMap converts a struct to a map[string]interface{}
func StructToMap(data interface{}) (map[string]interface{}, error) {
	// Marshal the struct to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON into a map
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, err
	}

	return result, nil
}
