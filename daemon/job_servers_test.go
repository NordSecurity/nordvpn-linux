package daemon

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

type mockServersAPI struct{}

func (mockServersAPI) Servers() (core.Servers, http.Header, error) {
	return serversList(), nil, nil
}

func (mockServersAPI) RecommendedServers(filter core.ServersFilter, _ float64, _ float64) (core.Servers, http.Header, error) {
	if filter.Group == config.DedicatedIP {
		return nil, nil, fmt.Errorf("API must not be called for Dedicated IP")
	}

	var servers core.Servers
	for _, server := range serversList() {
		if server.Status != core.Online || isDedicatedIP(server) {
			continue
		}

		servers = append(servers, server)
	}

	return servers, nil, nil
}

func (mockServersAPI) Server(serverID int64) (*core.Server, error) {
	for _, server := range serversList() {
		if server.ID == serverID {
			return &server, nil
		}
	}

	return nil, fmt.Errorf("not found")
}

func (mockServersAPI) ServersCountries() (core.Countries, http.Header, error) {
	return countriesList(), nil, nil
}

func (mockServersAPI) ServersTechnologiesConfigurations(string, int64, core.ServerTechnology) ([]byte, error) {
	return nil, nil
}

type mockFailingServersAPI struct{}

func (mockFailingServersAPI) Servers() (core.Servers, http.Header, error) {
	return nil, nil, fmt.Errorf("500")
}

func (mockFailingServersAPI) RecommendedServers(core.ServersFilter, float64, float64) (core.Servers, http.Header, error) {
	return nil, nil, fmt.Errorf("500")
}

func (mockFailingServersAPI) Server(int64) (*core.Server, error) {
	return nil, fmt.Errorf("500")
}

func (mockFailingServersAPI) ServersCountries() (core.Countries, http.Header, error) {
	return nil, nil, fmt.Errorf("500")
}

func (mockFailingServersAPI) ServersTechnologiesConfigurations(string, int64, core.ServerTechnology) ([]byte, error) {
	return nil, fmt.Errorf("500")
}

type mockConfigManager struct {
	c config.Config
}

func newMockConfigManager() *mockConfigManager {
	return &mockConfigManager{c: config.Config{
		Firewall:  true,
		UsersData: &config.UsersData{Notify: config.UidBoolMap{}, NotifyOff: config.UidBoolMap{}, TrayOff: config.UidBoolMap{}},
		TokensData: map[int64]config.TokenData{
			1337: {
				OpenVPNUsername: "bad",
				OpenVPNPassword: "actor",
				TokenExpiry:     time.Now().Add(time.Hour * 1).Format(internal.ServerDateFormat),
				ServiceExpiry:   time.Now().Add(time.Hour * 1).Format(internal.ServerDateFormat),
			},
		},
		AutoConnectData: config.AutoConnectData{
			ID:       1337,
			Protocol: config.Protocol_UDP,
		},
		Technology: config.Technology_OPENVPN,
		Mesh:       true,
		MeshDevice: &mesh.Machine{
			ID: uuid.New(),
		},
	}}
}

func (m *mockConfigManager) SaveWith(f config.SaveFunc) error {
	m.c = f(m.c)
	return nil
}

func (m *mockConfigManager) Load(c *config.Config) error {
	c.Technology = m.c.Technology
	c.Firewall = m.c.Firewall
	c.Routing = m.c.Routing
	c.KillSwitch = m.c.KillSwitch
	c.AutoConnect = m.c.AutoConnect
	c.IPv6 = m.c.IPv6
	c.AutoConnectData = m.c.AutoConnectData
	c.UsersData = m.c.UsersData
	c.TokensData = m.c.TokensData
	c.MachineID = m.c.MachineID
	c.Meshnet = m.c.Meshnet
	c.Mesh = m.c.Mesh
	c.MeshDevice = m.c.MeshDevice
	c.MeshPrivateKey = m.c.MeshPrivateKey
	c.VirtualLocation = m.c.VirtualLocation
	return nil
}

func (m *mockConfigManager) Reset() error {
	*m = *newMockConfigManager()
	return nil
}

type failingConfigManager struct {
}

func (failingConfigManager) SaveWith(f config.SaveFunc) error {
	return errors.New("failed")
}

func (failingConfigManager) Load(c *config.Config) error {
	return errors.New("failed")
}

