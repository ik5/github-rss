package main

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"
)

const baseAddr = "https://github.com/%s.private.atom"
const defaultHost = ":8030"

func doSignals(quit chan<- bool) {
	quitSigs := make(chan os.Signal, 1)
	// hupSig := make(chan os.Signal, 1)
	defer close(quitSigs)
	// defer close(hupSig)

	// signal.Notify(hupSig, syscall.SIGHUP)
	signal.Notify(quitSigs, syscall.SIGINT,
		syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT)

	for {
		select {
		// case <-hupSig:
		case <-quitSigs:
			quit <- true
		}
	}
}

func generateRSSPath(rssToken, rssUser string) (string, error) {
	urlAddress, err := url.Parse(fmt.Sprintf(baseAddr, rssUser))
	if err != nil {
		return "", err
	}

	query := urlAddress.Query()
	query.Set("token", rssToken)
	urlAddress.RawQuery = query.Encode()

	return urlAddress.String(), nil

}

func main() {
	rssToken := os.Getenv("GHTOKEN")
	rssUser := os.Getenv("GHUSER")
	host := os.Getenv("HTTPHOST")

	if host == "" {
		host = defaultHost
	}

	if rssToken == "" {
		panic("GHTOKEN cannot be empty")
	}
	if rssUser == "" {
		panic("GHUSER cannot be empty")
	}
	rssAddress, err := generateRSSPath(rssToken, rssUser)
	if err != nil {
		panic(err)
	}

	quit := make(chan bool)

	go doSignals(quit)
	go execServer(quit, host, rssAddress)

	<-quit
}
