package server

import (
	"Envoy-Pilot/cmd/server/model"
	"Envoy-Pilot/cmd/server/service"
	"Envoy-Pilot/cmd/server/storage"
	"Envoy-Pilot/cmd/server/util"
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

const envoySubscriberKey = "envoySubscriber"

var defaultPushService *service.DefaultPushService
var xdsConfigDao *storage.XdsConfigDao
var v2Helper *service.V2HelperService

func init() {
	defaultPushService = service.GetDefaultPushService()
	xdsConfigDao = storage.GetXdsConfigDao()
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
			defaultPushService.DeleteSubscriber(subscriber)
			return err
		}
		if i == 0 {
			if !IsValidSubscriber(req) {
				log.Printf("[%s] Error: Invalid cluster or node id %+v\n", xdsType, req)
				cancel()
				return errors.New("Invalid cluster or node id")
			}
			subscriber = &model.EnvoySubscriber{
				Cluster:            req.Node.Cluster,
				Node:               req.Node.Id,
				SubscribedTo:       xdsType,
				LastUpdatedVersion: util.TrimVersion(req.VersionInfo),
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

func IsValidSubscriber(req *v2.DiscoveryRequest) bool {
	fmt.Printf("Node: %s -- len: %d\n", req.Node.Id, len(req.Node.Id))
	fmt.Printf("Cluster: %s -- len: %d\n", req.Node.Cluster, len(req.Node.Cluster))
	return (len(req.Node.Cluster) > 0) && (len(req.Node.Id) > 0)
}
