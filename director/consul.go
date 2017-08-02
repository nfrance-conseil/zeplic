// Package director contains: agent.go - consul.go - director.go - extract.go - slave.go
//
// Consul get the status of consul member
//
package director

import (
	"github.com/hashicorp/consul/api"
)

// Alive returns if the Consul server is available
func Alive() bool {
	var alive bool

	// Create a new client
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		w.Err("[ERROR > director/consul.go:16]@[CONSUL] it was not possible to create a new client.")
	}

	// Get the operator endpoints
	agent := client.Agent()

	// Get members list
	members, _ := agent.Members(false)
	if err != nil {
		w.Err("[ERROR > director/consul.go:25]@[CONSUL] it was not possible to get the members list.")
	}

	if len(members) > 0 {
		alive = true
	}
	return alive
}
