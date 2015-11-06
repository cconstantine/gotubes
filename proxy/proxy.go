package proxy

import (
	"fmt"
	"io"
	"net"
	"os"

  "github.com/cconstantine/gotubes/logging"
)

type PortMapper interface {
	GetRandomPort() (string, error)
}

func (s* Server) Run() {
	logging.Info.Printf("Proxying from %s\n", s.LocalAddr)

	laddr, err := net.ResolveTCPAddr("tcp", s.LocalAddr)
	check(err)
	listener, err := net.ListenTCP("tcp", laddr)
	check(err)

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			logging.Info.Printf("Failed to accept connection '%s'\n", err)
			continue
		}
		s.Connid++

		raddr_string, err := s.Ports.GetRandomPort()
		if err != nil {
			conn.Close()
			logging.Info.Printf("Unable to find a proxy target: %s", err)
			continue
		}
		raddr, err := net.ResolveTCPAddr("tcp", raddr_string)
		if err != nil {
			conn.Close()
			logging.Info.Printf("Unable to resolve target: %s", err)
			continue
		}
		p := &proxy{
			lconn:  conn,
			laddr:  laddr,
			raddr:  raddr,
			erred:  false,
			errsig: make(chan bool),
			prefix: fmt.Sprintf("Connection #%03d ", s.Connid),
			verbose: s.Verbose,
		}
		go p.start()
	}
}

type Server struct {
	LocalAddr string
	Verbose   bool
	Connid    uint32
	Ports     PortMapper
}

//A proxy represents a pair of connections and their state
type proxy struct {
	sentBytes     int64
	receivedBytes int64
	laddr, raddr  *net.TCPAddr
	lconn, rconn  *net.TCPConn
	erred         bool
	errsig        chan bool
	prefix        string
  verbose       bool

}

func (p *proxy) log(s string, args ...interface{}) {
	logging.Info.Printf(p.prefix+s+"\n", args...)
}

func (p *proxy) err(s string, err error) {
	if p.erred {
		return
	}
	if err != io.EOF {
		fmt.Printf(p.prefix+s+"\n", err)
	}
	p.errsig <- true
	p.erred = true
}

func (p *proxy) start() {
	defer p.lconn.Close()
	//connect to remote

	rconn, err := net.DialTCP("tcp", nil, p.raddr)
	if err != nil {
		p.err("Remote connection failed: %s", err)
		return
	}
	p.rconn = rconn
	defer p.rconn.Close()

	//display both ends
	p.log("Opened %s >>> %s", p.lconn.LocalAddr().String(), p.rconn.RemoteAddr().String())
	//bidirectional copy
	lfinished := make(chan int64)
	rfinished := make(chan int64)

	go p.pipe(p.lconn, p.rconn, lfinished)
	go p.pipe(p.rconn, p.lconn, rfinished)
	//wait for close...
	select {
	case p.receivedBytes = <-lfinished:
		p.rconn.Close()
		p.sentBytes = <-rfinished
	case p.sentBytes = <-rfinished:
		p.lconn.Close()
		p.receivedBytes = <-lfinished
	case <-p.errsig:
	}

	p.log("Closed (%d bytes sent, %d bytes recieved)", p.sentBytes, p.receivedBytes)
}

func (p *proxy) pipe(src, dst *net.TCPConn, finished chan int64) {
	bytesCopied, _ := io.Copy(dst, src)
	finished <- bytesCopied
}

//helper functions

func check(err error) {
	if err != nil {
		logging.Info.Printf(err.Error())
		os.Exit(1)
	}
}
