# Envoy-Pilot

Envoy Pilot or Envoy xDS is a control plane implementation for Envoy written in Golang and uses Consul for persistence.

Currently Supports
   * CDS
   * LDS
   * RDS
   * ADS (for the above)

*Note: Some infrequent configurations might not be mapped. Feel free to PR* 

## Running Docker Compose

From root directory 
```
cd consul && docker-compose up
docker-compose up
```

xDS Server will be exposed on port 7777

Run Envoy Proxy with the following configurations or use `--service-node` && `--service-cluster`
```
node:
  id: TN
  cluster: India
```

Every *DS requires two keys to be set in consul
  * config
  * version

And the key template is `cluster/CLUSTER_NAME/node/NODE_NAME/DISCOVERY_TYPE/(config|version)`

For CDS add KV pairs
  * `cluster/India/node/TN/cluster/version` => "1.0"
  * `cluster/India/node/TN/cluster/config` => `"[{
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