func (failingConfigManager) Reset() error {
	return errors.New("failed")
}

// TestJobServers and its sub-tests check if the servers list gets populated properly
func TestJobServers(t *testing.T) {
	category.Set(t, category.Integration)
	defer testsCleanup()
	dm := testNewDataManager()
	err := JobServers(dm, newMockConfigManager(), &mockServersAPI{}, true)()
	assert.NoError(t, err)

	t.Run("obfuscated server exists", func(t *testing.T) {
		obfsExist := false
		for _, s := range dm.GetServersData().Servers {
			if core.IsObfuscated()(s) {
				obfsExist = true
				break
			}
		}
		assert.True(t, obfsExist)
	})

	t.Run("regular server exists", func(t *testing.T) {
		servExist := false
		for _, s := range dm.GetServersData().Servers {
			if !core.IsObfuscated()(s) {
				servExist = true
				break
			}
		}
		assert.True(t, servExist)
	})

	t.Run("server with atleast one TCP technology available exists", func(t *testing.T) {
		tcpExists := false
		for _, s := range dm.GetServersData().Servers {
			if core.IsConnectableVia(core.OpenVPNTCP)(s) ||
				core.IsConnectableVia(core.OpenVPNTCPObfuscated)(s) {
				tcpExists = true
				break
			}
		}
		assert.True(t, tcpExists)
	})

	t.Run("server with atleast one UDP technology available exists", func(t *testing.T) {
		udpExists := false
		for _, s := range dm.GetServersData().Servers {
			if core.IsConnectableVia(core.OpenVPNUDP)(s) ||
				core.IsConnectableVia(core.OpenVPNUDPObfuscated)(s) {
				udpExists = true
				break
			}
		}
		assert.True(t, udpExists)
	})

	t.Run("openvpn server exists", func(t *testing.T) {
		isOVPN := false
		for _, s := range dm.GetServersData().Servers {
			if core.IsConnectableVia(core.OpenVPNTCP)(s) ||
				core.IsConnectableVia(core.OpenVPNTCPObfuscated)(s) ||
				core.IsConnectableVia(core.OpenVPNUDP)(s) ||
				core.IsConnectableVia(core.OpenVPNUDPObfuscated)(s) {
				isOVPN = true
				break
			}
			assert.True(t, isOVPN)
		}
	})
}

// TestJobServers_InvalidData checks if unparsable document returns an error
func TestJobServers_InvalidData(t *testing.T) {
	category.Set(t, category.Integration)
	defer testsCleanup()
	dm := testNewDataManager()
	err := JobServers(dm, newMockConfigManager(), &mockFailingServersAPI{}, true)()
	assert.Error(t, err)
}

// TestJobServers_Valid checks if IsValid() condition is executed correctly
func TestJobServers_Valid(t *testing.T) {
	category.Set(t, category.Integration)

	defer testsCleanup()

	internal.FileCopy(TestdataS2DatPath, TestdataPath+TestServersFile)
	internal.FileCopy(TestdataC2DatPath, TestdataPath+TestCountryFile)
	internal.FileCopy(TestdataPath+"i2.dat", TestdataPath+TestInsightsFile)
	internal.FileCopy(TestdataVersionDatPath, TestdataPath+TestVersionFile)

	dm := testNewDataManager()
	assert.NoError(t, dm.LoadData())
	original := dm.GetServersData().Servers
	dm.SetServersData(time.Now(), original, "")

	err := JobServers(dm, newMockConfigManager(), &mockFailingServersAPI{}, true)()
	assert.NoError(t, err)
	assert.ElementsMatch(t, dm.GetServersData().Servers, original)
}

// TestJobServers_Expired checks if IsValid() condition is executed correctly
func TestJobServers_Expired(t *testing.T) {
	category.Set(t, category.Integration)

	defer testsCleanup()

	internal.FileCopy(TestdataS2DatPath, TestdataPath+TestServersFile)

	dm := testNewDataManager()
	original, _, _ := mockServersAPI{}.Servers() // do not use filesystem
	dm.SetServersData(time.Now().Add(time.Duration(-300)*time.Minute), original, "")
	err := JobServers(dm, newMockConfigManager(), &mockServersAPI{}, true)()
	assert.NoError(t, err)
	assert.False(t, reflect.DeepEqual(dm.GetServersData().Servers, original))
}
