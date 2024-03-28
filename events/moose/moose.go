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
	workerVersion                = "8.2.0"
	eventEncoding                = "application/json"
	eventEndpoint                = "/app-events"
	errCodeEventSendSuccess      = 0
	errCodeEventSendDisabled     = 1
	errCodeRequestCreationFailed = 2
	errCodeRequestDoFailed       = 3
	errCodeResponseStatus        = 4
	applicationName              = "linux-app"
)

// Subscriber listen events, send to moose engine
type Subscriber struct {
	connectionStartTime time.Time
	EventsDbPath        string
	Config              config.Manager
	Version             string
	Environment         string
	Domain              string
	Subdomain           string
	DeviceID            string
	currentDomain       string
	enabled             bool
	mux                 sync.RWMutex
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
	if err := s.response(uint32(worker.Stop())); err != nil {
		return err
	}
	return s.response(moose.MooseNordvpnappDeinit())
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

	if err := s.response(moose.MooseNordvpnappInit(
		s.EventsDbPath,
		internal.IsProdEnv(s.Environment),
		s,
		s,
	)); err != nil {
		if !strings.Contains(err.Error(), "moose: already initiated") {
			return fmt.Errorf("starting tracker: %w", err)
		}
	}

	if err := s.response(moose.NordvpnappSetContextApplicationName(applicationName)); err != nil {
		return fmt.Errorf("setting application name: %w", err)
	}

	if err := s.response(moose.NordvpnappSetContextApplicationVersion(s.Version)); err != nil {
		return fmt.Errorf("setting application version: %w", err)
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

	if err := s.response(uint32(worker.Start(
		s.EventsDbPath,
		workerVersion,
		s.currentDomain,
		uint64(timeBetweenEvents.Milliseconds()),
		uint64(timeBetweenBatchesOfEvents.Milliseconds()),
		sendEvents,
		batchSize,
		compressRequest,
	))); err != nil {
		return fmt.Errorf("starting worker: %w", err)
	}
	if err := s.response(moose.NordvpnappSetContextDeviceTimeZone(internal.Timezone())); err != nil {
		return fmt.Errorf("setting moose time zone: %w", err)
	}

	distroVersion, err := distro.ReleasePrettyName()
	if err != nil {
		return fmt.Errorf("determining device os: %w", err)
	}
	if err := s.response(moose.NordvpnappSetContextDeviceOs(distroVersion)); err != nil {
		return fmt.Errorf("setting moose device os: %w", err)
	}
	if err := s.response(moose.NordvpnappSetContextDeviceFp(s.DeviceID)); err != nil {
		return fmt.Errorf("setting moose device: %w", err)
	}
	var deviceT moose.NordvpnappDeviceType
	switch deviceType {
	case "desktop":
		deviceT = moose.NordvpnappDeviceTypeDesktop
	case "server":
		deviceT = moose.NordvpnappDeviceTypeServer
	default:
		deviceT = moose.NordvpnappDeviceTypeUndefined
	}
	if err := s.response(moose.NordvpnappSetContextDeviceType(deviceT)); err != nil {
		return fmt.Errorf("setting moose device type: %w", err)
	}

	if err := s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigCurrentStateIsOnVpnValue(false)); err != nil {
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
	return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesKillSwitchEnabledValue(data))
}

func (s *Subscriber) NotifyAccountCheck(core.ServicesResponse) error { return nil }

func (s *Subscriber) NotifyAutoconnect(data bool) error {
	return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesAutoConnectEnabledValue(data))
}

func (s *Subscriber) NotifyDefaults(any) error { return nil }

func (s *Subscriber) NotifyDNS(data events.DataDNS) error {
	if err := s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesCustomDnsEnabledMeta(fmt.Sprintf(`{"count":%d}`, len(data.Ips)))); err != nil {
		return err
	}
	return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesCustomDnsEnabledValue(len(data.Ips) > 0))
}

func (s *Subscriber) NotifyFirewall(data bool) error {
	return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesFirewallEnabledValue(data))
}

func (s *Subscriber) NotifyRouting(data bool) error {
	return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesRoutingEnabledValue(data))
}

func (s *Subscriber) NotifyIpv6(data bool) error {
	if err := s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigCurrentStateIpv6EnabledValue(data)); err != nil {
		return err
	}
	return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesIpv6EnabledValue(data))
}

func (s *Subscriber) NotifyLogin(any) error { return nil }

func (s *Subscriber) NotifyRate(data events.ServerRating) error {
	return s.response(moose.NordvpnappSendUserInterfaceUiItemsClick(
		"server_speed_rating",
		moose.NordvpnappUserInterfaceItemTypeButton,
		data.Server,
		fmt.Sprintf("%d", data.Rate),
	))
}

