package util

import (
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
