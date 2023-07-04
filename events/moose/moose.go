//go:build moose

// Package moose provides convenient wrappers for event sending.
package moose

// #cgo amd64 LDFLAGS: -L${SRCDIR}/../../bin/deps/nord/amd64/latest -lnord
// #cgo 386 LDFLAGS: -L${SRCDIR}/../../bin/deps/nord/i386/latest -lnord
// #cgo arm LDFLAGS: -L${SRCDIR}/../../bin/deps/nord/armel/latest -lnord
// #cgo arm LDFLAGS: -L${SRCDIR}/../../bin/deps/nord/armhf/latest -lnord
// #cgo arm64 LDFLAGS: -L${SRCDIR}/../../bin/deps/nord/aarch64/latest -lnord
// #cgo LDFLAGS: -ldl -lm
import "C"
import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/distro"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/request"

	moose "moose/events"
	worker "moose/worker"
)

const (
	mooseVersion                 = "0.29.1"
	workerVersion                = "5.2.0"
	eventEncoding                = "application/json"
	eventEndpoint                = "/app-events"
	errCodeEventSendSuccess      = 0
	errCodeEventSendDisabled     = 1
	errCodeRequestCreationFailed = 2
	errCodeRequestDoFailed       = 3
	errCodeResponseStatus        = 4
)

// Subscriber listen events, send to moose engine
type Subscriber struct {
	connectedAt   time.Time
	EventsDbPath  string
	Config        config.Manager
	Version       string
	Environment   string
	Domain        string
	Subdomain     string
	DeviceID      string
	currentDomain string
	enabled       bool
	mux           sync.RWMutex
}

// Enable moose analytics engine
func (s *Subscriber) Enable() error {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.enabled {
		return nil
	}
	s.enabled = true
	return s.mooseInit()
}

// Disable moose analytics engine
func (s *Subscriber) Disable() error {
	s.mux.Lock()
	defer s.mux.Unlock()
	if !s.enabled {
		return nil
	}
	s.enabled = false
	if err := s.response(worker.Stop()); err != nil {
		return err
	}
	return s.response(moose.Deinit())
}

func (s *Subscriber) isEnabled() bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.enabled
}

// mooseInit initializes moose libs
func (s *Subscriber) mooseInit() error {
	var cfg config.Config
	if err := s.Config.Load(&cfg); err != nil {
		return err
	}

	deviceType := "server"
	if _, err := exec.LookPath("xrandr"); err == nil {
		deviceType = "desktop"
	}

	err := s.updateEventDomain()
	if err != nil {
		return fmt.Errorf("initializing event domain: %w", err)
	}

	if err := s.response(moose.Init(
		s.EventsDbPath,
		"linux-app",
		s.Version,
		mooseVersion,
		internal.IsProdEnv(s.Environment),
		initCallback,
		errorCallback,
	)); err != nil {
		if !strings.Contains(err.Error(), "moose: already initiated") {
			return fmt.Errorf("starting tracker: %w", err)
		}
	}

	timeBetweenEvents, _ := time.ParseDuration("100ms")
	timeBetweenBatchesOfEvents, _ := time.ParseDuration("1s")
	if internal.IsProdEnv(s.Environment) {
		timeBetweenEvents, _ = time.ParseDuration("2s")
		timeBetweenBatchesOfEvents, _ = time.ParseDuration("2h")
	}
	sendEvents := true
	var batchSize uint = 20
	compressRequest := true

	if err := s.response(worker.Start(
		s.EventsDbPath,
		workerVersion,
		s.currentDomain,
		uint64(timeBetweenEvents.Milliseconds()),
		uint64(timeBetweenBatchesOfEvents.Milliseconds()),
		sendEvents,
		batchSize,
		compressRequest,
	)); err != nil {
		return fmt.Errorf("starting worker: %w", err)
	}
	if err := s.response(moose.Set_context_device_timeZone(internal.Timezone())); err != nil {
		return fmt.Errorf("setting moose time zone: %w", err)
	}

	distroVersion, err := distro.ReleasePrettyName()
	if err != nil {
		return fmt.Errorf("determining device os: %w", err)
	}
	if err := s.response(moose.Set_context_device_os(distroVersion)); err != nil {
		return fmt.Errorf("setting moose device os: %w", err)
	}
	if err := s.response(moose.Set_context_device_fp(s.DeviceID)); err != nil {
		return fmt.Errorf("setting moose device: %w", err)
	}
	var deviceT moose.Enum_SS_NordvpnappDeviceType
	switch deviceType {
	case "desktop":
		deviceT = moose.Enum_SS_NordvpnappDeviceType(moose.DeviceTypeDesktop)
	case "server":
		deviceT = moose.Enum_SS_NordvpnappDeviceType(moose.DeviceTypeServer)
	default:
		deviceT = moose.Enum_SS_NordvpnappDeviceType(moose.DeviceTypeUndefined)
	}
	if err := s.response(moose.Set_context_device_type(deviceT)); err != nil {
		return fmt.Errorf("setting moose device type: %w", err)
	}

	if err := s.response(moose.Set_context_application_config_currentState_isOnVpn_value(false)); err != nil {
		return fmt.Errorf("setting moose is on vpn: %w", err)
	}

	sub := &Subscriber{}
	if err := sub.NotifyFirewall(true); err != nil {
		return fmt.Errorf("setting moose firewall: %w", err)
	}

	if err := sub.NotifyProtocol(cfg.AutoConnectData.Protocol); err != nil {
		return fmt.Errorf("setting moose protocol: %w", err)
	}

	if err := sub.NotifyTechnology(cfg.Technology); err != nil {
		return fmt.Errorf("setting moose technology: %w", err)
	}
	return nil
}

