package ws

import (
	"errors"
	log "github.com/kuhufu/cm/logger"
	"net"
	"net/http"
	"strings"
	"sync"
)
import "github.com/gorilla/websocket"

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
	opts      Options
	scheme    string
	host      string
	path      string
	upgrader  websocket.Upgrader
	exitC     chan struct{}
	connC     chan net.Conn
	closeOnce sync.Once
	addr      net.Addr
}

func Listen(network, addr string, opts Options) (*Listener, error) {
	if network != "ws" && network != "wss" {
		return nil, errors.New("not support network: " + network)
	}

	if err := opts.Init(); err != nil {
		return nil, err
	}

	var path = "/"
	index := strings.IndexByte(addr, '/')
	if index >= 0 {
		path = addr[index:]
		addr = addr[:index]
	}

	ln := &Listener{
		scheme: network,
		opts:   opts,
		host:   addr,
		path:   path,
		connC:  make(chan net.Conn, 4),
		exitC:  make(chan struct{}),
		addr: &Addr{
			network: network,
			addr:    addr,
		},
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	go ln.runUpgrader()
	return ln, nil
}

func (w *Listener) Accept() (net.Conn, error) {
	select {
	case <-w.exitC:
		return nil, errors.New("listener closed")
	case conn := <-w.connC:
		return conn, nil
	}
}

func (w *Listener) runUpgrader() {
	opts := w.opts

	opts.ServeMux.HandleFunc(w.path, func(writer http.ResponseWriter, reader *http.Request) {
		log.Println("收到ws升级请求")
		conn, err := w.upgrader.Upgrade(writer, reader, nil)
		if err != nil {
			log.Error("ws升级失败:", err)
			writer.Write([]byte(err.Error()))
			return
		}
		w.connC <- &Conn{
			Conn:         conn,
			ReadTimeout:  w.opts.ReadTimeout,
			WriteTimeout: w.opts.WriteTimeout,
		}
	})

	var err error
	switch w.scheme {
	case "ws":
		log.Printf("http://%v%v", w.host, w.path)
		err = http.ListenAndServe(w.host, nil)
	case "wss":
		log.Printf("https://%v%v", w.host, w.path)
		err = http.ListenAndServeTLS(w.host, opts.CertFile, opts.KeyFile, nil)
	}

	if err != nil {
		log.Println(err)
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
