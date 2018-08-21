package server

import (
	"Envoy-Pilot/cmd/server/model"
	"Envoy-Pilot/cmd/server/service"
	"Envoy-Pilot/cmd/server/storage"
	"context"
	"log"
	"strings"
)

const envoySubscriberKey = "envoySubscriber"

var defaultPushService *service.DefaultPushService
var xdsConfigDao *storage.XdsConfigDao
var v2Helper *service.V2HelperService

func init() {
	defaultPushService = service.GetDefaultPushService()
	xdsConfigDao = storage.GetXdsConfigDao()
}

func getReqVersion(version string) string {
	if len(version) != 0 {
		return strings.Trim(version, `"'`)
	}
	return ""
}

// Server struct will impl CDS, LDS, RDS & ADS
type Server struct{}

// BiDiStreamFor common bi-directional stream impl for cds,lds,rds
func (s *Server) BiDiStreamFor(xdsType string, stream service.XDSStreamServer) error {
	log.Printf("[%s] -------------- Starting a %s stream ------------------\n", xdsType, xdsType)

	serverCtx, cancel := context.WithCancel(context.Background())
	dispatchChannel := make(chan string)
	i := 0
	var subscriber *model.EnvoySubscriber

	for {
		req, err := stream.Recv()
		if err != nil {
			log.Printf("[%s] Disconnecting client %s\n", xdsType, subscriber.BuildInstanceKey())
			log.Println(err)
			cancel()
			return err
		}
		if i == 0 {
			subscriber = &model.EnvoySubscriber{
				Cluster:            req.Node.Cluster,
				Node:               req.Node.Id,
				SubscribedTo:       xdsType,
				LastUpdatedVersion: getReqVersion(req.VersionInfo),
			}
			serverCtx = context.WithValue(serverCtx, envoySubscriberKey, subscriber)
			defaultPushService.RegisterEnvoy(serverCtx, stream, subscriber, dispatchChannel)
			i++
		}

		log.Printf("[%s] Received Request from %s\n %+v\n", xdsType, subscriber.BuildInstanceKey(), req)

		if xdsConfigDao.IsACK(subscriber, req.ResponseNonce) {
			defaultPushService.HandleACK(subscriber, req)
			continue
		} else {
			log.Printf("[%s] Response nonce not recognized %s", xdsType, req.ResponseNonce)
		}
	}
}
