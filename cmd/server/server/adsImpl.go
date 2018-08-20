package server

import (
	"Envoy-xDS/cmd/server/constant"
	"Envoy-xDS/cmd/server/model"
	"context"
	"log"

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
)

//IncrementalAggregatedResources - Not implemented
func (s *Server) IncrementalAggregatedResources(_ discovery.AggregatedDiscoveryService_IncrementalAggregatedResourcesServer) error {
	panic("Not implemented")
}

// StreamAggregatedResources - ADS server impl
func (s *Server) StreamAggregatedResources(stream discovery.AggregatedDiscoveryService_StreamAggregatedResourcesServer) error {
	log.Printf("[%s] -------------- Starting a %s stream ------------------\n", constant.SUBSCRIBE_ADS, constant.SUBSCRIBE_ADS)

	serverCtx, cancel := context.WithCancel(context.Background())
	dispatchChannel := make(chan string)
	i := 0
	var subscriber *model.EnvoySubscriber

	for {
		req, err := stream.Recv()
		if err != nil {
			log.Printf("[%s] Disconnecting client %s\n", constant.SUBSCRIBE_ADS, subscriber.BuildInstanceKey())
			log.Println(err)
			cancel()
			return err
		}
		if i == 0 {
			subscriber = &model.EnvoySubscriber{
				Cluster:            req.Node.Cluster,
				Node:               req.Node.Id,
				SubscribedTo:       constant.SUBSCRIBE_ADS,
				LastUpdatedVersion: getReqVersion(req.VersionInfo),
				AdsList:            make(map[string]*model.EnvoySubscriber),
			}
			serverCtx = context.WithValue(serverCtx, envoySubscriberKey, subscriber)
			defaultPushService.RegisterEnvoy(serverCtx, stream, subscriber, dispatchChannel)
			i++
		}

		topic := v2Helper.GetTopicFor(req.TypeUrl)
		var currentSubscriber *model.EnvoySubscriber

		if subscriber.AdsList[topic] == nil {
			currentSubscriber = &model.EnvoySubscriber{
				Cluster:            req.Node.Cluster,
				Node:               req.Node.Id,
				SubscribedTo:       topic,
				LastUpdatedVersion: getReqVersion(req.VersionInfo),
			}
			subscriber.AdsList[topic] = currentSubscriber
			defaultPushService.RegisterEnvoy(serverCtx, stream, subscriber, dispatchChannel)
		} else {
			currentSubscriber = subscriber.AdsList[topic]
		}

		log.Printf("[%s] Received Request from %s\n %+v\n", constant.SUBSCRIBE_ADS, currentSubscriber.BuildInstanceKey(), req)

		if xdsConfigDao.IsACK(currentSubscriber, req.ResponseNonce) {
			defaultPushService.HandleACK(currentSubscriber, req)
			continue
		} else {
			log.Printf("[%s] Response nonce not recognized %s", constant.SUBSCRIBE_ADS, req.ResponseNonce)
		}
	}
}
