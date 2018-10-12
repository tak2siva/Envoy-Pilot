package util

import (
	"encoding/json"
	"log"
	"strings"
)

func Check(err error) {
	if err != nil {
		log.Println("[Util] Error..", err)
	}
}

func CheckAndPanic(err error) {
	if err != nil {
		log.Println("[Util] Error..", err)
		panic(err)
	}
}

func TrimVersion(version string) string {
	if len(version) != 0 {
		return strings.Trim(version, `"'`)
	}
	return ""
}

func CheckNil(obj interface{}) {
	if obj == nil {
		log.Fatal("Object is nil")
	}
}

func ToJson(obj interface{}) []byte {
	res, _ := json.MarshalIndent(&obj, "", "\t")
	return res
}