func (s *Subscriber) NotifyHeartBeat(timePeriodMinutes int) error {
	return s.response(moose.NordvpnappSendServiceQualityStatusHeartbeat(int32(timePeriodMinutes)))
}

func (s *Subscriber) NotifyNotify(bool) error { return nil }

func (s *Subscriber) NotifyMeshnet(data bool) error {
	return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesMeshnetEnabledValue(data))
}

func (s *Subscriber) NotifyObfuscate(data bool) error {
	return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesObfuscationEnabledValue(data))
}

func (s *Subscriber) NotifyPeerUpdate([]string) error { return nil }

func (s *Subscriber) NotifySelfRemoved(any) error { return nil }

func (s *Subscriber) NotifyThreatProtectionLite(data bool) error {
	return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigCurrentStateThreatProtectionLiteEnabledValue(data))
}

func (s *Subscriber) NotifyProtocol(data config.Protocol) error {
	var protocol moose.NordvpnappVpnConnectionProtocol
	switch data {
	case config.Protocol_UDP:
		protocol = moose.NordvpnappVpnConnectionProtocolUdp
	case config.Protocol_TCP:
		protocol = moose.NordvpnappVpnConnectionProtocolTcp
	case config.Protocol_UNKNOWN_PROTOCOL:
		fallthrough
	default:
		protocol = moose.NordvpnappVpnConnectionProtocolRecommended
	}
	if err := s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigCurrentStateProtocolValue(protocol)); err != nil {
		return err
	}
	return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesProtocolValue(protocol))
}

func (s *Subscriber) NotifyAllowlist(data events.DataAllowlist) error {
	enabled := len(data.UDPPorts) != 0 || len(data.TCPPorts) != 0 || len(data.Subnets) != 0
	if err := s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesSplitTunnelingEnabledMeta(
		fmt.Sprintf(`{"udp_ports":%d,"tcp_ports:%d,"subnets":%d}`, len(data.UDPPorts), len(data.TCPPorts), len(data.Subnets)),
	)); err != nil {
		return err
	}
	return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesSplitTunnelingEnabledValue(enabled))
}

func (s *Subscriber) NotifyTechnology(data config.Technology) error {
	var technology moose.NordvpnappVpnConnectionTechnology
	switch data {
	case config.Technology_NORDLYNX:
		technology = moose.NordvpnappVpnConnectionTechnologyNordlynx
	case config.Technology_OPENVPN:
		technology = moose.NordvpnappVpnConnectionTechnologyOpenvpn
	case config.Technology_UNKNOWN_TECHNOLOGY:
		return errors.New("unknown technology")
	default:
		technology = moose.NordvpnappVpnConnectionTechnologyRecommended
	}
	if err := s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigCurrentStateTechnologyValue(technology)); err != nil {
		return err
	}
	return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesTechnologyValue(technology))
}

func (s *Subscriber) NotifyConnect(data events.DataConnect) error {
	if data.IsMeshnetPeer {
		return nil
	}

	var threatProtection moose.NordvpnappOptBool
	if data.ThreatProtectionLite {
		threatProtection = moose.NordvpnappOptBoolTrue
	} else {
		threatProtection = moose.NordvpnappOptBoolFalse
	}

	eventDurationMs := -1
	var eventStatus moose.NordvpnappEventStatus
	switch data.Type {
	case events.ConnectAttempt:
		eventStatus = moose.NordvpnappEventStatusAttempt
		s.connectionStartTime = time.Now()
	case events.ConnectSuccess:
		eventStatus = moose.NordvpnappEventStatusSuccess
		eventDurationMs = int(time.Since(s.connectionStartTime).Milliseconds())
	case events.ConnectFailure:
		eventStatus = moose.NordvpnappEventStatusFailureDueToRuntimeException
		eventDurationMs = int(time.Since(s.connectionStartTime).Milliseconds())
	default:
		eventStatus = moose.NordvpnappEventStatusAttempt
	}

	var protocol moose.NordvpnappVpnConnectionProtocol
	switch data.Protocol {
	case config.Protocol_TCP:
		protocol = moose.NordvpnappVpnConnectionProtocolTcp
	case config.Protocol_UDP:
		protocol = moose.NordvpnappVpnConnectionProtocolUdp
	case config.Protocol_UNKNOWN_PROTOCOL:
		protocol = moose.NordvpnappVpnConnectionProtocolNone
	default:
		protocol = moose.NordvpnappVpnConnectionProtocolRecommended
	}

	var technology moose.NordvpnappVpnConnectionTechnology
	switch data.Technology {
	case config.Technology_OPENVPN:
		technology = moose.NordvpnappVpnConnectionTechnologyOpenvpn
	case config.Technology_NORDLYNX:
		technology = moose.NordvpnappVpnConnectionTechnologyNordlynx
	case config.Technology_UNKNOWN_TECHNOLOGY:
		technology = moose.NordvpnappVpnConnectionTechnologyNone
	default:
		technology = moose.NordvpnappVpnConnectionTechnologyRecommended
	}

	var server moose.NordvpnappServerListSource
	if data.ServerFromAPI {
		server = moose.NordvpnappServerListSourceRecommendedByApi
	} else {
		server = moose.NordvpnappServerListSourceLocallyCachedServerList
	}

	var rule moose.NordvpnappServerSelectionRule
	switch data.TargetServerSelection {
	default:
		rule = moose.NordvpnappServerSelectionRuleRecommended
	}
	return s.response(moose.NordvpnappSendServiceQualityServersConnect(
		int32(eventDurationMs), // milliseconds
		eventStatus,
		moose.NordvpnappEventTriggerUser,
		moose.NordvpnappVpnConnectionTriggerNone,
		moose.NordvpnappVpnConnectionPresetNone,
		rule,
		server,
		data.TargetServerGroup,
		data.TargetServerDomain,
		data.TargetServerIP,
		data.TargetServerCountry,
		data.TargetServerCity,
		protocol,
		technology,
		threatProtection,
		int32(-1),
		"",
		int32(-1),
	))
}

