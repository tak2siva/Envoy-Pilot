# Envoy-Pilot 

[![Build Status](https://travis-ci.org/tak2siva/Envoy-Pilot.svg?branch=master)](https://travis-ci.org/tak2siva/Envoy-Pilot)    ![Version](https://img.shields.io/badge/version-v0.1.7-yellowgreen.svg)


Envoy Pilot or Envoy xDS is a control plane implementation for [Envoy](https://github.com/envoyproxy/envoy) written in Golang and uses Consul for persistence.

Currently Supports
   * CDS
   * LDS
   * RDS
   * EDS
   * ADS (for the above)

*Note: Some infrequent configurations might not be mapped. Feel free to PR* 

Checkout [Envoy XDS PROTOCOL Overview](https://github.com/envoyproxy/data-plane-api/blob/master/XDS_PROTOCOL.md) for more detail

## Usage

xDS Server will be exposed on port 7777

Run Envoy Proxy with the following configurations or use `--service-node` && `--service-cluster`
```
node:
  id: ride-service-replica-2
  cluster: ride-service
```

Every *DS requires two keys to be set in consul
  * config
  * version

And the key template is `cluster/CLUSTER_NAME/node/NODE_NAME/DISCOVERY_TYPE/(config|version)`

For CDS add KV pairs
  * `xDS/app-cluster/ride-service/CDS/version` => "1.0"
  * `xDS/app-cluster/ride-service/CDS/config` => `"[{
      {
        "name": "app1",
        "connect_timeout": "0.250s",
        "type": "STRICT_DNS",
        "lb_policy": "RANDOM",
        "hosts": [{
          "socket_address": {
           "address": "127.0.0.2",
           "port_value": 1234
          }
        }]
    }
  }]"`

Pushing new configuration
  * Envoy-Pilot will be polling for version change every 10 seconds.  
  * If there is a version mismatch for any of `xDS/app-cluster/ride-service/(CDS|LDS|RDS|EDS)/version` then new config `xDS/app-cluster/ride-service/(CDS|LDS|RDS|EDS)/config` will be pushed to subscriber envoy.
  * If update succeed there will be an ACK log for the instance.

## Running Docker Compose

From root directory 
```
docker network create envoy-pilot_xds-demo
docker-compose -f docker-compose.consul.yaml up
docker-compose -f docker-compose.server.yaml up --build
```


## Runnnig Docker

Consul url need to be set in .env

env_values.txt
```
CONSUL_PATH="http://consul-server:8500"
CONSUL_PREFIX=""xDS"
```

Docker run
```
docker run -v $(pwd)/env_values.txt:/.env -p 7777:7777 -p 9090:9090 tak2siva/envoy-pilot:latest
```

## Helm Chart

Install using the [Helm Chart for Envoy-Pilot](https://github.com/tak2siva/Envoy-Pilot-Helm).

## Debugging

* xDS-Server is running on port 7777
* A http server is running on port 9090 for debugging

`localhost:9090/dump/KEY_TEMPLATE` will give a json dump of proto mapping

  **Ex:** 
  ```
  http://localhost:9090/dump/xDS/app-cluster/ride-service/(CDS|LDS|RDS|EDS)/config
  ```


