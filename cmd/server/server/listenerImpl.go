package server

import (
	"Envoy-xDS/cmd/server/constant"
	"Envoy-xDS/cmd/server/model"
	"context"
	"log"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func (s *Server) FetchListeners(context.Context, *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	panic("Not implemented")
}

func (s *Server) StreamListeners(stream v2.ListenerDiscoveryService_StreamListenersServer) error {
	log.Printf("-------------- Starting a listener stream ------------------\n")

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
				SubscribedTo:       constant.SUBSCRIBE_LDS,
				LastUpdatedVersion: getReqVersion(req.VersionInfo),
			}
			serverCtx = context.WithValue(serverCtx, envoySubscriberKey, subscriber)
			defaultPushService.RegisterEnvoy(serverCtx, stream, subscriber, dispatchChannel)
			i++
		}

		log.Printf("Received Request from %s\n %+v\n", subscriber.BuildInstanceKey(), req)

		if xdsConfigDao.IsACK(subscriber, req.ResponseNonce) {
			defaultPushService.HandleACK(subscriber, req)
			continue
		} else {
			log.Printf("Response nonce not recognized %s", req.ResponseNonce)
		}
	}
}
