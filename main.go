package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/stgoder/gtmps/netx"
)

func main() {
	s1 := netx.NewTCPServer("0.0.0.0", 9000)
	s1.Name = "s1"
	err := s1.Start()
	if err != nil {
		fmt.Println("server start err", err)
		return
	}

	c1 := netx.NewTCPClient("127.0.0.1", 9000)
	c1.ID = "c1"
	err = c1.Connect()
	if c1 != nil {
		fmt.Println("client connect err", err)
	}

	for {
		s1.Print()
		c1.Send([]byte("hello-" + strconv.FormatInt(time.Now().Unix(), 10)))
		time.Sleep(time.Second * 3)
	}
}
