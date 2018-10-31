require 'rest-client'
require 'json'
require 'plissken'
require 'diplomat'

Diplomat.configure do |config|
    config.url = 'http://localhost:8500'
end

listener0_json = File.read 'json/listener_0.json'
listener1_json = File.read 'json/listener_1.json'
listener2_json = File.read 'json/listener_2.json'
listener3_json = File.read 'json/listener_3.json'

listeners_json = %Q{
    [
    #{listener0_json},
    #{listener1_json},
    #{listener2_json},
    #{listener3_json}
    ]
}

cluster0_json = File.read 'json/cluster_0.json'
cluster1_json = File.read 'json/cluster_1.json'
cluster2_json = File.read 'json/cluster_2.json'
cluster3_json = File.read 'json/cluster_3.json'

clusters_json = %Q{
    [
        #{cluster0_json},
        #{cluster1_json},
        #{cluster2_json},
        #{cluster3_json}
    ]
}

route0_json = %Q{
    {
        "name": "listener_1_route",
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
    }
}

routes_json = %Q{
    [
        #{route0_json}
    ]
}

endpoint0_json = %Q{
    {
        "cluster_name": "some_service",
        "endpoints": [
            {
                "lb_endpoints": [
                    {
                        "endpoint": {
                            "address": {
                                "socket_address": {
                                    "address": "app-server",
                                    "port_value": 8123
                                }
                            }
                        }
                    }
                ]
            }
        ]
    }
}

endpoints_json = %Q{
    [
        #{endpoint0_json}
    ]
}

cluster_version = "1.0"
listener_version = "1.0"
route_version = "1.0"
endpoint_version = "1.0"

def getDynamicCluster port, idx
    resp = RestClient.get "http://localhost:#{port}/config_dump"
    json = JSON.parse(resp)
    actual = json["configs"]["clusters"]["dynamicActiveClusters"][idx]["cluster"]
    actual = actual.to_snake_keys
    return actual
end

def getDynamicListener port, idx
    resp = RestClient.get "http://localhost:#{port}/config_dump"
    json = JSON.parse(resp)
    actual = json["configs"]["listeners"]["dynamicActiveListeners"][idx]["listener"]
    actual = actual.to_snake_keys
    return actual
end

def getDynamicRoute port, idx
    resp = RestClient.get "http://localhost:#{port}/config_dump"
    json = JSON.parse(resp)
    actual = json["configs"]["routes"]["dynamicRouteConfigs"][idx]["routeConfig"]
    actual = actual.to_snake_keys
    return actual
end

def getVersion port, key1, key2, idx
    resp = RestClient.get "http://localhost:#{port}/config_dump"
    json = JSON.parse(resp)
    actualVersion = json["configs"][key1][key2][idx]["versionInfo"] 
end

