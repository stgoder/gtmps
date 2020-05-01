package netx

import (
	"testing"
)

func Test_TCP(t *testing.T) {
	s1 := NewTCPServer("0.0.0.0", 9000)
	s1.Name = "s1"
	err := s1.Start()
	if err != nil {
		t.Log("server start err", err)
		return
	}

	c1 := NewTCPClient("127.0.0.1", 9000)
	err = c1.Connect()
	if c1 != nil {
		t.Log("client connect err", err)
	}
	c1.Send([]byte("hello1"))
}
