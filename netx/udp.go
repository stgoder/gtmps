package netx

import (
	"fmt"
	"net"
	"strconv"
)

type UDPServer struct {
	Name string
	Host string
	Port int
	c    *net.UDPConn
	Addr *net.UDPAddr
}

func NewUDPServer(host string, port int) (s *UDPServer) {
	s = &UDPServer{
		host + ":" + strconv.Itoa(port),
		host,
		port,
		nil,
		nil,
	}
	return
}

func (s *UDPServer) Start() error {
	addr, err := net.ResolveUDPAddr("udp", s.Host+":"+strconv.Itoa(s.Port))
	if err != nil {
		return err
	}
	s.Addr = addr

	c, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	s.c = c
	go s.read()

	return nil
}

func (s *UDPServer) read() {
	defer s.c.Close()

	buf := make([]byte, 1024)
	for {
		n, addr, err := s.c.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("read from udp err", err)
		}
		data := buf[0:n]
		go s.handleData(data, addr)
	}
}

func (s *UDPServer) handleData(data []byte, addr *net.UDPAddr) {
	fmt.Println(s.Name + " recv=" + string(data))
}

func (s *UDPServer) Send(data []byte, addr *net.UDPAddr) {
	s.c.WriteToUDP(data, addr)
}
