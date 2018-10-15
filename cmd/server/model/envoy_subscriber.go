package model

import (
	"Envoy-Pilot/cmd/server/constant"
	"Envoy-Pilot/cmd/server/util"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type EnvoySubscriber struct {
	Guid               string
	Cluster            string
	Node               string
	UpdateSuccess      int
	UpdateFailures     int
	LastUpdatedVersion string
	// LastUpdatedVersionV2 map[string]string // for ads
	LastUpdatedTimestamp time.Time
	SubscribedTo         string
	// SubscribedToV2       []string // for ads
	AdsList   map[string]*EnvoySubscriber
	IpAddress string
}

func (e *EnvoySubscriber) ToJSON() string {
	json, err := json.Marshal(e)
	if err != nil {
		log.Println("Error converting envoySubscriber to json..")
		panic(err)
	}
	return string(json)
}

func (e *EnvoySubscriber) BuildInstanceKey2() string {
	// return fmt.Sprintf("cluster/%s/node/%s/%s/%d", e.Cluster, e.Node, e.SubscribedTo, e.Id)
	return fmt.Sprintf("%s/app-cluster/%s/%s/%s", constant.CONSUL_PREFIX, e.Cluster, e.SubscribedTo, e.Guid)
}

func (e *EnvoySubscriber) BuildRootKey() string {
	// return fmt.Sprintf("cluster/%s/node/%s/%s/", e.Cluster, e.Node, e.SubscribedTo)
	return fmt.Sprintf("%s/app-cluster/%s/%s/", constant.CONSUL_PREFIX, e.Cluster, e.SubscribedTo)
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

func (e *EnvoySubscriber) IsOutdated(newVersion string) bool {
	latest := util.TrimVersion(newVersion)
	actual := util.TrimVersion(e.LastUpdatedVersion)
	return latest != actual
}
