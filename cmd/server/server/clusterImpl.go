package server

import (
	"Envoy-xDS/cmd/server/constant"
	"Envoy-xDS/cmd/server/model"
	"context"
	"errors"
	"log"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func (s *Server) FetchClusters(ctx context.Context, in *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	log.Printf("%+v\n", in)
	panic("Not implemented")
	return &v2.DiscoveryResponse{VersionInfo: "2"}, nil
}

func (s *Server) IncrementalClusters(_ v2.ClusterDiscoveryService_IncrementalClustersServer) error {
	return errors.New("not implemented")
}

// StreamClusters bi directional stream to update cluster config
func (s *Server) StreamClusters(stream v2.ClusterDiscoveryService_StreamClustersServer) error {
	log.Printf("-------------- Starting a cluster stream ------------------\n")

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
}
