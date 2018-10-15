package storage

import (
	"Envoy-Pilot/cmd/server/model"
	"Envoy-Pilot/cmd/server/util"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/rs/xid"
)

var onceFileConfigDao *FileConfigDao
var onceVersion string

type FileConfigDao struct {
	consulWrapper ConsulWrapper
}

func filePath(sub *model.EnvoySubscriber) string {
	return fmt.Sprintf("my_config/%s/%s.yaml", sub.Cluster, sub.SubscribedTo)
}

func (dao *FileConfigDao) GetLatestVersion(sub *model.EnvoySubscriber) string {
	return util.TrimVersion(GetFileConfigVersion())
}

func (dao *FileConfigDao) GetLatestVersionFor(subscriberKey string) string {
	return util.TrimVersion(GetFileConfigVersion())
}

func (dao *FileConfigDao) IsRepoPresent(sub *model.EnvoySubscriber) bool {
	if _, err := os.Stat(filePath(sub)); !os.IsNotExist(err) {
		return true
	}
	return false
}

func (dao *FileConfigDao) GetConfigJson(sub *model.EnvoySubscriber) (string, string) {
	dat, err := ioutil.ReadFile("/tmp/dat")
	util.Check(err)
	return string(dat), dao.GetLatestVersion(sub)
}

func GetFileConfigDao() *FileConfigDao {
	if onceFileConfigDao == nil {
		onceFileConfigDao = &FileConfigDao{consulWrapper: GetConsulWrapper()}
	}
	return onceFileConfigDao
}

func GetFileConfigVersion() string {
	if len(onceVersion) == 0 {
		onceVersion = xid.New().String()
	}
	return onceVersion
}
