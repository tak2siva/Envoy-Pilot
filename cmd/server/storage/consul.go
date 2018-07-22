package storage

import (
	"log"
	"strconv"

	consul "github.com/hashicorp/consul/api"
)

var singletonConsulWrapper ConsulWrapper

const (
	envoySubscriberSequenceKey = "envoySubscriberSequence"
)

type ConsulWrapper struct {
	client *consul.Client
}

func GetConsulWrapper() ConsulWrapper {
	if singletonConsulWrapper.client == nil {
		singletonConsulWrapper = ConsulWrapper{}
		client, err := consul.NewClient(&consul.Config{Address: "host.docker.internal:8500"})
		if err != nil {
			panic(err)
		}
		singletonConsulWrapper.client = client
	}
	return singletonConsulWrapper
}

// TODO add retry
func (c *ConsulWrapper) GetUniqId() int {
	pair, _, err := c.client.KV().Get(envoySubscriberSequenceKey, nil)
	if err != nil {
		panic(err)
	}
	if pair == nil {
		log.Println("nil value...")
		c.Set(envoySubscriberSequenceKey, "1")
		pair = c.Get(envoySubscriberSequenceKey)
	}
	id, err := strconv.Atoi(string(pair.Value))
	if err != nil {
		log.Printf("Err getting uniq id: %s\n", pair.Value)
		panic(err)
	}

	// log.Printf("Last id value is %d\n", id)
	newId := id + 1
	pair.Value = []byte(strconv.Itoa(newId))
	res, _, err := c.client.KV().CAS(pair, nil)

	if !res {
		panic("Error Updating uniq id CAS")
	}
	if err != nil {
		log.Println("Error updating uniq CAS")
		panic(err)
	}
	// log.Printf("New uniq id is %d\n", newId)
	return newId
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
	if pair == nil {
		log.Printf("Nil value for key %s\n", key)
	}
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