func (s *Subscriber) NotifyKillswitch(data bool) error {
	return s.response(moose.Set_context_application_config_userPreferences_killSwitchEnabled_value(data))
}

func (s *Subscriber) NotifyAccountCheck(core.ServicesResponse) error { return nil }

func (s *Subscriber) NotifyAutoconnect(data bool) error {
	return s.response(moose.Set_context_application_config_userPreferences_autoConnectEnabled_value(data))
}

func (s *Subscriber) NotifyDefaults(any) error { return nil }

func (s *Subscriber) NotifyDNS(data events.DataDNS) error {
	if err := s.response(moose.Set_context_application_config_userPreferences_customDnsEnabled_meta(fmt.Sprintf(`{"count":%d}`, len(data.Ips)))); err != nil {
		return err
	}
	return s.response(moose.Set_context_application_config_userPreferences_customDnsEnabled_value(data.Enabled))
}

func (s *Subscriber) NotifyFirewall(data bool) error {
	return s.response(moose.Set_context_application_config_userPreferences_firewallEnabled_value(data))
}

func (s *Subscriber) NotifyRouting(data bool) error {
	return s.response(moose.Set_context_application_config_userPreferences_routingEnabled_value(data))
}

func (s *Subscriber) NotifyIpv6(data bool) error {
	if err := s.response(moose.Set_context_application_config_currentState_ipv6Enabled_value(data)); err != nil {
		return err
	}
	return s.response(moose.Set_context_application_config_currentState_ipv6Enabled_value(data))
}

func (s *Subscriber) NotifyLogin(any) error { return nil }

func (s *Subscriber) NotifyRate(data events.ServerRating) error {
	return s.response(moose.Send_userInterface_uiItems_click(
		"server_speed_rating",
		data.Server,
		moose.Enum_SS_NordvpnappUserInterfaceItemType(moose.UserInterfaceItemTypeButton),
		fmt.Sprintf("%d", data.Rate),
	))
}

func (s *Subscriber) NotifyHeartBeat(timePeriodMinutes int) error {
	return s.response(moose.Send_serviceQuality_status_heartbeat(timePeriodMinutes))
}

func (s *Subscriber) NotifyNotify(bool) error { return nil }

func (s *Subscriber) NotifyMeshnet(data bool) error {
	return s.response(moose.Set_context_application_config_userPreferences_meshnetEnabled_value(data))
}

func (s *Subscriber) NotifyObfuscate(data bool) error {
	return s.response(moose.Set_context_application_config_userPreferences_obfuscationEnabled_value(data))
}

func (s *Subscriber) NotifyPeerUpdate([]string) error { return nil }

func (s *Subscriber) NotifySelfRemoved(any) error { return nil }

func (s *Subscriber) NotifyThreatProtectionLite(data bool) error {
	return s.response(moose.Set_context_application_config_userPreferences_threatProtectionLiteEnabled_value(data))
}

