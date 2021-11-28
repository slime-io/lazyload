package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"net/http"
	"slime.io/slime/modules/lazyload/pkg/proxy"
)

func main() {
	var addr = flag.String("addr", "0.0.0.0:80", "The addr of the application.")
	flag.Parse()

	handler := &proxy.Proxy{}

	log.Println("Starting proxy server on", *addr)
	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
