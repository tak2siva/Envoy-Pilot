package constant

const (
	SUBSCRIBE_CDS = "cluster"
	SUBSCRIBE_LDS = "listener"
	SUBSCRIBE_RDS = "route"
	SUBSCRIBE_ADS = "ads"
)

const ENVOY_SUBSCRIBER_KEY = "envoySubscriber"

// Not really constants
var SUPPORTED_TYPES = []string{SUBSCRIBE_CDS, SUBSCRIBE_LDS, SUBSCRIBE_RDS}
var ENV_PATH = "/.env"
