// Package lib contains: cleaner.go - consul.go - destroy.go - runner.go - snapshot.go - sync.go - take.go - tracker.go - uuid.go
//
// Consul get the status of consul member
//
package lib

import (
	"fmt"
	"os"

	"github.com/hashicorp/consul/api"
)

// Alive returns if the Consul server is available
func Alive() bool {
	var alive bool

	// Create a new client
	client, _ := Client()

	// Get the operator endpoints
	agent := client.Agent()

	// Get members list
	members, err := agent.Members(false)
	if err != nil {
		w.Err("[ERROR > lib/consul.go:25]@[CONSUL] it was not possible to get the members list.")
	} else {
		if len(members) > 0 {
			alive = true
		}
	}
	return alive
}

// Client returns a new client of Consul
func Client() (*api.Client, *api.KV) {
	// Create a new client
	var kv *api.KV
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		w.Err("[ERROR > lib/consul.go:40]@[CONSUL] it was not possible to create a new client.")
	} else {
		kv = client.KV()
	}
	return client, kv
}

// Host returns the hostname of the node
func Host() string {
	// Resolve hostname
	hostname, err := os.Hostname()
	if err != nil {
		w.Err("[ERROR > lib/consul.go:52] it was not possible to resolve the hostname.")
	}
	return hostname
}

// DeleteKV removes a KV pair
func DeleteKV(key string, value string, datacenter string) int {
	// Options to Delete() function
	q := &api.WriteOptions{Datacenter: datacenter}
	_, kv := Client()

	// Format of KV pair
	pair := fmt.Sprintf("%s:%s", key, value)

	var code int
	_, err := kv.Delete(key, q)
	if err != nil {
		w.Err("[ERROR > lib/consul.go:69]@[CONSUL] it was not possible to destroy the KV pair '"+pair+"'.")
		code = 1
	} else {
		w.Info("[INFO]@[CONSUL] the KV pair '"+pair+"' has been destroyed.")
		code = 0
	}
	return code
}

// ListKV gets KV pair of Consul
func ListKV(keyfix string, datacenter string) api.KVPairs {
	// Options to List() function
	q := &api.QueryOptions{Datacenter: datacenter}
	_, kv := Client()

	// Gets the list of KV pairs
	pairs, _, err := kv.List(keyfix, q)
	if err != nil {
		w.Err("[ERROR > lib/consul.go:87]@[CONSUL] it was not possible to get the KV pairs.")
	}
	return pairs
}

// PutKV puts a new KV pair in Consul
func PutKV(key string, value string, datacenter string) int {
	// Options to Put() function
	p := &api.KVPair{Key: key, Value: []byte(value)}
	q := &api.WriteOptions{Datacenter: datacenter}
	_, kv := Client()

	var code int
	_, err := kv.Put(p, q)
	if err != nil {
		w.Err("[ERROR > lib/consul.go:102]@[CONSUL] it was not possible to edit the KV pair.")
		code = 1
	} else {
		code = 0
	}
	return code
}
