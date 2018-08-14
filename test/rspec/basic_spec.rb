require 'rest-client'
require 'json'
require 'plissken'
require 'diplomat'

Diplomat.configure do |config|
    config.url = 'http://localhost:8500'
end

describe "xDS" do
    xit "Add a cluster" do
        CLUSTER_KEY = "cluster/cdstest-cluster/node/cdstest-node/cluster"
        cdelete(CLUSTER_KEY)

        version = "1.0"
        jsonStr = %Q{
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

        cset("#{CLUSTER_KEY}/config", jsonStr)
        cset("#{CLUSTER_KEY}/version", version)

        sleep 12

        resp = RestClient.get 'http://localhost:9901/config_dump'
        json = JSON.parse(resp)
        actual = json["configs"]["clusters"]["dynamicActiveClusters"][0]["cluster"]
        actual = actual.to_snake_keys
        actualVersion = json["configs"]["clusters"]["dynamicActiveClusters"][0]["versionInfo"]
        
        expected = JSON.parse(jsonStr)
        
        expect(actual).to eq(expected)
        expect(actualVersion).to eq(version)
    end

    it 'Add a listener withoud rds' do
        LISTENER_KEY = "cluster/cdstest-cluster/node/cdstest-node/listener"

        version = "1.0"
        jsonStr = %Q{
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

        # cset("#{LISTENER_KEY}/config", jsonStr)
        # cset("#{LISTENER_KEY}/version", version)
        # sleep 12

        resp = RestClient.get 'http://localhost:9901/config_dump'
        json = JSON.parse(resp)
        actual = json["configs"]["listeners"]["dynamicActiveListeners"][0]["listener"]
        actual = actual.to_snake_keys
        actualVersion = json["configs"]["listeners"]["dynamicActiveListeners"][0]["versionInfo"]

        expected = JSON.parse(jsonStr)
        
        expect(actual).to eq(expected)
        expect(actualVersion).to eq(version)
    end
  end

  def cset key, val
    Diplomat::Kv.put(key, val)
  end

  def cdelete key
    Diplomat::Kv.delete(key)
  end