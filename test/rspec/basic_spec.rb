require 'rest-client'
require 'json'
require 'plissken'
require 'diplomat'

Diplomat.configure do |config|
    config.url = 'http://localhost:8500'
end

listener0_json = %Q{
    {
        "name": "listener_0",
        "address": {
            "socket_address": {
                "address": "0.0.0.0",
                "port_value": 80
            }
        },
        "filter_chains": [
            {
                "filters": [
                    {
                        "name": "envoy.http_connection_manager",
                        "config": {
                            "stat_prefix": "ingress_http",
                            "access_log": [],
                            "codec_type": "HTTP2",
                            "route_config": {
                                "name": "local_http_router",
                                "virtual_hosts": [
                                    {
                                        "name": "local_service",
                                        "domains": [
                                            "*"
                                        ],
                                        "routes": [
                                            {
                                                "match": {
                                                    "prefix": "/"
                                                },
                                                "route": {
                                                    "cluster": "app1"
                                                }
                                            }
                                        ]
                                    }
                                ]
                            },
                            "http_filters": [
                                {
                                    "name": "envoy.health_check",
                                    "config": {
                                        "pass_through_mode": false,
                                        "endpoint": "/healthz"
                                    }
                                },
                                {
                                    "name": "envoy.router"
                                }
                            ]
                        }
                    }
                ]
            }
        ]
    }
}

listener1_json = %Q{
    {
        "name": "listener_1",
        "address": {
           "socket_address": {
              "address": "127.0.0.1",
              "port_value": 10001
           }
        },
        "filter_chains": [
           {
              "filters": [
                 {
                    "name": "envoy.http_connection_manager",
                    "config": {
                       "stat_prefix": "ingress_http",
                       "access_log": [
                          {
                             "name": "envoy.file_access_log",
                             "config": {
                                "path": "/dev/stdout",
                                "format": "some-format"
                             }
                          }
                       ],
                       "codec_type": "HTTP2",
                       "rds": {
                          "route_config_name": "listener_1_route",
                          "config_source": {
                             "api_config_source": {
                                "api_type": "GRPC",
                                "grpc_services": [{
                                   "envoy_grpc": {
                                      "cluster_name": "xds_cluster"
                                   }
                                }]
                             }
                          }
                       },
                       "http_filters": [
                          {
                             "name": "envoy.router"
                          }
                       ]
                    }
                 }
              ]
           }
        ]
     }
}

listeners_json = %Q{
    [
    #{listener0_json},
    #{listener1_json}
    ]
}

cluster0_json = %Q{
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
}

cluster_version = "1.0"
listener_version = "1.0"

describe "xDS" do
    before(:all) do
        CLUSTER_KEY = "cluster/cdstest-cluster/node/cdstest-node/cluster"
        LISTENER_KEY = "cluster/cdstest-cluster/node/cdstest-node/listener"
        cdelete(CLUSTER_KEY)
        cdelete(LISTENER_KEY)

        cset("#{CLUSTER_KEY}/config", cluster0_json)
        cset("#{CLUSTER_KEY}/version", cluster_version)

        cset("#{LISTENER_KEY}/config", listeners_json)
        cset("#{LISTENER_KEY}/version", listener_version)
        sleep 15
    end

    it "Add a cluster" do
        resp = RestClient.get 'http://localhost:9901/config_dump'
        json = JSON.parse(resp)
        actual = json["configs"]["clusters"]["dynamicActiveClusters"][0]["cluster"]
        actual = actual.to_snake_keys
        actualVersion = json["configs"]["clusters"]["dynamicActiveClusters"][0]["versionInfo"]
        
        expected = JSON.parse(cluster0_json)
        
        expect(actual).to eq(expected)
        expect(actualVersion).to eq(cluster_version)
    end

    it 'Add a listener without rds' do
        resp = RestClient.get 'http://localhost:9901/config_dump'
        json = JSON.parse(resp)
        actual = json["configs"]["listeners"]["dynamicActiveListeners"][0]["listener"]
        actual = actual.to_snake_keys
        actualVersion = json["configs"]["listeners"]["dynamicActiveListeners"][0]["versionInfo"]

        expected = JSON.parse(listeners_json)
        
        expect(actual).to eq(expected[0])
        expect(actualVersion).to eq(listener_version)
    end

    it 'Add a listener with rds' do
        resp = RestClient.get 'http://localhost:9901/config_dump'
        json = JSON.parse(resp)
        actual = json["configs"]["listeners"]["dynamicActiveListeners"][1]["listener"]
        actual = actual.to_snake_keys
        actualVersion = json["configs"]["listeners"]["dynamicActiveListeners"][1]["versionInfo"]

        expected = JSON.parse(listeners_json)
        
        expect(actual).to eq(expected[1])
        expect(actualVersion).to eq(listener_version)
    end
  end

  def cset key, val
    Diplomat::Kv.put(key, val)
  end

  def cdelete key
    Diplomat::Kv.delete(key)
  end