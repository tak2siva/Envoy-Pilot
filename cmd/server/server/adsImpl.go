package server

import (
	"Envoy-xDS/cmd/server/constant"
	"Envoy-xDS/cmd/server/model"
	"context"
	"log"
	"time"

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
)

func (s *Server) IncrementalAggregatedResources(_ discovery.AggregatedDiscoveryService_IncrementalAggregatedResourcesServer) error {
	panic("Not implemented")
}

func (s *Server) StreamAggregatedResources(stream discovery.AggregatedDiscoveryService_StreamAggregatedResourcesServer) error {
	log.Printf("-------------- Starting a ADS stream ------------------\n")

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
		log.Printf("Received Request from %s\n %+v\n", "", req)
		continue
		time.Sleep(1000000 * time.Microsecond)
		if i == 0 {
			subscriber = &model.EnvoySubscriber{
				Cluster:            req.Node.Cluster,
				Node:               req.Node.Id,
				SubscribedTo:       constant.SUBSCRIBE_CDS,
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
	return nil
}
