package constant

const (
	SUBSCRIBE_CDS = "CDS"
	SUBSCRIBE_LDS = "LDS"
	SUBSCRIBE_RDS = "RDS"
	SUBSCRIBE_ADS = "ADS"
	SUBSCRIBE_EDS = "EDS"
)

const ENVOY_SUBSCRIBER_KEY = "envoySubscriber"

// Not really constants
var SUPPORTED_TYPES = []string{SUBSCRIBE_CDS, SUBSCRIBE_LDS, SUBSCRIBE_RDS, SUBSCRIBE_EDS}
var ENV_PATH = "/.env"
var CONSUL_PREFIX = "xDS"
var FILE_MODE = false
var FOLDER_PATH = ""
