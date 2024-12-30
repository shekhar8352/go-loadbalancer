package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockServer is a mock implementation of the Server interface.
type MockServer struct {
	addr      string
	isAlive   bool
	callCount int
}

func (m *MockServer) Address() string {
	return m.addr
}

func (m *MockServer) IsAlive() bool {
	return m.isAlive
}

func (m *MockServer) Serve(rw http.ResponseWriter, req *http.Request) {
	m.callCount++
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("Request served by " + m.addr))
}

func TestLoadBalancer_RoundRobin(t *testing.T) {
	// Create mock servers
	server1 := &MockServer{addr: "http://server1.com", isAlive: true}
	server2 := &MockServer{addr: "http://server2.com", isAlive: true}
	server3 := &MockServer{addr: "http://server3.com", isAlive: true}

	// Initialize the load balancer
	lb := NewLoadBalancer("8000", []Server{server1, server2, server3})

	// Create a request and response recorder
	req := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()

	// Call the load balancer multiple times and check round-robin behavior
	lb.serveProxy(rw, req)
	if server1.callCount != 1 {
		t.Errorf("Expected server1 to be called once, got %d", server1.callCount)
	}

	lb.serveProxy(rw, req)
	if server2.callCount != 1 {
		t.Errorf("Expected server2 to be called once, got %d", server2.callCount)
	}

	lb.serveProxy(rw, req)
	if server3.callCount != 1 {
		t.Errorf("Expected server3 to be called once, got %d", server3.callCount)
	}

	// Ensure the round-robin starts over
	lb.serveProxy(rw, req)
	if server1.callCount != 2 {
		t.Errorf("Expected server1 to be called twice, got %d", server1.callCount)
	}
}

func TestLoadBalancer_SkipInactiveServers(t *testing.T) {
	// Create mock servers
	server1 := &MockServer{addr: "http://server1.com", isAlive: true}
	server2 := &MockServer{addr: "http://server2.com", isAlive: false}
	server3 := &MockServer{addr: "http://server3.com", isAlive: true}

	// Initialize the load balancer
	lb := NewLoadBalancer("8000", []Server{server1, server2, server3})

	// Create a request and response recorder
	req := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()

	// Call the load balancer and check that it skips inactive servers
	lb.serveProxy(rw, req)
	if server1.callCount != 1 {
		t.Errorf("Expected server1 to be called once, got %d", server1.callCount)
	}

	lb.serveProxy(rw, req)
	if server2.callCount != 0 {
		t.Errorf("Expected server2 to not be called, got %d", server2.callCount)
	}
	if server3.callCount != 1 {
		t.Errorf("Expected server3 to be called once, got %d", server3.callCount)
	}

	// Ensure the round-robin skips server2 and returns to server1
	lb.serveProxy(rw, req)
	if server1.callCount != 2 {
		t.Errorf("Expected server1 to be called twice, got %d", server1.callCount)
	}
}

func TestLoadBalancer_ServeRequests(t *testing.T) {
	// Create mock servers
	server1 := &MockServer{addr: "http://server1.com", isAlive: true}
	server2 := &MockServer{addr: "http://server2.com", isAlive: true}

	// Initialize the load balancer
	lb := NewLoadBalancer("8000", []Server{server1, server2})

	// Create a request and response recorder
	req := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()

	// Call the load balancer and serve the request
	lb.serveProxy(rw, req)

	// Check if the response is correct
	resp := rw.Result()
	body := rw.Body.String()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}

	expectedBody := "Request served by http://server1.com"
	if body != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, body)
	}
}
