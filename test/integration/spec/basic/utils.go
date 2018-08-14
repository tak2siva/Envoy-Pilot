package basic

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-test/deep"
	consul "github.com/hashicorp/consul/api"
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

const (
	consulPath = "http://localhost:8500"
)

func handleErr(err error) {
	if err != nil {
		log.Println(err)
		panic(err)
	}
}

func consulGet(key string) string {
	config := &consul.Config{Address: consulPath}
	client, err := consul.NewClient(config)
	handleErr(err)
	pair, _, err := client.KV().Get(key, nil)
	handleErr(err)
	if pair == nil || pair.Value == nil {
		return ""
	}
	return string(pair.Value)
}

func consulSet(key string, value string) {
	config := &consul.Config{Address: consulPath}
	client, err := consul.NewClient(config)
	handleErr(err)
	p := &consul.KVPair{Key: key, Value: []byte(value)}
	_, err = client.KV().Put(p, nil)
	handleErr(err)
}

func compareJsonCluster(resp *http.Response, jsonString string) string {
	contents, _ := ioutil.ReadAll(resp.Body)
	extra.SetNamingStrategy(extra.LowerCaseWithUnderscores)
	var actualVal map[string]interface{}
	var expectedVal map[string]interface{}

	jsoniter.Unmarshal(contents, &actualVal)
	jsoniter.UnmarshalFromString(jsonString, &expectedVal)

	configs := actualVal["configs"].(map[string]interface{})
	clusters := configs["clusters"].(map[string]interface{})
	dyClusters := clusters["dynamicActiveClusters"].([]interface{})
	myCluster := dyClusters[0].(map[string]interface{})
	// version := myCluster["versionInfo"].(string)
	myCluster = myCluster["cluster"].(map[string]interface{})

	if diff := deep.Equal(&myCluster, &expectedVal); diff != nil {
		// val, err := jsoniter.Marshal(myCluster)
		// if err != nil {
		// 	log.Panic(err)
		// }
		// return string(val)
		return fmt.Sprintln(diff)
	}
	return ""
}
