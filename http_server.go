package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

const (
	readTimeout    = 45 * time.Second
	writeTimeout   = 45 * time.Second
	maxHeaderBytes = 1 << 20
	checkEvery     = 1 * time.Minute
)

type serverDetails struct {
	host       string
	rssAddress string
	ctx        context.Context
	cancelFunc context.CancelFunc
	server     *http.Server
	quit       chan bool

	parser     *gofeed.Parser
	rssContent *gofeed.Feed
}

var server *serverDetails

func execServer(quit chan bool, host, rssAddress string) {

	server = &serverDetails{
		quit:       quit,
		host:       host,
		rssAddress: rssAddress,
		parser:     gofeed.NewParser(),
	}

	server.ctx, server.cancelFunc = context.WithCancel(context.Background())
	server.server = initHTTPServer(server)

	go server.shutdown()
	go server.getFeed()

	var err error
	// Get the first content
	server.rssContent, err = server.parser.ParseURL(server.rssAddress)
	if err != nil {
		panic(err)
	}

	err = server.server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func initHTTPServer(s *serverDetails) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.root)

	server := http.Server{
		Addr:           s.host,
		Handler:        mux,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
		BaseContext: func(l net.Listener) context.Context {
			return s.ctx
		},
	}

	return &server
}

func (s *serverDetails) root(resp http.ResponseWriter, req *http.Request) {
	log.Printf("request: %+v\n", *req)
	if req.Method == http.MethodGet {
		err := generateMainTemplate(resp, s.rssContent)
		if err != nil {
			resp.WriteHeader(http.StatusInternalServerError)
			resp.Write([]byte(fmt.Sprintf("err: %s", err)))
			log.Printf("Error: %+v\n", err)
			return
		}
	}

}

func (s *serverDetails) getFeed() {
	ticker := time.NewTicker(checkEvery)
	var err error

	for {
		select {
		case <-s.quit:
			ticker.Stop()
			return
		case <-ticker.C:
			s.rssContent, err = s.parser.ParseURL(s.rssAddress)
			log.Printf("Got a new rss\n")
			if err != nil {
				panic(err)
			}
		}
	}
}

func (s *serverDetails) shutdown() error {
	select {
	case <-s.quit:
		s.cancelFunc()
		return s.server.Shutdown(s.ctx)
	}
}
