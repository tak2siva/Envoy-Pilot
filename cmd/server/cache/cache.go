package cache

import "sync"

var SUBSCRIBER_CACHE = sync.Map{}
var NONCE_CACHE = sync.Map{}
