package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"net"
)

var connid = uint64(0)
var localAddr = flag.String("l", ":9999", "local address")
var remoteAddr = flag.String("r", "localhost:80", "remote address")
var verbose = flag.Bool("v", false, "display server actions")

func main() {
	flag.Parse()
	fmt.Printf("Proxying from %v to %v\n", *localAddr, *remoteAddr)

	laddr, err := net.ResolveTCPAddr("tcp", *localAddr)
	check(err)
	raddr, err := net.ResolveTCPAddr("tcp", *remoteAddr)
	check(err)
	listener, err := net.ListenTCP("tcp", laddr)
	check(err)

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Printf("Failed to accept connection '%s'\n", err)
			continue
		}
		connid++

		p := &proxy{
			lconn:    conn,
			laddr:    laddr,
			raddr:    raddr,
			erred:    false,
			errsig:   make(chan bool),
			prefix:   fmt.Sprintf("Connection #%03d ", connid),
		}
		go p.start()
	}
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
}

func (p *proxy) log(s string, args ...interface{}) {
	if *verbose {
		fmt.Printf(p.prefix+s, args...)
	}
}

func (p *proxy) err(s string, err error) {
	if p.erred {
		return
	}
	if err != io.EOF {
		fmt.Printf(p.prefix+s, err)
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
	p.log("Opened %s >>> %s", p.lconn.RemoteAddr().String(), p.rconn.RemoteAddr().String())
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
	case p.sentBytes =<-rfinished:
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
		fmt.Printf(err.Error())
		os.Exit(1)
	}
}