func (s *Subscriber) NotifyDisconnect(data events.DataDisconnect) error {
	event := moose.NordvpnappEventStatusAttempt
	switch data.Type {
	case events.DisconnectAttempt:
		event = moose.NordvpnappEventStatusAttempt
	case events.DisconnectSuccess:
		event = moose.NordvpnappEventStatusSuccess
	case events.DisconnectFailure:
		event = moose.NordvpnappEventStatusFailureDueToRuntimeException
	}

	var technology moose.NordvpnappVpnConnectionTechnology
	switch data.Technology {
	case config.Technology_OPENVPN:
		technology = moose.NordvpnappVpnConnectionTechnologyOpenvpn
	case config.Technology_NORDLYNX:
		technology = moose.NordvpnappVpnConnectionTechnologyNordlynx
	case config.Technology_UNKNOWN_TECHNOLOGY:
		technology = moose.NordvpnappVpnConnectionTechnologyNone
	default:
		technology = moose.NordvpnappVpnConnectionTechnologyRecommended
	}

	var protocol moose.NordvpnappVpnConnectionProtocol
	switch data.Protocol {
	case config.Protocol_TCP:
		protocol = moose.NordvpnappVpnConnectionProtocolTcp
	case config.Protocol_UDP:
		protocol = moose.NordvpnappVpnConnectionProtocolUdp
	case config.Protocol_UNKNOWN_PROTOCOL:
		protocol = moose.NordvpnappVpnConnectionProtocolNone
	default:
		protocol = moose.NordvpnappVpnConnectionProtocolRecommended
	}

	var server moose.NordvpnappServerListSource
	if data.ServerFromAPI {
		server = moose.NordvpnappServerListSourceRecommendedByApi
	} else {
		server = moose.NordvpnappServerListSourceLocallyCachedServerList
	}

	var rule moose.NordvpnappServerSelectionRule
	switch data.TargetServerSelection {
	default:
		rule = moose.NordvpnappServerSelectionRuleRecommended
	}

	var threatProtection moose.NordvpnappOptBool
	if data.ThreatProtectionLite {
		threatProtection = moose.NordvpnappOptBoolTrue
	} else {
		threatProtection = moose.NordvpnappOptBoolFalse
	}
	return s.response(moose.NordvpnappSendServiceQualityServersDisconnect(
		int32(-1),
		event,
		moose.NordvpnappEventTriggerUser,
		moose.NordvpnappVpnConnectionTriggerNone, // pass proper trigger
		moose.NordvpnappVpnConnectionPresetNone,
		rule,
		server,
		"",
		"",
		"",
		"",
		"",
		protocol,
		technology,
		threatProtection,
		int32(time.Since(s.connectionStartTime).Seconds()), // seconds
		"",
		int32(-1),
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

	var eventStatus moose.NordvpnappEventStatus
	if data.Error != nil {
		eventStatus = moose.NordvpnappEventStatusSuccess
	} else {
		eventStatus = moose.NordvpnappEventStatusFailureDueToRuntimeException
	}
	return s.response(fn(
		int32(data.Duration.Milliseconds()),
		eventStatus,
		moose.NordvpnappEventTriggerApp,
		data.Request.URL.Host,
		int32(responseCode),
		data.Request.Proto,
		0,
		"",
		"",
		"",
		"",
		"",
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

func DrainStart(dbPath string) uint {
	return worker.Start(
		dbPath,
		workerVersion,
		"http://localhost",
		100,
		1000,
		false,
		20,
		false,
	)
}
