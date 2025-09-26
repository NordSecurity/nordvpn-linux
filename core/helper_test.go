package core

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/mock"

	"golang.org/x/crypto/ssh"
)

const (
	TestdataPath       = "testdata/"
	LocalServerPath    = "http://localhost"
	TestUserCreateJSON = "usercreate.json"
	TestPlansJSON      = "plans.json"
)

const (
	// Use different ports than rpc package does to avoid port collision when all tests are
	// executed.
	GeneralInfo int = iota + 9602
	InvalidInfo
)

type Handler struct {
	pattern string
	f       func(http.ResponseWriter, *http.Request)
}

var privateKey *rsa.PrivateKey
var publicKey ssh.PublicKey

func TestMain(m *testing.M) {
	// make local servers for functions relying on API
	servers := make([]*http.Server, 0)
	generalInfoHandler := []Handler{
		{"/v1/users", mockAPI(TestUserCreateJSON).handler},
		{"/v1/plans", mockAPI(TestPlansJSON).handler},
	}
	servers = append(servers, StartServer(GeneralInfo, generalInfoHandler))

	invalidInfoHandler := []Handler{
		{"/v1/users", mockAPI(TestUserCreateJSON).invalidHandler},
		{"/v1/plans", mockAPI(TestPlansJSON).invalidHandler},
	}

	var err error
	privateKey, publicKey, err = mock.GenerateKeyPair()
	if err != nil {
		log.Fatalf("error on generating RSA key pair: %+v", err)
	}

	servers = append(servers, StartServer(InvalidInfo, invalidInfoHandler))
	res := m.Run()
	for _, server := range servers {
		server.Shutdown(context.Background())
	}
	os.Exit(res)
}

func localServerPath(port int) string {
	return fmt.Sprintf("%s:%d", LocalServerPath, port)
}

type mockAPI string

// handler gives content of file specified in mockAPI in http.Response with correctly set HTTP headers
func (api mockAPI) handler(w http.ResponseWriter, r *http.Request) {
	byteData, _ := internal.FileRead(TestdataPath + string(api))
	setHeaders(w, byteData)
	w.Write(byteData)
}

func (api mockAPI) invalidHandler(w http.ResponseWriter, r *http.Request) {
	byteData, _ := internal.FileRead(TestdataPath + string(api))
	byteData = byteData[:len(byteData)/2]
	setHeaders(w, byteData)
	w.Write(byteData)
}

func setHeaders(w http.ResponseWriter, data []byte) {
	headers, err := mock.GenerateValidHeaders(privateKey, data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for key := range headers {
		w.Header().Set(key, headers.Get(key))
	}
	w.Header().Set("Content-Type", "application/json")
}

func StartServer(port int, handlers []Handler) *http.Server {
	log.Println("Port:", port)
	srv := &http.Server{}

	mux := http.NewServeMux()
	if len(handlers) > 0 {
		for _, h := range handlers {
			mux.HandleFunc(h.pattern, h.f)
		}
	}

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal(err)
	}
	srv.Handler = mux

	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
		}
	}()

	// check for race conditions
	CheckServer(listener, 10)
	return srv
}

func CheckServer(listener net.Listener, attempts int) {
	resp, err := http.Get("http://" + listener.Addr().String())
	if err != nil {
		if attempts <= 0 {
			log.Fatal("Error starting server")
		}
		CheckServer(listener, attempts-1)
	}
	defer resp.Body.Close()
}

// testNewSimpleAPI returns a pointer to initialized and
// ready for use in tests SimpleAPI
func testNewSimpleAPI(port int) RawClientAPI {
	return NewSimpleAPI(
		"",
		localServerPath(port),
		http.DefaultClient,
		response.NoopValidator{},
	)
}