func (s *Subscriber) NotifyProtocol(data config.Protocol) error {
	var protocol moose.Enum_SS_NordvpnappVpnConnectionProtocol
	switch data {
	case config.Protocol_UDP:
		protocol = moose.Enum_SS_NordvpnappVpnConnectionProtocol(moose.VpnConnectionProtocolUdp)
	case config.Protocol_TCP:
		protocol = moose.Enum_SS_NordvpnappVpnConnectionProtocol(moose.VpnConnectionProtocolTcp)
	case config.Protocol_UNKNOWN_PROTOCOL:
		fallthrough
	default:
		protocol = moose.Enum_SS_NordvpnappVpnConnectionProtocol(moose.VpnConnectionProtocolRecommended)
	}
	if err := s.response(moose.Set_context_application_config_currentState_protocol_value(protocol)); err != nil {
		return err
	}
	return s.response(moose.Set_context_application_config_userPreferences_protocol_value(protocol))
}

func (s *Subscriber) NotifyWhitelist(data events.DataWhitelist) error {
	enabled := data.UDPPorts != 0 || data.TCPPorts != 0 || data.Subnets != 0
	if err := s.response(moose.Set_context_application_config_userPreferences_splitTunnelingEnabled_meta(fmt.Sprintf(`{"udp_ports":%d,"tcp_ports:%d,"subnets":%d}`, data.UDPPorts, data.TCPPorts, data.Subnets))); err != nil {
		return err
	}
	return s.response(moose.Set_context_application_config_userPreferences_splitTunnelingEnabled_value(enabled))
}

func (s *Subscriber) NotifyTechnology(data config.Technology) error {
	var technology moose.Enum_SS_NordvpnappVpnConnectionTechnology
	switch data {
	case config.Technology_NORDLYNX:
		technology = moose.Enum_SS_NordvpnappVpnConnectionTechnology(moose.VpnConnectionTechnologyNordlynx)
	case config.Technology_OPENVPN:
		technology = moose.Enum_SS_NordvpnappVpnConnectionTechnology(moose.VpnConnectionTechnologyOpenvpn)
	case config.Technology_UNKNOWN_TECHNOLOGY:
		return errors.New("unknown technology")
	default:
		technology = moose.Enum_SS_NordvpnappVpnConnectionTechnology(moose.VpnConnectionTechnologyRecommended)
	}
	if err := s.response(moose.Set_context_application_config_currentState_technology_value(technology)); err != nil {
		return err
	}
	return s.response(moose.Set_context_application_config_userPreferences_technology_value(technology))
}

