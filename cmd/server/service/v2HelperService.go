package service

import (
	"Envoy-xDS/cmd/server/constant"
	"fmt"

	"github.com/envoyproxy/go-control-plane/pkg/cache"
)

type V2HelperService struct{}

func (v *V2HelperService) GetTypeUrlFor(topic string) string {
	switch topic {
	case constant.SUBSCRIBE_CDS:
		return cache.ClusterType
	case constant.SUBSCRIBE_LDS:
		return cache.ListenerType
	case constant.SUBSCRIBE_RDS:
		return cache.RouteType
	default:
		panic(fmt.Sprintf("No TypeUrl found for type %s\n", topic))
	}
}

func (v *V2HelperService) GetTopicFor(typeUrl string) string {
	switch typeUrl {
	case cache.ClusterType:
		return constant.SUBSCRIBE_CDS
	case cache.ListenerType:
		return constant.SUBSCRIBE_LDS
	case cache.RouteType:
		return constant.SUBSCRIBE_RDS
	default:
		panic(fmt.Sprintf("No Topic found for typeUrl %s\n", typeUrl))
	}
}
