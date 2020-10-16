package ws

import (
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
)
import "github.com/gorilla/websocket"

var defaultUpgrader = websocket.Upgrader{
	HandshakeTimeout: 0,
	ReadBufferSize:   0,
	WriteBufferSize:  0,
	WriteBufferPool:  nil,
	Subprotocols:     nil,
	Error:            nil,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	EnableCompression: false,
}

type Addr struct {
	network string
	addr    string
}

func (a *Addr) Network() string {
	return a.network
}

func (a *Addr) String() string {
	return a.addr
}

type Listener struct {
	Address   string
	Host      string
	Path      string
	upgrader  websocket.Upgrader
	exitC     chan struct{}
	connC     chan *Conn
	closeOnce sync.Once
	addr      net.Addr
}

func Listen(network, addr string) (*Listener, error) {
	if network != "ws" && network != "wss" {
		return nil, errors.New("not support network: " + network)
	}

	var path = "/"
	index := strings.IndexByte(addr, '/')
	if index >= 0 {
		path = addr[index:]
		addr = addr[:index]
	}

	w := &Listener{
		Address: addr,
		Path:    path,
		connC:   make(chan *Conn, 4),
		addr: &Addr{
			network: network,
			addr:    addr,
		},
		upgrader: defaultUpgrader,
	}
	go w.RunHttpUpgrader()
	return w, nil
}

func (w *Listener) Accept() (net.Conn, error) {
	select {
	case <-w.exitC:
		return nil, errors.New("listener closed")
	case res := <-w.connC:
		return &Conn{conn: res.conn}, nil
	}
}

func (w *Listener) RunHttpUpgrader() {
	http.HandleFunc(w.Path, func(writer http.ResponseWriter, reader *http.Request) {
		log.Println("收到ws升级请求")
		conn, err := w.upgrader.Upgrade(writer, reader, nil)
		if err != nil {
			log.Println("ws升级失败")
			writer.Write([]byte(err.Error()))
			return
		}
		w.connC <- &Conn{conn: conn}
	})

	log.Printf("http://%v%v", w.Address, w.Path)

	err := http.ListenAndServe(w.Address, nil)
	if err != nil {
		panic(err)
	}
}

func (w *Listener) Close() error {
	w.closeOnce.Do(func() {
		close(w.exitC)
	})
	return nil
}

func (w *Listener) Addr() net.Addr {
	return w.addr
}
