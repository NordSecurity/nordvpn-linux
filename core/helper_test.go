package core

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
)

const (
	TestdataPath       = "testdata/"
	TestUserCreateJSON = TestdataPath + "usercreate.json"
	TestPlansJSON      = TestdataPath + "plans.json"
)

const (
	// Use different ports than rpc package does to avoid port collision when all tests are
	// executed.
	GeneralInfo int = iota
	InvalidInfo
)

var (
	workingServer *mock.HTTPTestServer // server used when GeneralInfo is used in tests
	brokenServer  *mock.HTTPTestServer // server returns broken responses, for InvalidInfo
)

func partialReadFromFile(fileName string) string {
	byteData, _ := internal.FileRead(fileName)
	return string(byteData[:len(byteData)/2])
}

func TestMain(m *testing.M) {
	// make local servers for functions relying on API
	workingServer = mock.NewHTTPTestServer(nil,
		[]mock.Handler{
			mock.NewHandleWithFileContent("/v1/users", TestUserCreateJSON),
			mock.NewHandleWithFileContent("/v1/plans", TestPlansJSON),
		},
	)

	brokenServer = mock.NewHTTPTestServer(nil,
		[]mock.Handler{
			mock.NewHandleWithResponse("/v1/users", partialReadFromFile(TestUserCreateJSON)),
			mock.NewHandleWithResponse("/v1/plans", partialReadFromFile(TestPlansJSON)),
		},
	)

	servers := []*mock.HTTPTestServer{workingServer, brokenServer}
	for _, server := range servers {
		server.Start()
	}

	res := m.Run()

	for _, server := range servers {
		server.Close()
	}
	os.Exit(res)
}

// testNewSimpleAPI returns a pointer to initialized and
// ready for use in tests SimpleAPI
func testNewSimpleAPI(serverType int) RawClientAPI {
	var baseURL string
	if serverType == GeneralInfo {
		baseURL = workingServer.URL()
	} else if serverType == InvalidInfo {
		baseURL = brokenServer.URL()
	} else {
		panic(fmt.Sprintf("Invalid value: %d", serverType))
	}
	return NewSimpleAPI(
		"",
		baseURL,
		http.DefaultClient,
		response.NoopValidator{},
	)
}