func (s *Subscriber) NotifyConnect(data events.DataConnect) error {
	s.connectedAt = time.Now()
	dnsResolutionTime := int(data.DNSResolutionTime.Milliseconds())
	var threatProtection moose.Enum_SS_NordvpnappOptBool
	if data.ThreatProtectionLite {
		threatProtection = moose.Enum_SS_NordvpnappOptBool(moose.OptBoolTrue)
	} else {
		threatProtection = moose.Enum_SS_NordvpnappOptBool(moose.OptBoolFalse)
	}

	var result moose.Enum_SS_NordvpnappEventStatus
	switch data.Type {
	case events.ConnectAttempt:
		result = moose.Enum_SS_NordvpnappEventStatus(moose.EventStatusAttempt)
	case events.ConnectSuccess:
		result = moose.Enum_SS_NordvpnappEventStatus(moose.EventStatusSuccess)
	case events.ConnectFailure:
		result = moose.Enum_SS_NordvpnappEventStatus(moose.EventStatusFailureDueToRuntimeException)
	default:
		result = moose.Enum_SS_NordvpnappEventStatus(moose.EventStatusAttempt)
	}

	var protocol moose.Enum_SS_NordvpnappVpnConnectionProtocol
	switch data.Protocol {
	case config.Protocol_TCP:
		protocol = moose.Enum_SS_NordvpnappVpnConnectionProtocol(moose.VpnConnectionProtocolTcp)
	case config.Protocol_UDP:
		protocol = moose.Enum_SS_NordvpnappVpnConnectionProtocol(moose.VpnConnectionProtocolUdp)
	case config.Protocol_UNKNOWN_PROTOCOL:
		protocol = moose.Enum_SS_NordvpnappVpnConnectionProtocol(moose.VpnConnectionProtocolNone)
	default:
		protocol = moose.Enum_SS_NordvpnappVpnConnectionProtocol(moose.VpnConnectionProtocolRecommended)
	}

	var technology moose.Enum_SS_NordvpnappVpnConnectionTechnology
	switch data.Technology {
	case config.Technology_OPENVPN:
		technology = moose.Enum_SS_NordvpnappVpnConnectionTechnology(moose.VpnConnectionTechnologyOpenvpn)
	case config.Technology_NORDLYNX:
		technology = moose.Enum_SS_NordvpnappVpnConnectionTechnology(moose.VpnConnectionTechnologyNordlynx)
	case config.Technology_UNKNOWN_TECHNOLOGY:
		technology = moose.Enum_SS_NordvpnappVpnConnectionTechnology(moose.VpnConnectionTechnologyNone)
	default:
		technology = moose.Enum_SS_NordvpnappVpnConnectionTechnology(moose.VpnConnectionTechnologyRecommended)
	}

	var server moose.Enum_SS_NordvpnappServerListSource
	if data.ServerFromAPI {
		server = moose.Enum_SS_NordvpnappServerListSource(moose.ServerListSourceRecommendedByApi)
	} else {
		server = moose.Enum_SS_NordvpnappServerListSource(moose.ServerListSourceLocallyCachedServerList)
	}

	var rule moose.Enum_SS_NordvpnappServerSelectionRule
	switch data.TargetServerSelection {
	default:
		rule = moose.Enum_SS_NordvpnappServerSelectionRule(moose.ServerSelectionRuleRecommended)
	}
	return s.response(moose.Send_serviceQuality_servers_connect(
		-1,
		dnsResolutionTime,
		result,
		moose.Enum_SS_NordvpnappEventTrigger(moose.EventTriggerUser),
		moose.Enum_SS_NordvpnappVpnConnectionPreset(moose.VpnConnectionPresetNone),
		protocol,
		data.TargetServerCity,
		data.TargetServerCountry,
		data.TargetServerDomain,
		data.TargetServerGroup,
		data.TargetServerIP,
		server,
		rule,
		technology,
		threatProtection,
		moose.Enum_SS_NordvpnappVpnConnectionTrigger(moose.VpnConnectionTriggerNone), // pass proper trigger
	))
}

func (s *Subscriber) NotifyDisconnect(data events.DataDisconnect) error {
	event := moose.Enum_SS_NordvpnappEventStatus(moose.EventStatusAttempt)
	switch data.Type {
	case events.DisconnectAttempt:
		event = moose.Enum_SS_NordvpnappEventStatus(moose.EventStatusAttempt)
	case events.DisconnectSuccess:
		event = moose.Enum_SS_NordvpnappEventStatus(moose.EventStatusSuccess)
	case events.DisconnectFailure:
		event = moose.Enum_SS_NordvpnappEventStatus(moose.EventStatusFailureDueToRuntimeException)
	}

	var technology moose.Enum_SS_NordvpnappVpnConnectionTechnology
	switch data.Technology {
	case config.Technology_OPENVPN:
		technology = moose.Enum_SS_NordvpnappVpnConnectionTechnology(moose.VpnConnectionTechnologyOpenvpn)
	case config.Technology_NORDLYNX:
		technology = moose.Enum_SS_NordvpnappVpnConnectionTechnology(moose.VpnConnectionTechnologyNordlynx)
	case config.Technology_UNKNOWN_TECHNOLOGY:
		technology = moose.Enum_SS_NordvpnappVpnConnectionTechnology(moose.VpnConnectionTechnologyNone)
	default:
		technology = moose.Enum_SS_NordvpnappVpnConnectionTechnology(moose.VpnConnectionTechnologyRecommended)
	}

	var protocol moose.Enum_SS_NordvpnappVpnConnectionProtocol
	switch data.Protocol {
	case config.Protocol_TCP:
		protocol = moose.Enum_SS_NordvpnappVpnConnectionProtocol(moose.VpnConnectionProtocolTcp)
	case config.Protocol_UDP:
		protocol = moose.Enum_SS_NordvpnappVpnConnectionProtocol(moose.VpnConnectionProtocolUdp)
	case config.Protocol_UNKNOWN_PROTOCOL:
		protocol = moose.Enum_SS_NordvpnappVpnConnectionProtocol(moose.VpnConnectionProtocolNone)
	default:
		protocol = moose.Enum_SS_NordvpnappVpnConnectionProtocol(moose.VpnConnectionProtocolRecommended)
	}

	var server moose.Enum_SS_NordvpnappServerListSource
	if data.ServerFromAPI {
		server = moose.Enum_SS_NordvpnappServerListSource(moose.ServerListSourceRecommendedByApi)
	} else {
		server = moose.Enum_SS_NordvpnappServerListSource(moose.ServerListSourceLocallyCachedServerList)
	}

	var rule moose.Enum_SS_NordvpnappServerSelectionRule
	switch data.TargetServerSelection {
	default:
		rule = moose.Enum_SS_NordvpnappServerSelectionRule(moose.ServerSelectionRuleRecommended)
	}

	var threatProtection moose.Enum_SS_NordvpnappOptBool
	if data.ThreatProtectionLite {
		threatProtection = moose.Enum_SS_NordvpnappOptBool(moose.OptBoolTrue)
	} else {
		threatProtection = moose.Enum_SS_NordvpnappOptBool(moose.OptBoolFalse)
	}
	return s.response(moose.Send_serviceQuality_servers_disconnect(
		int(time.Since(s.connectedAt).Seconds()),
		0,
		event,
		moose.Enum_SS_NordvpnappEventTrigger(moose.EventTriggerUser),
		moose.Enum_SS_NordvpnappVpnConnectionPreset(moose.VpnConnectionPresetNone),
		protocol,
		"",
		"",
		"",
		"",
		"",
		server,
		rule,
		technology,
		threatProtection,
		moose.Enum_SS_NordvpnappVpnConnectionTrigger(moose.VpnConnectionTriggerNone), // pass proper trigger
	))
}

