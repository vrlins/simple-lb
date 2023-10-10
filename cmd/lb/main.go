package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/vrlins/simple-lb/pkg/serverpool"
	"github.com/vrlins/simple-lb/pkg/utils"
)

var sp serverpool.ServerPool

func LoadBalancer(w http.ResponseWriter, r *http.Request) {
	attempts := utils.GetAttemptsFromContext(r)
	if attempts > utils.MaxAttempts {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	peer := sp.GetNextPeer()
	if peer != nil {
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func main() {
	var serverList string
	var port int
	flag.StringVar(&serverList, "backends", "", "Load balanced backends, use commas to separate")
	flag.IntVar(&port, "port", utils.DefaultPort, "Port to serve")
	flag.Parse()

	if len(serverList) == 0 {
		log.Fatal("Please provide one or more backends to load balance")
	}

	serverpool.InitializeServerPool(serverList, &sp)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(LoadBalancer),
	}

	go serverpool.HealthCheckRoutine(&sp)

	log.Printf("Load Balancer started at :%d\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
