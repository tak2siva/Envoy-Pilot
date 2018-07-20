package storage

import (
	"Envoy-xDS/cmd/server/model"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	consul "github.com/hashicorp/consul/api"
)

var dao XdsConfigDao

func init() {
	config := &consul.Config{Address: "localhost:8500"}
	client, err := consul.NewClient(config)
	if err != nil {
		panic(err)
	}
	wrapper = ConsulWrapper{client: client}
	dao = XdsConfigDao{consulWrapper: wrapper}
}
func TestXdsConfigDao_RegisterSubscriber(t *testing.T) {
	subscriber := model.EnvoySubscriber{
		Id:                   0,
		Cluster:              "some-cluster",
		Node:                 "some-node",
		UpdateSuccess:        1,
		UpdateFailures:       0,
		LastUpdatedVersion:   "1.0",
		LastUpdatedTimestamp: time.Now(),
	}

	fmt.Println(subscriber.ToJSON())
	dao.RegisterSubscriber(&subscriber)
	jsonString := wrapper.GetString(subscriber.BuildInstanceKey() + "/meta")

	var subscriber2 model.EnvoySubscriber
	err := json.Unmarshal([]byte(jsonString), &subscriber2)
	if err != nil {
		t.Errorf("Error unmarshalling \n")
	}
	subscriber.Id = subscriber2.Id
	isEqual := subscriber2.IsEqual(&subscriber)

	if !isEqual {
		fmt.Printf("--------%+v\n", subscriber)
		fmt.Printf("--------%+v\n", subscriber2)
		t.Errorf("Marshal & Unmarshal arent rqual")
	}
}
