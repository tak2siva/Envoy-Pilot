package storage

import (
	"testing"

	consul "github.com/hashicorp/consul/api"
)

var wrapper ConsulWrapper

func init() {
	config := &consul.Config{Address: "localhost:8500"}
	client, err := consul.NewClient(config)
	if err != nil {
		panic(err)
	}
	wrapper = ConsulWrapper{client: client}
}

func TestGetUniqId(t *testing.T) {
	wrapper.Set("envoySubscriberSequence", "21")
	val := wrapper.GetUniqId()
	if val != 22 {
		t.Errorf("Error generating uniq Id")
	}
}
