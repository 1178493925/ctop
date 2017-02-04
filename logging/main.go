package logging

import (
	"net"
	"time"

	"github.com/op/go-logging"
)

const (
	size = 64
	path = "/tmp/ctop.sock"
)

var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

type CTopLogger struct {
	*logging.Logger
	backend *logging.MemoryBackend
}

func (log *CTopLogger) Serve() {
	ln, err := net.Listen("unix", path)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		go log.handler(conn)
	}
}

func (log *CTopLogger) handler(conn net.Conn) {
	defer conn.Close()
	for msg := range log.tail() {
		conn.Write([]byte(msg))
	}
}

func (log *CTopLogger) tail() chan string {
	stream := make(chan string)

	node := log.backend.Head()
	go func() {
		for {
			stream <- node.Record.Formatted(0)
			for {
				nnode := node.Next()
				if nnode != nil {
					node = nnode
					break
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()

	return stream
}

func New(serverEnabled string) *CTopLogger {

	log := &CTopLogger{
		logging.MustGetLogger("ctop"),
		logging.NewMemoryBackend(size),
	}

	logging.SetBackend(logging.NewBackendFormatter(log.backend, format))
	log.Info("initialized logging")

	if serverEnabled == "1" {
		go log.Serve()
	}

	return log
}
