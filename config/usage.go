// Package config contains: json.go - syslog.go - usage.go
//
// Show zeplic help
//
package config

import (
	"fmt"
)

// Usage returns user help
func Usage() {
	fmt.Printf("Usage: zeplic -z <command>\n\n")
	fmt.Printf("   agent\tListen ZFS orders from director\n")
	fmt.Printf("   director\tSend ZFS orders to agent\n")
	fmt.Printf("   quit\t\tGracefully shutdown\n")
	fmt.Printf("   reload\tRestart zeplic to sleep state\n")
	fmt.Printf("   run\t\tStart zeplic as background\n")
	fmt.Printf("   slave\tReceive a new snapshot from agent\n")
	fmt.Printf("   version\tShow version of zeplic\n")
	fmt.Println("")
}
