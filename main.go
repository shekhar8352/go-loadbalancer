package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type Server interface {
	Address() string
	IsAlive() bool
	Serve(rw http.ResponseWriter, req *http.Request)
}

type simpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}

func (s *simpleServer) Address() string { 
	return s.addr 
}

func (s *simpleServer) IsAlive() bool {
	 return true 
}

func (s *simpleServer) Serve(rw http.ResponseWriter, req *http.Request) {
	s.proxy.ServeHTTP(rw, req)
}

// newSimpleServer returns a simple server that proxies incoming requests to the
// specified target address.
func newSimpleServer(addr string) *simpleServer {
	serverUrl, err := url.Parse(addr)
	handleErr(err)

	return &simpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

type LoadBalancer struct {
	port            string
	roundRobinCount int
	servers         []Server
}

func NewLoadBalancer(port string, servers []Server) *LoadBalancer {
	return &LoadBalancer{
		port:            port,
		roundRobinCount: 0,
		servers:         servers,
	}
}

// getNextAvailableServer selects the next available server using round-robin
// strategy. It checks the servers' availability and skips any that are not
// alive, ensuring the load balancer forwards requests to active servers only.
func (lb *LoadBalancer) getNextAvailableServer() Server {
	server := lb.servers[lb.roundRobinCount%len(lb.servers)]
	for !server.IsAlive() {
		lb.roundRobinCount++
		server = lb.servers[lb.roundRobinCount%len(lb.servers)]
	}
	lb.roundRobinCount++

	return server
}

// serveProxy forwards incoming HTTP requests to the next available server
// in the load balancer's server pool. It uses the round-robin strategy to
// select the target server and logs the forwarding action. This function
// ensures that requests are served by active servers.
func (lb *LoadBalancer) serveProxy(rw http.ResponseWriter, req *http.Request) {
	targetServer := lb.getNextAvailableServer()

	fmt.Printf("forwarding request to address %q\n", targetServer.Address())

	targetServer.Serve(rw, req)
}

func main() {
	servers := []Server{
		newSimpleServer("https://www.facebook.com"),
		newSimpleServer("https://www.bing.com"),
		newSimpleServer("https://www.duckduckgo.com"),
	}

	lb := NewLoadBalancer("8000", servers)
	handleRedirect := func(rw http.ResponseWriter, req *http.Request) {
		lb.serveProxy(rw, req)
	}

	http.HandleFunc("/", handleRedirect)

	fmt.Printf("serving requests at 'localhost:%s'\n", lb.port)

	http.ListenAndServe(":"+lb.port, nil)
}

func handleErr(err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}