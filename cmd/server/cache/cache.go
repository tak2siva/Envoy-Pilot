package cache

import "Envoy-Pilot/cmd/server/model"

var SUBSCRIBER_CACHE = make(map[string]*model.EnvoySubscriber)
