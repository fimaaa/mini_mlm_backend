package util

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Pair struct {
	FilterName string
	Value      interface{}
}

func structToMap(data interface{}) map[string]Pair {
	result := make(map[string]Pair)
	val := reflect.ValueOf(data)
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		tag := typ.Field(i).Tag.Get("json")
		if tag == "" {
			continue
		}
		filter := typ.Field(i).Tag.Get("filter")
		result[tag] = Pair{
			FilterName: filter,
			Value:      field.Interface(),
		}
	}
	return result
}

func GenerateMongoFilter(data interface{}) ([]bson.M, error) {
	obj := structToMap(data)

	keywordFilter := make([]bson.M, 0)

	for key, pair := range obj {
		fiterName := pair.FilterName
		value := pair.Value

		fmt.Printf("Key: %s, Value: %v, FilterName: %s\n", key, pair.Value, pair.FilterName)

		if value == nil {
			continue
		}

		switch fiterName {
		case "skip":
			continue
		case "gte":
			{
				keyword := bson.M{key: bson.M{"$gte": value}}
				keywordFilter = append(keywordFilter, keyword)
			}
		case "lte":
			{
				keyword := bson.M{key: bson.M{"$lte": value}}
				keywordFilter = append(keywordFilter, keyword)
			}
		case "similiar":
			{
				strVal, ok := value.(string)
				if !ok || strVal == "" {
					continue
				}
				keyword := bson.M{key: primitive.Regex{Pattern: strVal, Options: "i"}}
				keywordFilter = append(keywordFilter, keyword)
			}
		default:
			{
				if mVal, ok := value.(primitive.M); ok {
					keywordFilter = append(keywordFilter, mVal)
				}
			}
		}
	}

	if CreatedAtFrom, ok := obj["created_at_from"].Value.(time.Time); ok {
		if !CreatedAtFrom.IsZero() {
			var CreatedAtTo time.Time
			if CreatedAtTo, ok = obj["created_at_to"].Value.(time.Time); !ok {
				CreatedAtTo = time.Now()
			}
			keyword := bson.M{
				"created_at": bson.M{
					"$gte": CreatedAtFrom, // Greater than or equal to date_from
					"$lte": CreatedAtTo,   // Less than or equal to date_to
				},
			}

			// Append the date filter to the keywordFilter slice
			keywordFilter = append(keywordFilter, keyword)
		}
	}

	if UpdatedAtFrom, ok := obj["updated_at_from"].Value.(time.Time); ok {
		if !UpdatedAtFrom.IsZero() {
			var UpdatedAtTo time.Time
			if UpdatedAtTo, ok = obj["updated_at_to"].Value.(time.Time); !ok {
				UpdatedAtTo = time.Now()
			}
			keyword := bson.M{
				"created_at": bson.M{
					"$gte": UpdatedAtFrom, // Greater than or equal to date_from
					"$lte": UpdatedAtTo,   // Less than or equal to date_to
				},
			}

			// Append the date filter to the keywordFilter slice
			keywordFilter = append(keywordFilter, keyword)
		}
	}

	return keywordFilter, nil
}

func StructToBSONM(data interface{}) bson.M {
	result := bson.M{}

	v := reflect.ValueOf(data)
	t := reflect.TypeOf(data)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if field.Kind() == reflect.Ptr && field.IsNil() && strings.Contains(fieldType.Tag.Get("json"), "omitempty") {
			continue // Skip nil pointers
		}

		bsonTag := fieldType.Tag.Get("bson")
		if bsonTag == "" || bsonTag == "-" {
			continue // Skip fields with no bson tag or explicitly ignored fields
		}

		if bsonTag == "omitempty" {
			bsonTag = fieldType.Tag.Get("json")
		}

		// .Contains(str, substr)
		// Handle nil pointers by setting a default value of nil
		// if field.Kind() == reflect.Ptr && field.IsNil() {
		// 	result[bsonTag] = nil
		// 	continue
		// }

		fmt.Println("TAG RESULT ", bsonTag, " ==> ", field.Interface())
		result[bsonTag] = field.Interface()
	}

	return result
}
