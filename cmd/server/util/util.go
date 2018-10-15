package util

import (
	"encoding/json"
	"log"
	"strings"
	"sync"

	// yaml "gopkg.in/yaml.v2"
	yaml "github.com/ghodss/yaml"
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

func ImportJsonOrYaml(jsonStr string) []interface{} {
	var rawArr []interface{}
	jsErr := json.Unmarshal([]byte(jsonStr), &rawArr)
	if jsErr == nil {
		return rawArr
	}

	yamlErr := yaml.Unmarshal([]byte(jsonStr), &rawArr)
	if yamlErr == nil {
		return rawArr
	}

	panic("Invalid json or yaml..")
}

func SyncMapExists(m *sync.Map, k string) bool {
	_, res := m.Load(k)
	return res
}

func SyncMapGetString(m *sync.Map, k string) string {
	res, _ := m.Load(k)
	return res.(string)
}

func SyncMapSet(m *sync.Map, k string, v interface{}) {
	m.Store(k, v)
}

func SyncMapDelete(m *sync.Map, k string) {
	m.Delete(k)
}
