package netx

import (
	"testing"
)

func Test_UDP(t *testing.T) {
	s1 := NewUDPServer("127.0.0.1", 60000)
	s1.Name = "s2"
	err := s1.Start()
	if err != nil {
		t.Log(s1.Name+" start err", err)
		return
	}

	s2 := NewUDPServer("127.0.0.1", 60001)
	s2.Name = "s3"
	err = s2.Start()
	if err != nil {
		t.Log(s2.Name+" start err", err)
		return
	}

	s1.Send([]byte(s2.Name+" send"), s2.Addr)
	s2.Send([]byte(s1.Name+" send"), s1.Addr)
}
