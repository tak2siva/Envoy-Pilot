package storage

import (
	"Envoy-xDS/cmd/server/model"
	"encoding/json"
	"log"
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

	log.Println(subscriber.ToJSON())
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
		log.Printf("--------%+v\n", subscriber)
		log.Printf("--------%+v\n", subscriber2)
		t.Errorf("Marshal & Unmarshal arent rqual")
	}
}

func TestXdsConfigDao_GetLatestVersion(t *testing.T) {
	subscriber := model.EnvoySubscriber{
		Id:                   0,
		Cluster:              "vs-cluster",
		Node:                 "vs-node",
		UpdateSuccess:        1,
		UpdateFailures:       0,
		LastUpdatedVersion:   "1.0",
		LastUpdatedTimestamp: time.Now(),
	}

	key := subscriber.BuildRootKey() + "version"
	wrapper.Set(key, "5.2")
	if "5.2" != dao.GetLatestVersion(&subscriber) {
		t.Errorf("Error fetching version..\n")
	}
}

func TestXdsConfigDao_GetClusterACK(t *testing.T) {
	subscriber := model.EnvoySubscriber{
		Id:                   0,
		Cluster:              "vs-cluster",
		Node:                 "vs-node",
		UpdateSuccess:        1,
		UpdateFailures:       0,
		LastUpdatedVersion:   "1.0",
		LastUpdatedTimestamp: time.Now(),
		SubscribedTo:         "cluster",
	}
	nonce := "xvas-1231-sfg13-112312"
	dao.SaveNonce(&subscriber, nonce)
	if !dao.IsACK(&subscriber, nonce) {
		t.Errorf("Nonce is not set properly..\n")
	}
}
