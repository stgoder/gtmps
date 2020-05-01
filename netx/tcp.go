package netx

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"
)

// codec

func Decode(buf []byte, ch chan []byte) []byte {
	len := len(buf)
	var i int

	for i = 0; i < len; i = i + 1 {
		if len < 2+4+i {
			break
		}
		if buf[i] == 0xEB && buf[i+1] == 0x90 {
			length := BytesToInt(buf[i+2 : i+6])
			if len < 2+4+length+i {
				break
			}
			data := buf[2+4+i : 2+4+i+length]
			ch <- data
			i += 2 + 4 + length - 1
		}
	}

	if i == len {
		return make([]byte, 0)
	}
	return buf[i:]
}

func Encode(data []byte) []byte {
	head := []byte{0xEB, 0x90}
	return append(append([]byte(head), IntToBytes(len(data))...), data...)
}

func IntToBytes(n int) []byte {
	x := int32(n)
	bf := bytes.NewBuffer([]byte{})
	binary.Write(bf, binary.LittleEndian, x)
	return bf.Bytes()
}

func BytesToInt(b []byte) int {
	bf := bytes.NewBuffer(b)
	var x int32
	binary.Read(bf, binary.LittleEndian, &x)
	return int(x)
}

// TCP server

type TCPServer struct {
	Name  string
	Host  string
	Port  int
	Conns map[string]*Conn
	l     *net.TCPListener
	addr  *net.TCPAddr
}

func NewTCPServer(host string, port int) (s *TCPServer) {
	s = &TCPServer{
		host + ":" + strconv.Itoa(port),
		host,
		port,
		make(map[string]*Conn, 1024),
		nil,
		nil,
	}
	return
}

func (s *TCPServer) Start() error {
	addr, err := net.ResolveTCPAddr("tcp", s.Host+":"+strconv.Itoa(s.Port))
	if err != nil {
		fmt.Println("addr err=", err)
		return err
	}
	s.addr = addr

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println(s.Name+" listen err=", err)
		return err
	}
	s.l = l

	go s.accept()

	fmt.Println(s.Name + " start port " + strconv.Itoa(s.Port))

	return nil
}

func (s *TCPServer) accept() {
	defer s.l.Close()
	for {
		c, err := s.l.AcceptTCP()
		if err != nil {
			fmt.Println(s.Name+" accept err=", err)
			continue
		}

		addr := c.RemoteAddr()
		id := addr.String()
		ch := make(chan []byte, 16)
		conn := Conn{
			id,
			addr,
			time.Now().Unix(),
			0,
			c,
			ch,
			s,
		}

		go s.handleConn(conn)
	}
}

func (s *TCPServer) handleConn(c Conn) {
	defer func() {
		s.handleClose(c)
	}()

	fmt.Println(c.ID + " connect")
	s.Conns[c.ID] = &c

	tmpBuf := make([]byte, 0)
	go c.Recv()

	buf := make([]byte, 1024)
	for {
		n, err := c.c.Read(buf)
		if err != nil || n < 0 {
			return
		}
		tmpBuf = Decode(append(tmpBuf, buf[:n]...), c.ch)
	}
}

func (s *TCPServer) handleClose(c Conn) {
	fmt.Println(s.Name + " conn " + c.ID + " disconnect")
	c.Close()
	delete(s.Conns, c.ID)
}

func (s *TCPServer) Send(id string, data []byte) {
	c := s.Conns[id]
	if c != nil {
		c.Send(data)
	}
}

func (s *TCPServer) Broadcast(data []byte) {
	for _, c := range s.Conns {
		if c != nil {
			c.Send(data)
		}
	}
}

func (s *TCPServer) Print() {
	fmt.Println("Print conns")
	for _, c := range s.Conns {
		fmt.Println(c.String())
	}
}

// Connection

type Conn struct {
	ID       string
	Addr     net.Addr
	ConnTime int64
	RecvTime int64
	c        net.Conn
	ch       chan []byte
	s        *TCPServer
}

func (c *Conn) Recv() {
	for {
		select {
		case data := <-c.ch:
			fmt.Println("TCPServer " + c.s.Name + " recv=" + string(data) + " from " + c.ID)
			c.RecvTime = time.Now().Unix()
			c.Send(data)
		}
	}
}

func (c *Conn) Send(data []byte) (err error) {
	p := Encode(data)
	n := 0

	for n < len(p) {
		n1, err := c.c.Write(p)
		if err != nil {
			return err
		}

		if n1 < 0 {
			return errors.New("Conn write -")
		}

		if n1 < len(p) {
			p = p[n1:]
			n = 0
		} else {
			break
		}
	}
	return
}

func (c *Conn) Close() {
	c.c.Close()
}

func (c *Conn) String() string {
	return "ID=" + c.ID + ", Addr=" + c.Addr.String() +
		", ConnTime=" + strconv.FormatInt(c.ConnTime, 10) +
		", RecvTime=" + strconv.FormatInt(c.RecvTime, 10)
}

// TCP client

type TCPClient struct {
	ID       string
	Host     string
	Port     int
	ConnTime int64
	RecvTime int64
	c        *net.TCPConn
	ch       chan []byte
}

func NewTCPClient(host string, port int) (c *TCPClient) {
	ch := make(chan []byte, 16)
	c = &TCPClient{
		host + ":" + strconv.Itoa(port),
		host,
		port,
		0,
		0,
		nil,
		ch,
	}
	return
}

func (c *TCPClient) Connect() error {
	addr, err := net.ResolveTCPAddr("tcp", c.Host+":"+strconv.Itoa(c.Port))
	if err != nil {
		return err
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return err
	}

	c.c = conn
	c.ConnTime = time.Now().Unix()

	go c.read()

	return nil
}

func (c *TCPClient) read() {
	go c.recv()

	tmpBuf := make([]byte, 0)

	buf := make([]byte, 1024)
	for {
		n, err := c.c.Read(buf)
		if err != nil || n < 0 {
			c.c.Close()
			return
		}
		tmpBuf = Decode(append(tmpBuf, buf[:n]...), c.ch)
	}
}

func (c *TCPClient) recv() {
	for {
		select {
		case data := <-c.ch:
			fmt.Println("TCPClient " + c.ID + " recv=" + string(data) + " from " + c.c.RemoteAddr().String())
			c.RecvTime = time.Now().Unix()
		}
	}
}

func (c *TCPClient) Send(data []byte) (err error) {
	p := Encode(data)
	n := 0

	for n < len(p) {
		n1, err := c.c.Write(p)
		if err != nil {
			return err
		}

		if n1 < 0 {
			return errors.New("TCPCliend write -")
		}

		if n1 < len(p) {
			p = p[n1:]
			n = 0
		} else {
			break
		}
	}
	return
}

func (c *TCPClient) Close() {
	c.c.Close()
}
