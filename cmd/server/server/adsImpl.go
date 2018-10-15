package server

import (
	"Envoy-Pilot/cmd/server/constant"
	"Envoy-Pilot/cmd/server/metrics"
	"Envoy-Pilot/cmd/server/model"
	"Envoy-Pilot/cmd/server/util"
	"context"
	"errors"
	"log"

	"google.golang.org/grpc/peer"

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
)

//IncrementalAggregatedResources - Not implemented
func (s *Server) IncrementalAggregatedResources(_ discovery.AggregatedDiscoveryService_IncrementalAggregatedResourcesServer) error {
	panic("Not implemented")
}

// StreamAggregatedResources - ADS server impl
func (s *Server) StreamAggregatedResources(stream discovery.AggregatedDiscoveryService_StreamAggregatedResourcesServer) error {
	clientPeer, ok := peer.FromContext(stream.Context())
	clientIP := "unknown"
	if ok {
		clientIP = clientPeer.Addr.String()
	}
	log.Printf("[%s] -------------- Starting a %s stream from %s ------------------\n", constant.SUBSCRIBE_ADS, constant.SUBSCRIBE_ADS, clientIP)

	serverCtx, cancel := context.WithCancel(context.Background())

	dispatchChannel := make(chan model.ConfigMeta)
	i := 0
	var subscriber *model.EnvoySubscriber

	for {
		req, err := stream.Recv()
		if err != nil {
			log.Printf("[%s] Disconnecting client %s\n", constant.SUBSCRIBE_ADS, subscriber.BuildInstanceKey2())
			log.Println(err)
			cancel()
			registerService.DeleteSubscriber(subscriber)
			metrics.DecActiveConnections(subscriber)
			metrics.DecActiveSubscribers(subscriber)
			return err
		}
		if i == 0 {
			if !IsValidSubscriber(req) {
				log.Printf("[%s] Error: Invalid cluster or node id %+v\n", constant.SUBSCRIBE_ADS, req)
				cancel()
				return errors.New("Invalid cluster or node id")
			}
			subscriber = &model.EnvoySubscriber{
				Cluster:            req.Node.Cluster,
				Node:               req.Node.Id,
				SubscribedTo:       constant.SUBSCRIBE_ADS,
				LastUpdatedVersion: util.TrimVersion(req.VersionInfo),
				AdsList:            make(map[string]*model.EnvoySubscriber),
				IpAddress:          clientIP,
			}
			serverCtx = context.WithValue(serverCtx, envoySubscriberKey, subscriber)
			registerService.RegisterEnvoy(serverCtx, stream, subscriber, dispatchChannel)
			metrics.IncActiveConnections(subscriber)
			i++
		}

		topic := v2Helper.GetTopicFor(req.TypeUrl)
		var currentSubscriber *model.EnvoySubscriber

		if subscriber.AdsList[topic] == nil {
			currentSubscriber = &model.EnvoySubscriber{
				Cluster:            req.Node.Cluster,
				Node:               req.Node.Id,
				SubscribedTo:       topic,
				LastUpdatedVersion: util.TrimVersion(req.VersionInfo),
				IpAddress:          clientIP,
			}
			subscriber.AdsList[topic] = currentSubscriber
			registerService.RegisterEnvoyADS(serverCtx, stream, currentSubscriber, dispatchChannel)
			metrics.IncActiveSubscribers(subscriber, currentSubscriber.SubscribedTo)
		} else {
			currentSubscriber = subscriber.AdsList[topic]
		}

		log.Printf("[%s] Received Request from %s\n %s\n", constant.SUBSCRIBE_ADS, currentSubscriber.BuildInstanceKey2(), util.ToJson(req))

		if subscriberDao.IsACK(currentSubscriber, req.ResponseNonce) {
			dispatchService.HandleACK(currentSubscriber, req)
			continue
		} else {
			log.Printf("[%s] Response nonce not recognized %s", constant.SUBSCRIBE_ADS, req.ResponseNonce)
		}
	}
}
