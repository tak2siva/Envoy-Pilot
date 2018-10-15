package server

import (
	"Envoy-Pilot/cmd/server/metrics"
	"Envoy-Pilot/cmd/server/model"
	"Envoy-Pilot/cmd/server/service"
	"Envoy-Pilot/cmd/server/storage"
	"Envoy-Pilot/cmd/server/util"
	"context"
	"errors"
	"log"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"google.golang.org/grpc/peer"
)

const envoySubscriberKey = "envoySubscriber"

var registerService *service.RegisterService
var dispatchService *service.DispatchService
var xdsConfigDao storage.XdsConfigDao
var subscriberDao *storage.SubscriberDao
var v2Helper *service.V2HelperService

func init() {
	registerService = service.GetRegisterService()
	dispatchService = service.GetDispatchService()
	xdsConfigDao = storage.GetXdsConfigDao()
	subscriberDao = storage.GetSubscriberDao()
}

// Server struct will impl CDS, LDS, RDS & ADS
type Server struct{}

// BiDiStreamFor common bi-directional stream impl for cds,lds,rds
func (s *Server) BiDiStreamFor(xdsType string, stream service.XDSStreamServer) error {
	clientPeer, ok := peer.FromContext(stream.Context())
	clientIP := "unknown"
	if ok {
		clientIP = clientPeer.Addr.String()
	}
	log.Printf("[%s] -------------- Starting a %s stream from %s ------------------\n", xdsType, xdsType, clientIP)

	serverCtx, cancel := context.WithCancel(context.Background())
	dispatchChannel := make(chan model.ConfigMeta)
	i := 0
	var subscriber *model.EnvoySubscriber

	for {
		req, err := stream.Recv()
		if err != nil {
			log.Printf("[%s] Disconnecting client %s\n", xdsType, subscriber.BuildInstanceKey2())
			log.Println(err)
			cancel()
			registerService.DeleteSubscriber(subscriber)
			metrics.DecActiveConnections(subscriber)
			metrics.DecActiveSubscribers(subscriber)
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
				IpAddress:          clientIP,
			}
			serverCtx = context.WithValue(serverCtx, envoySubscriberKey, subscriber)
			registerService.RegisterEnvoy(serverCtx, stream, subscriber, dispatchChannel)
			metrics.IncActiveConnections(subscriber)
			metrics.IncActiveSubscribers(subscriber, subscriber.SubscribedTo)
			i++
		}

		log.Printf("[%s] Received Request from %s\n %s\n", xdsType, subscriber.BuildInstanceKey2(), util.ToJson(req))

		if subscriberDao.IsACK(subscriber, req.ResponseNonce) {
			dispatchService.HandleACK(subscriber, req)
			continue
		} else {
			log.Printf("[%s] Response nonce not recognized %s", xdsType, req.ResponseNonce)
		}
	}
}

func IsValidSubscriber(req *v2.DiscoveryRequest) bool {
	return (len(req.Node.Cluster) > 0) && (len(req.Node.Id) > 0)
}
