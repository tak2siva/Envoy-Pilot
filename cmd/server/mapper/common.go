package mapper

import (
	"errors"
	"log"
	"strconv"

	google_protobuf1 "github.com/gogo/protobuf/types"
)

func getInt(obj map[string]interface{}, key string) (int, error) {
	switch t := obj[key].(type) {
	case string:
		val, err := strconv.Atoi(obj[key].(string))
		if err != nil {
			log.Printf("Error parsing int %s\n", obj[key].(string))
			return 0, err
		}
		return val, err
	case float64:
		val := int(obj[key].(float64))
		return val, nil
	default:
		log.Printf("Unknown field type for value %+v and type %s", obj[key], t)
		return 0, errors.New("Unable to parse")
	}
}

func getUInt(obj map[string]interface{}, key string) (uint32, error) {
	val, err := getInt(obj, key)
	if err != nil {
		return 0, err
	}
	return uint32(val), err
}

func getUIntValue(obj map[string]interface{}, key string) (google_protobuf1.UInt32Value, error) {
	val, err := getInt(obj, key)
	if err != nil {
		return google_protobuf1.UInt32Value{}, err
	}
	uVal := uint32(val)
	return google_protobuf1.UInt32Value{Value: uVal}, err
}

func getString(obj map[string]interface{}, key string) string {
	return obj[key].(string)
}

func getBoolean(obj map[string]interface{}, key string) bool {
	return obj[key].(bool)
}

func getBoolValue(obj map[string]interface{}, key string) google_protobuf1.BoolValue {
	return google_protobuf1.BoolValue{
		Value: getBoolean(obj, key),
	}
}

func getStringArray(obj map[string]interface{}, key string) []string {
	arr := obj[key].([]interface{})
	res := make([]string, len(arr))
	for i, value := range arr {
		res[i] = value.(string)
	}
	return res
}

func getFloat(obj map[string]interface{}, key string) float64 {
	return obj[key].(float64)
}

func toMap(obj interface{}) map[string]interface{} {
	return obj.(map[string]interface{})
}

func toArray(obj interface{}) []interface{} {
	return obj.([]interface{})
}

func keyExists(objMap map[string]interface{}, key string) bool {
	return objMap[key] != nil
}
