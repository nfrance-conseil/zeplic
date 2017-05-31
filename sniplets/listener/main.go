// sniplet listener for json orders
// deserilalize json

package main;

import (
	"bufio"
	"fmt"
	"encoding/json"
	"net"
	"os"
)
type  ZFSOrder struct {
	OrderUUID string
	Action string
	Dataset string
	Name string
}

func listenServer(c net.Conn) {
	reader := bufio.NewReader(c)
	data, _ := reader.ReadString('\n')
	var f ZFSOrder
	err := json.Unmarshal([]byte(data), &f)
	if err != nil {
		println("Json error", err.Error())
		return
	}
	fmt.Printf("Order : %s is to %s\n", f.OrderUUID, f.Action)
	c.Close()
}

func main () {
	err := os.Remove("/tmp/listener.sock")
	l, err := net.Listen("unix", "/tmp/listener.sock")
	if err != nil {
		println("Listen error", err.Error())
		return
	}
	for {
		fd, err := l.Accept()
		defer fd.Close()
		if err != nil {
			println("Accept error", err.Error())
			return
		}
		go listenServer(fd)
	}
}
