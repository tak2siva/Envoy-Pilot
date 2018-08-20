package model

import (
	"Envoy-xDS/cmd/server/constant"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type EnvoySubscriber struct {
	Id                 int
	Cluster            string
	Node               string
	UpdateSuccess      int
	UpdateFailures     int
	LastUpdatedVersion string
	// LastUpdatedVersionV2 map[string]string // for ads
	LastUpdatedTimestamp time.Time
	SubscribedTo         string
	// SubscribedToV2       []string // for ads
	AdsList map[string]*EnvoySubscriber
}

func (e *EnvoySubscriber) ToJSON() string {
	json, err := json.Marshal(e)
	if err != nil {
		log.Println("Error converting envoySubscriber to json..")
		panic(err)
	}
	return string(json)
}

func (e *EnvoySubscriber) BuildInstanceKey() string {
	return fmt.Sprintf("cluster/%s/node/%s/%s/%d", e.Cluster, e.Node, e.SubscribedTo, e.Id)
}

func (e *EnvoySubscriber) BuildRootKey() string {
	return fmt.Sprintf("cluster/%s/node/%s/%s/", e.Cluster, e.Node, e.SubscribedTo)
}

func (e *EnvoySubscriber) IsEqual(that *EnvoySubscriber) bool {
	return e.Cluster == that.Cluster && e.Node == that.Node && e.UpdateSuccess == that.UpdateSuccess && e.UpdateFailures == that.UpdateFailures && e.LastUpdatedVersion == that.LastUpdatedVersion
}

func (e *EnvoySubscriber) IsADS() bool {
	return e.SubscribedTo == constant.SUBSCRIBE_ADS
}

func (e *EnvoySubscriber) GetAdsSubscriber(topic string) *EnvoySubscriber {
	return e.AdsList[topic]
}
