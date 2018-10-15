package storage

import (
	"Envoy-Pilot/cmd/server/model"
	"fmt"
)

type XdsConfigDao interface {
	GetLatestVersion(sub *model.EnvoySubscriber) string
	GetLatestVersionFor(subscriberKey string) string
	IsRepoPresent(sub *model.EnvoySubscriber) bool
	GetConfigJson(sub *model.EnvoySubscriber) (string, string)
}

func nonceStreamKey(sub *model.EnvoySubscriber, nonce string) string {
	return fmt.Sprintf("%s/Nonce/Stream/%s", sub.BuildInstanceKey2(), nonce)
}

func GetXdsConfigDao() XdsConfigDao {
	return GetConsulConfigDao()
}