describe "xDS" do
    let(:port) { 9901 }
    before(:all) do
        ['xdstest-cluster', 'adstest-cluster'].each { |type|
            clusterKey = "xDS/app-cluster/#{type}/CDS"
            listenerKey = "xDS/app-cluster/#{type}/LDS"
            routeKey = "xDS/app-cluster/#{type}/RDS"
            endpointKey = "xDS/app-cluster/#{type}/EDS"
    
            cdelete(clusterKey)
            cdelete(listenerKey)
            cdelete(routeKey)
    
            cset("#{clusterKey}/config", clusters_json)
            cset("#{clusterKey}/version", cluster_version)
    
            cset("#{listenerKey}/config", listeners_json)
            cset("#{listenerKey}/version", listener_version)
    
            cset("#{routeKey}/config", routes_json)
            cset("#{routeKey}/version", route_version)
    
            cset("#{endpointKey}/config", endpoints_json)
            cset("#{endpointKey}/version", endpoint_version)    
        }
        sleep 60
    end

    describe "CDS" do
        it "Add a cluster" do
            actual = getDynamicCluster(port, 0)
            actualVersion = getVersion(port, "clusters", "dynamicActiveClusters", 0)
            
            expected = JSON.parse(clusters_json)
            expected[0]["type"] = expected[0]["type"].upcase

            expect(actual).to eq(expected[0])
            expect(actualVersion).to eq(cluster_version)
        end

        it "Add a cluster with http2 options" do
            actual = getDynamicCluster(port, 1)
            actualVersion = getVersion(port, "clusters", "dynamicActiveClusters", 1)
            
            expected = JSON.parse(clusters_json)
            expected[1]["type"] = expected[1]["type"].upcase

            expect(actual).to eq(expected[1])
            expect(actualVersion).to eq(cluster_version)
        end

        it "Add a cluster without http2" do
            actual = getDynamicCluster(port, 2)
            actualVersion = getVersion(port, "clusters", "dynamicActiveClusters", 2)
            
            expected = JSON.parse(clusters_json)
            expected[2]["type"] = expected[2]["type"].upcase

            expected[2].delete("lb_policy")

            expect(actual).to eq(expected[2])
            expect(actualVersion).to eq(cluster_version) 
        end
    end

    describe "LDS" do
        it 'Add a listener without rds' do
            actual = getDynamicListener(port, 0)
            actualVersion = getVersion(port, "listeners", "dynamicActiveListeners", 0)

            expected = JSON.parse(listeners_json)
            
            expect(actual).to eq(expected[0])
            expect(actualVersion).to eq(listener_version)
        end

        it 'Add a listener with rds' do
            actual = getDynamicListener(port, 1)
            actualVersion = getVersion(port, "listeners", "dynamicActiveListeners", 1)

            expected = JSON.parse(listeners_json)
            
            expect(actual).to eq(expected[1])
            expect(actualVersion).to eq(listener_version)
        end

        it 'Add a listener with tls context' do
            actual = getDynamicListener(port, 3)
            actualVersion = getVersion(port, "listeners", "dynamicActiveListeners", 3)

            expected = JSON.parse(listeners_json)
            
            expect(actual).to eq(expected[3])
            expect(actualVersion).to eq(listener_version)
        end
    end

    describe "RDS" do
        it 'Add a dynamic route' do
            actual = getDynamicRoute(port, 0)
            actualVersion = getVersion(port, "routes", "dynamicRouteConfigs", 0)

            expected = JSON.parse(routes_json)
            
            expect(actual).to eq(expected[0])
            expect(actualVersion).to eq(route_version)
        end
    end

    describe "EDS" do
        it 'Add a cluster with eds' do
            actual = getDynamicCluster(port, 3)
            actualVersion = getVersion(port, "clusters", "dynamicActiveClusters", 3)
            
            expected = JSON.parse(clusters_json)
            expected[2]["type"] = expected[3]["type"].upcase

            expect(actual).to eq(expected[3])
            expect(actualVersion).to eq(cluster_version) 

            appResponse = RestClient.get "http://localhost:18123/abc"
            expect(appResponse.code).to eq(200)
            expect(appResponse.body).to eq("Responding to abc!")
        end
    end

    describe "Aggregated Discovery Services(ADS)" do
        let(:port) { 9902 }

        describe "CDS" do
            it "Add a cluster" do
                actual = getDynamicCluster(port, 0)
                actualVersion = getVersion(port, "clusters", "dynamicActiveClusters", 0)
                
                expected = JSON.parse(clusters_json)
                expected[0]["type"] = expected[0]["type"].upcase
    
                expect(actual).to eq(expected[0])
                expect(actualVersion).to eq(cluster_version)
            end
    
            it "Add a cluster with http2 options" do
                actual = getDynamicCluster(port, 1)
                actualVersion = getVersion(port, "clusters", "dynamicActiveClusters", 1)
                
                expected = JSON.parse(clusters_json)
                expected[1]["type"] = expected[1]["type"].upcase
    
                expect(actual).to eq(expected[1])
                expect(actualVersion).to eq(cluster_version)
            end
        end
    
        describe "LDS" do
            it 'Add a listener without rds' do
                actual = getDynamicListener(port, 0)
                actualVersion = getVersion(port, "listeners", "dynamicActiveListeners", 0)
    
                expected = JSON.parse(listeners_json)
                
                expect(actual).to eq(expected[0])
                expect(actualVersion).to eq(listener_version)
            end
    
            it 'Add a listener with rds' do
                actual = getDynamicListener(port, 1)
                actualVersion = getVersion(port, "listeners", "dynamicActiveListeners", 1)
    
                expected = JSON.parse(listeners_json)
                
                expect(actual).to eq(expected[1])
                expect(actualVersion).to eq(listener_version)
            end
        end
    
        describe "RDS" do
            it 'Add a dynamic route' do
                actual = getDynamicRoute(port, 0)
    
                expected = JSON.parse(routes_json)
                
                expect(actual).to eq(expected[0])
            end
        end

        describe "EDS" do
            it 'Add a cluster with eds' do
                actual = getDynamicCluster(port, 3)
                actualVersion = getVersion(port, "clusters", "dynamicActiveClusters", 3)
                
                expected = JSON.parse(clusters_json)
                expected[2]["type"] = expected[3]["type"].upcase
    
                expect(actual).to eq(expected[3])
                expect(actualVersion).to eq(cluster_version) 
    
                appResponse = RestClient.get "http://localhost:28123/abc"
                expect(appResponse.code).to eq(200)
                expect(appResponse.body).to eq("Responding to abc!")
            end
        end
    end
  end

  def cset key, val
    Diplomat::Kv.put(key, val)
  end

  def cdelete key
    Diplomat::Kv.delete(key)
  end