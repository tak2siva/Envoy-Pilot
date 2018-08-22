package storage

import (
	"Envoy-Pilot/cmd/server/constant"
	"log"
	"os"
	"strconv"
	"sync"

	consul "github.com/hashicorp/consul/api"
	"github.com/joho/godotenv"
)

var singletonConsulWrapper ConsulWrapper
var CONSUL_PATH string
var mux sync.Mutex

const (
	envoySubscriberSequenceKey = "envoySubscriberSequence"
)

type ConsulWrapper struct {
	client *consul.Client
}

func GetConsulWrapper() ConsulWrapper {
	mux.Lock()
	defer mux.Unlock()
	if singletonConsulWrapper.client == nil {
		err := godotenv.Load(constant.ENV_PATH)
		if err != nil {
			log.Print(err)
			log.Fatal("Error loading .env file")
		}
		consulPath := os.Getenv("CONSUL_PATH")
		log.Printf("Consul Path: %s\n", consulPath)
		config := &consul.Config{Address: consulPath}

		singletonConsulWrapper = ConsulWrapper{}
		client, err := consul.NewClient(config)
		if err != nil {
			panic(err)
		}
		singletonConsulWrapper.client = client
	}
	return singletonConsulWrapper
}

// TODO add retry
func (c *ConsulWrapper) GetUniqId() int {
	for i := 0; i < 100; i++ {
		res, id, err := c.checkAndSetUniqId()
		if err != nil {
			log.Println("Error updating uniq CAS")
			panic(err)
		}
		if res {
			return id
		}
		log.Println("Re generating uniq id")
	}

	panic("Unable to generate new id")
}

func (c *ConsulWrapper) checkAndSetUniqId() (bool, int, error) {
	pair, _, err := c.client.KV().Get(envoySubscriberSequenceKey, nil)
	if err != nil {
		panic(err)
	}
	if pair == nil {
		log.Println("nil value...")
		c.Set(envoySubscriberSequenceKey, "1")
		return true, 1, nil
		// pair = c.Get(envoySubscriberSequenceKey)
	}
	id, err := strconv.Atoi(string(pair.Value))
	if err != nil {
		log.Printf("Err getting uniq id: %s\n", pair.Value)
		panic(err)
	}

	log.Printf("Last id value is %d\n", id)
	newId := id + 1
	pair.Value = []byte(strconv.Itoa(newId))
	res, _, err := c.client.KV().CAS(pair, nil)
	return res, newId, err
}

func (c *ConsulWrapper) Set(key string, value string) {
	p := &consul.KVPair{Key: key, Value: []byte(value)}
	_, err := c.client.KV().Put(p, nil)
	if err != nil {
		log.Println(err)
		panic(err)
	}
}

func (c *ConsulWrapper) Get(key string) *consul.KVPair {
	pair, _, err := c.client.KV().Get(key, nil)
	if err != nil {
		panic(err)
	}
	// if pair == nil {
	// log.Printf("Nil value for key %s\n", key)
	// }
	return pair
}

func (c *ConsulWrapper) GetString(key string) string {
	pair := c.Get(key)
	return string(pair.Value)
}

func (c *ConsulWrapper) GetInt(key string) int {
	pair := c.Get(key)
	id, err := strconv.Atoi(string(pair.Value))
	if err != nil {
		log.Println("Err getting uniq id")
		panic(err)
	}
	return id
}

func (c *ConsulWrapper) Delete(key string) error {
	_, err := c.client.KV().Delete(key, nil)
	if err != nil {
		log.Printf("Error deleting key %s\n", key)
		return err
	}
	return nil
}
