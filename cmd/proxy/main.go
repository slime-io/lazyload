package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"net/http"
	"slime.io/slime/modules/lazyload/pkg/proxy"
)

func main() {
	go func() {
		handler := &proxy.HealthzProxy{}
		log.Println("Starting health check on :8080")
		if err := http.ListenAndServe(":8080", handler); err != nil {
			log.Fatal("ListenAndServe:", err)
		}
	}()

	var addr = flag.String("addr", "0.0.0.0:80", "The addr of the application.")
	flag.Parse()

	handler := &proxy.Proxy{}

	log.Println("Starting proxy server on", *addr)
	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
