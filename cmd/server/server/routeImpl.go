package server

import (
	"Envoy-xDS/cmd/server/constant"
	"Envoy-xDS/cmd/server/model"
	"context"
	"log"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func (s *Server) FetchRoutes(context.Context, *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	panic("Not implemented")
}

func (s *Server) IncrementalRoutes(v2.RouteDiscoveryService_IncrementalRoutesServer) error {
	panic("Not implemented")
}

func (s *Server) StreamRoutes(stream v2.RouteDiscoveryService_StreamRoutesServer) error {
	log.Printf("-------------- Starting a route stream ------------------\n")
	serverCtx, cancel := context.WithCancel(context.Background())
	dispatchChannel := make(chan bool)
	i := 0
	var subscriber *model.EnvoySubscriber

	for {
		req, err := stream.Recv()
		if err != nil {
			log.Printf("Disconnecting client %s\n", subscriber.BuildInstanceKey())
			log.Println(err)
			cancel()
			return err
		}
		if i == 0 {
			subscriber = &model.EnvoySubscriber{
				Cluster:            req.Node.Cluster,
				Node:               req.Node.Id,
				SubscribedTo:       constant.SUBSCRIBE_RDS,
				LastUpdatedVersion: getReqVersion(req.VersionInfo),
			}
			serverCtx = context.WithValue(serverCtx, envoySubscriberKey, subscriber)
			clusterService.RegisterEnvoy(serverCtx, stream, subscriber, dispatchChannel)
			i++
		}

		log.Printf("Received Request from %s\n %+v\n", subscriber.BuildInstanceKey(), req)

		if xdsConfigDao.IsACK(subscriber, req.ResponseNonce) {
			clusterService.HandleACK(subscriber, req)
			continue
		} else {
			log.Printf("Response nonce not recognized %s", req.ResponseNonce)
		}
	}
}
