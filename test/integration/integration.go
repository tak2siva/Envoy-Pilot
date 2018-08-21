package main

import (
	"Envoy-Pilot/cmd/server/constant"
	"Envoy-Pilot/cmd/server/storage"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/go-test/deep"
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

func main() {
	TestshouldUseCDS()
}

func TestshouldUseCDS() {
	storage.CONSUL_PATH = "localhost:8500"
	wrapper := storage.GetConsulWrapper()

	cluster := "cdstest-cluster"
	node := "cdstest-node"

	version := "1.0"
	jsonString := `
	{
		"name": "app1",
		"connect_timeout": "0.250s",
		"type": "STRICT_DNS",
		"lb_policy": "ROUND_ROBIN",
		"hosts": [
		 {
		  "socket_address": {
		   "address": "127.0.0.2",
		   "port_value": 1234
		  }
	}
	`
	log.Println(wrapper)
	wrapper.Set(fmt.Sprintf("cluster/%s/node/%s/%s/version", cluster, node, constant.SUBSCRIBE_CDS), version)
	wrapper.Set(fmt.Sprintf("cluster/%s/node/%s/%s/config", cluster, node, constant.SUBSCRIBE_CDS), jsonString)

	time.Sleep(12)
	resp, err := http.Get("http://localhost:9901/config_dump")
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	contents, _ := ioutil.ReadAll(resp.Body)

	extra.SetNamingStrategy(extra.LowerCaseWithUnderscores)

	var actualVal map[string]interface{}
	var expectedVal map[string]interface{}

	jsoniter.Unmarshal(contents, &actualVal)
	jsoniter.UnmarshalFromString(jsonString, &expectedVal)

	// actualVal["configs"].(map[string]interface{})
	if diff := deep.Equal(&actualVal, &expectedVal); diff != nil {
		log.Panic(diff)
	}
}
