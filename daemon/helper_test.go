package daemon

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/request"
	"github.com/NordSecurity/nordvpn-linux/test/keypair"
	"github.com/NordSecurity/nordvpn-linux/test/response"

	"golang.org/x/crypto/ssh"
)

const (
	FilesPath               = "/configs/"
	TestdataPath            = "testdata/"
	TestInsightsFile        = "tempinsights.dat"
	TestServersFile         = "tempserv.dat"
	TestCountryFile         = "tempcntr.dat"
	TestCyberSecFile        = "tempcyber.dat"
	TestVersionFile         = "tempversion.dat"
	TestTokenRenewJSON      = "tokenrenew.json"
	MixedServersJSON        = "mixed.json"
	CountryDataJSON         = "country.json"
	LocalServerPath         = "http://localhost"
	InsightsJSON            = "testinsights.json"
	TestVersionDeb          = "testdebparse"
	TestVersionRpm          = "testrpmparse"
	TestUserCreateJSON      = "usercreate.json"
	TestUserCredentialsJSON = "usercredentials.json"
	TestBadUserCreateJSON   = "badusercreate.json"
	TestPlansJSON           = "plans.json"
)

const (
	GeneralInfo int = iota + 9600
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
		{"/v1/servers", mockAPI(MixedServersJSON).handler},
		{"/v1/servers/countries", mockAPI(CountryDataJSON).handler},
		{"/v1/helpers/ips/insights/", mockAPI(InsightsJSON).handler},
		{"/v1/users", mockAPI(TestUserCreateJSON).handler},
		{"/v1/users/services/credentials/", mockAPI(TestUserCredentialsJSON).handler},
		{"/v1/users/tokens/renew", mockAPI(TestTokenRenewJSON).handler},
		{"/v1/plans", mockAPI(TestPlansJSON).handler},
	}
	servers = append(servers, StartServer(GeneralInfo, generalInfoHandler))

	invalidInfoHandler := []Handler{
		{"/v1/servers", mockAPI(MixedServersJSON).invalidHandler},
		{"/v1/servers/countries", mockAPI(MixedServersJSON).invalidHandler},
		{"/v1/users", mockAPI(MixedServersJSON).invalidHandler},
		{"/v1/plans", mockAPI(MixedServersJSON).invalidHandler},
	}

	var err error
	privateKey, publicKey, err = keypair.GenerateKeyPair()
	if err != nil {
		log.Fatalf("error on generating RSA key pair: %+v", err)
	}

	servers = append(servers, StartServer(InvalidInfo, invalidInfoHandler))
	res := m.Run()
	for _, server := range servers {
		server.Shutdown(context.Background())
	}
	testsCleanup()
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
	headers, err := response.GenerateValidHeaders(privateKey, data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for key := range headers {
		w.Header().Set(key, headers.Get(key))
	}
	w.Header().Set("Content-Type", "application/json")
}

func testsCleanup() {
	internal.FileDelete(TestdataPath + TestServersFile)
	internal.FileDelete(TestdataPath + TestCountryFile)
	internal.FileDelete(TestdataPath + TestCyberSecFile)
	internal.FileDelete(TestdataPath + TestInsightsFile)
	internal.FileDelete(TestdataPath + TestVersionFile)
}

func waitPortForListener(port int, timeoutSec int) (net.Listener, error) {
	passedTime := 0
	log.Printf("Waiting for port: %d to become available...\n", port)
	for {
		passedTime++
		listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
		if listener != nil {
			log.Printf("Port: %d is available!\n", port)
			return listener, nil
		}
		time.Sleep(1 * time.Second)
		if passedTime == timeoutSec {
			return nil, fmt.Errorf("Waiting port for listener timeouted after: %d seconds", passedTime)
		}
		log.Printf("Wait port for: %v!\n", err)
	}
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
	listener, err := waitPortForListener(port, 10)
	if err != nil {
		log.Println("Could not create listener!")
		log.Fatal(err)
	}
	fs := http.FileServer(http.Dir(TestdataPath))
	headerChange := func(h http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			byteData, err := internal.FileRead(strings.ReplaceAll(r.RequestURI, "/configs/", TestdataPath))
			if err != nil {
				log.Println("ERROR:", err)
			}
			setHeaders(w, byteData)
			h.ServeHTTP(w, r)
		}
	}
	mux.Handle(FilesPath, headerChange(http.StripPrefix(FilesPath[:len(FilesPath)-1], fs)))
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
	_, err := http.Get("http://" + listener.Addr().String())
	if err != nil {
		if attempts <= 0 {
			log.Fatal("Error starting server")
		}
		CheckServer(listener, attempts-1)
	}
}

// testNewDataManager returns a pointer to initialized and
// ready for use in tests DataManager
func testNewDataManager() *DataManager {
	return NewDataManager(
		TestdataPath+TestInsightsFile,
		TestdataPath+TestServersFile,
		TestdataPath+TestCountryFile,
		TestdataPath+TestVersionFile,
	)
}

// testNewCDNAPI returns a pointer to initialized and
// ready for use in tests CDNAPI
func testNewCDNAPI() *core.CDNAPI {
	return core.NewCDNAPI(
		"",
		localServerPath(GeneralInfo),
		request.NewHTTPClient(http.DefaultClient, nil, nil),
		nil,
	)
}

// testNewRepoAPI returns a pointer to initialized and
// ready for use in tests RepoAPI
func testNewRepoAPI() *RepoAPI {
	return NewRepoAPI(
		localServerPath(GeneralInfo),
		"",
		"",
		"",
		"",
		request.NewHTTPClient(http.DefaultClient, nil, nil),
	)
}
