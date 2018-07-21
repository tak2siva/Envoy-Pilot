package storage

import (
	"log"
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
	genID := wrapper.GetUniqId()
	log.Printf("-----%s----\n", wrapper.GetString("envoySubscriberSequence"))
	dbID := wrapper.GetInt("envoySubscriberSequence")
	if genID != 22 || genID != dbID {
		t.Errorf("Error generating uniq Id..\n genId: %d - dbId:%d", genID, dbID)
	}

	wrapper.Delete("envoySubscriberSequence")
	genID = wrapper.GetUniqId()

	dbID = wrapper.GetInt("envoySubscriberSequence")
	if genID != 2 || genID != dbID {
		t.Errorf("Error generating uniq Id from start ..\n genId: %d - dbId:%d", genID, dbID)
	}
}
