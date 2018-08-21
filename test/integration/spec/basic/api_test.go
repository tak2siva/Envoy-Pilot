package basic

import (
	"Envoy-Pilot/cmd/server/constant"
	"fmt"
	"log"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func getStatus(url string) (int, error) {
	resp, err := http.Get(url)
	return resp.StatusCode, err
}

func TestSpec(t *testing.T) {
	Convey("Get admin page", t, func() {
		status, err := getStatus("http://localhost:9901")
		log.Println(err)
		So(status, ShouldEqual, 200)
	})

	Convey("Add one cluster", t, func() {
		cluster := "cdstest-cluster"
		node := "cdstest-node"

		version := "1.0"
		jsonString := `
		{
			"name": "app1",
			"connect_timeout": "0.250s",
			"type": "STRICT_DNS",
			"lb_policy": "ROUND_ROBIN",
			"hosts": [{
			  "socket_address": {
			   "address": "127.0.0.2",
			   "port_value": 1234
			  }
			}]
		}
		`
		consulSet(fmt.Sprintf("cluster/%s/node/%s/%s/version", cluster, node, constant.SUBSCRIBE_CDS), version)
		consulSet(fmt.Sprintf("cluster/%s/node/%s/%s/config", cluster, node, constant.SUBSCRIBE_CDS), jsonString)

		// time.Sleep(12)
		resp, err := http.Get("http://localhost:9901/config_dump")
		handleErr(err)
		defer resp.Body.Close()

		res := compareJsonCluster(resp, jsonString)
		t.Error(res)
		So(res, ShouldBeNil)
		// consulSet("test", "1234")
		// So(consulGet("test"), ShouldEqual, "1234")
	})
}

func TestSpecNew(t *testing.T) {
	// version := "1.0"
	jsonString := `
	{
		"name": "app1",
		"connect_timeout": "0.250s",
		"type": "STRICT_DNS",
		"lb_policy": "ROUND_ROBIN",
		"hosts": [{
		  "socket_address": {
		   "address": "127.0.0.2",
		   "port_value": 1234
		  }
		}]
	}
	`

	resp, err := http.Get("http://localhost:9901/config_dump")
	handleErr(err)
	defer resp.Body.Close()

	res := compareJsonCluster(resp, jsonString)
	t.Error(res)
}
