package mock

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"golang.org/x/crypto/ssh"
)

type HTTPTestServer struct {
	t          *testing.T
	srv        *httptest.Server
	mux        *http.ServeMux
	privateKey *rsa.PrivateKey
	publicKey  ssh.PublicKey
	running    atomic.Bool
}

func (s *HTTPTestServer) URL() string {
	if s.Running() {
		return s.srv.URL
	}

	return ""
}

func (s *HTTPTestServer) Start() {
	if s.Running() {
		// already running
		return
	}
	s.srv.Start()
	s.running.Store(true)
	if s.t != nil {
		s.t.Cleanup(func() { s.Close() })
	}
}

func (s *HTTPTestServer) Close() {
	if !s.Running() {
		// already closed
		return
	}
	s.running.Store(false)
	s.srv.Close()
}

func (s *HTTPTestServer) Running() bool { return s.running.Load() }

type HTTPError struct {
	Code int
	Msg  string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Msg)
}

type Handler struct {
	Pattern string
	Fn      func() ([]byte, *HTTPError)
}

func NewHandleWithResponse(pattern string, response string) Handler {
	return Handler{
		Pattern: pattern,
		Fn: func() ([]byte, *HTTPError) {
			return []byte(response), nil
		},
	}
}

func NewHandleWithFileContent(pattern string, path string) Handler {
	return Handler{
		Pattern: pattern,
		Fn: func() ([]byte, *HTTPError) {
			byteData, err := internal.FileRead(path)
			if err != nil {
				fmt.Println(internal.ErrorPrefix, "Failed to read file", path, err)
			}
			return byteData, nil
		},
	}
}

type internalHandler struct {
	h      Handler
	server *HTTPTestServer
}

func (h *internalHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	byteData, err := h.h.Fn()
	if err != nil {
		http.Error(w, err.Msg, err.Code)
		return
	}
	h.setHeaders(w, byteData)
	w.Write(byteData)
}

func (h *internalHandler) setHeaders(w http.ResponseWriter, data []byte) {
	headers, err := GenerateValidHeaders(h.server.privateKey, data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for key := range headers {
		w.Header().Set(key, headers.Get(key))
	}
	w.Header().Set("Content-Type", "application/json")
}

func NewHTTPTestServer(t *testing.T, handlers []Handler) *HTTPTestServer {
	if t != nil {
		t.Helper()
	}

	mux := http.NewServeMux()
	ts := httptest.NewUnstartedServer(mux)
	privateKey, publicKey, err := GenerateKeyPair()

	if err != nil {
		t.Fatal("Cannot generate key pair")
	}

	server := &HTTPTestServer{
		t:          t,
		srv:        ts,
		privateKey: privateKey,
		publicKey:  publicKey,
		running:    atomic.Bool{},
	}

	if len(handlers) > 0 {
		for _, h := range handlers {
			handler := internalHandler{
				h:      h,
				server: server,
			}
			mux.Handle(h.Pattern, &handler)
		}
	}

	return server
}