func (s *Subscriber) NotifyRequestAPI(data events.DataRequestAPI) error {
	if data.Request == nil {
		return fmt.Errorf("request nil")
	}
	responseCode := 0
	if data.Response != nil {
		responseCode = data.Response.StatusCode
	}

	fn, err := pickNotifier(data.Request.URL.Path)
	if err != nil {
		return err
	}

	var eventStatus moose.Enum_SS_NordvpnappEventStatus
	if data.Error != nil {
		eventStatus = moose.Enum_SS_NordvpnappEventStatus(moose.EventStatusSuccess)
	} else {
		eventStatus = moose.Enum_SS_NordvpnappEventStatus(moose.EventStatusFailureDueToRuntimeException)
	}
	return s.response(fn(
		data.Request.URL.Host,
		0,
		int(data.Duration.Milliseconds()),
		eventStatus,
		moose.Enum_SS_NordvpnappEventTrigger(moose.EventTriggerApp),
		"",
		"",
		"",
		"",
		responseCode,
		"",
		data.Request.Proto,
	))
}

// sendEvent is used as a https://go.dev/ref/spec#Method_values in order be able
// to handle changing domains without involving channels.
//
// called by moose worker for each event
func (s *Subscriber) sendEvent(contentType, userAgent, requestBody string) int {
	if !s.isEnabled() {
		return errCodeEventSendDisabled
	}
	s.mux.Lock()
	domain := s.currentDomain
	s.mux.Unlock()
	req, err := request.NewRequest(
		http.MethodPost,
		userAgent,
		domain,
		eventEndpoint,
		contentType,
		fmt.Sprintf("%d", len(requestBody)),
		eventEncoding,
		strings.NewReader(requestBody),
	)
	if err != nil {
		return errCodeRequestCreationFailed
	}

	// Moose team requested specific timeout value
	client := request.NewStdHTTP(func(c *http.Client) { c.Timeout = time.Second * 30 })
	resp, err := client.Do(req)
	if err != nil {
		return errCodeRequestDoFailed
	}
	if resp.StatusCode >= 400 {
		return errCodeResponseStatus
	}
	return errCodeEventSendSuccess
}

func (s *Subscriber) updateEventDomain() error {
	domainUrl, err := url.Parse(s.Domain)
	if err != nil {
		return err
	}
	domainUrl.Host = s.Subdomain + "." + domainUrl.Host
	s.currentDomain = domainUrl.String()
	return nil
}
