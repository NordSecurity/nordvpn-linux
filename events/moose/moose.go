//go:build moose

// Package moose provides convenient wrappers for event sending.
package moose

// #cgo amd64 LDFLAGS: -L${SRCDIR}/../../bin/deps/lib/libmoose-nordvpnapp/current/amd64
// #cgo amd64 LDFLAGS: -L${SRCDIR}/../../bin/deps/lib/libmoose-worker/current/amd64
// #cgo 386 LDFLAGS: -L${SRCDIR}/../../bin/deps/lib/libmoose-nordvpnapp/current/i386
// #cgo 386 LDFLAGS: -L${SRCDIR}/../../bin/deps/lib/libmoose-worker/current/i386
// #cgo arm LDFLAGS: -L${SRCDIR}/../../bin/deps/lib/libmoose-nordvpnapp/current/armel
// #cgo arm LDFLAGS: -L${SRCDIR}/../../bin/deps/lib/libmoose-worker/current/armel
// #cgo arm LDFLAGS: -L${SRCDIR}/../../bin/deps/lib/libmoose-nordvpnapp/current/armhf
// #cgo arm LDFLAGS: -L${SRCDIR}/../../bin/deps/lib/libmoose-worker/current/armhf
// #cgo arm64 LDFLAGS: -L${SRCDIR}/../../bin/deps/lib/libmoose-nordvpnapp/current/aarch64
// #cgo arm64 LDFLAGS: -L${SRCDIR}/../../bin/deps/lib/libmoose-worker/current/aarch64
// #cgo LDFLAGS: -ldl -lm -lmoosenordvpnapp -lmooseworker -lsqlite3
import "C"

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	telemetrypb "github.com/NordSecurity/nordvpn-linux/daemon/pb/telemetry/v1"
	"github.com/NordSecurity/nordvpn-linux/daemon/telemetry"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/snapconf"
	"github.com/NordSecurity/nordvpn-linux/sysinfo"

	moose "moose/events"
	worker "moose/worker"
)

type mooseConsentFunc func(bool) uint32
type mooseSetConsentIntoContextFunc func(moose.NordvpnappConsentLevel) uint32
type mooseSetTokenRenewDateFunc func(int32) uint32
type mooseUnsetTokenRenewDateFunc func() uint32

// Subscriber listen events, send to moose engine
type Subscriber struct {
	eventsDbPath                 string
	config                       config.Manager
	buildTarget                  config.BuildTarget
	domain                       string
	subdomain                    string
	deviceID                     string
	clientAPI                    core.ClientAPI
	currentDomain                string
	connectionStartTime          time.Time
	connectionToMeshnetPeer      bool
	initialHeartbeatSent         bool
	mooseConsentLevelFunc        mooseConsentFunc
	mooseSetConsentIntoCtxFunc   mooseSetConsentIntoContextFunc
	mooseSetTokenRenewDateFunc   mooseSetTokenRenewDateFunc
	mooseUnsetTokenRenewDateFunc mooseUnsetTokenRenewDateFunc
	httpClient                   *http.Client
	canSendAllEvents             atomic.Bool
	mux                          sync.RWMutex
	isInitialized                bool
}

func NewSubscriber(
	eventsDbPath string,
	fs *config.FilesystemConfigManager,
	clientAPI core.ClientAPI,
	httpClient *http.Client,
	buildTarget config.BuildTarget,
	id string,
	eventsDomain string,
	eventsSubdomain string) *Subscriber {

	sub := &Subscriber{
		eventsDbPath:  eventsDbPath,
		config:        fs,
		buildTarget:   buildTarget,
		domain:        eventsDomain,
		subdomain:     eventsSubdomain,
		deviceID:      id,
		clientAPI:     clientAPI,
		httpClient:    httpClient,
		isInitialized: false,
	}
	return sub
}

// getConfig fetches always the current config via config.Manager
func (s *Subscriber) getConfig() (config.Config, error) {
	var cfg config.Config
	err := s.config.Load(&cfg)
	return cfg, err
}

// changeConsentState takes new consent state, compares it to the current one in the config
// and after running some sanity checks, updates moose accordingly (enable or switch analytics on demand)
func (s *Subscriber) changeConsentState(newState config.AnalyticsConsent) error {
	cfg, err := s.getConfig()
	if err != nil {
		return err
	}

	// the same state requested, no-op
	if cfg.AnalyticsConsent == newState {
		return nil
	}

	enabled := newState == config.ConsentGranted
	log.Println(internal.InfoPrefix, LogComponentPrefix, "request to set consent level to", enabled)
	if err := s.response(s.mooseConsentLevelFunc(enabled)); err != nil {
		return fmt.Errorf("setting new consent level: %w", err)
	}

	log.Println(internal.InfoPrefix, LogComponentPrefix, "update consent level into context with new value", newState.String())
	if err := setUserConsentLevelIntoContext(s, newState); err != nil {
		return err
	}
	s.canSendAllEvents.Store(newState == config.ConsentGranted)

	return nil
}

func setUserConsentLevelIntoContext(s *Subscriber, consent config.AnalyticsConsent) error {
	if consent == config.ConsentUndefined {
		return nil
	}
	consentLevel := moose.NordvpnappConsentLevelEssential
	if consent == config.ConsentGranted {
		consentLevel = moose.NordvpnappConsentLevelAnalytics
	}
	if err := s.response(s.mooseSetConsentIntoCtxFunc(consentLevel)); err != nil {
		return fmt.Errorf("setting user consent level: %w", err)
	}
	return nil
}

// Enable moose analytics engine
func (s *Subscriber) Enable() error {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.changeConsentState(config.ConsentGranted)
}

// Disable moose analytics engine
func (s *Subscriber) Disable() error {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.changeConsentState(config.ConsentDenied)
}

func (s *Subscriber) isEnabled() bool {
	return s.canSendAllEvents.Load()
}

// Init initializes moose libs. It has to be done before usage regardless of the enabled state.
// Disabled case should be handled by `set_opt_out` value.
func (s *Subscriber) Init(consent config.AnalyticsConsent) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.isInitialized {
		return nil
	}
	log.Println(internal.InfoPrefix, LogComponentPrefix, "initializing")

	s.mooseConsentLevelFunc = moose.MooseNordvpnappSetConsentLevel
	s.mooseSetConsentIntoCtxFunc = moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesConsentLevel
	s.mooseSetTokenRenewDateFunc = moose.NordvpnappSetContextApplicationNordvpnappConfigCurrentStateTokenRenewDateValue
	s.mooseUnsetTokenRenewDateFunc = moose.NordvpnappUnsetContextApplicationNordvpnappConfigCurrentStateTokenRenewDateValue

	cfg, err := s.getConfig()
	if err != nil {
		return err
	}

	err = s.updateEventDomain()
	if err != nil {
		return fmt.Errorf("initializing event domain: %w", err)
	}

	singleInterval := time.Second
	sequenceInterval := time.Second * 5
	var batchSize uint32 = 20
	compressRequest := true

	// Moose is now started only when the user has specified analytics consent,
	// so we can safely enable events from the very beginning.

	const allowEventsFromBeginning = true
	client := worker.NewHttpClientContext(s.currentDomain)
	client.Client = *s.httpClient
	if err := s.response(uint32(worker.StartWithClient(
		s.eventsDbPath,
		s.currentDomain,
		uint64(singleInterval.Milliseconds()),
		uint64(sequenceInterval.Milliseconds()),
		allowEventsFromBeginning,
		batchSize,
		compressRequest,
		&client,
	))); err != nil {
		return fmt.Errorf("starting worker: %w", err)
	}

	// can we send only essential or all?
	s.canSendAllEvents.Store(consent == config.ConsentGranted)
	log.Println(internal.InfoPrefix, LogComponentPrefix, "all events are sent:", s.canSendAllEvents.Load())

	if err := s.response(moose.MooseNordvpnappInit(
		s.eventsDbPath,
		internal.IsProdEnv(s.buildTarget.Environment),
		s,
		s,
		s.canSendAllEvents.Load(),
	)); err != nil {
		if !strings.Contains(err.Error(), "moose: already initiated") {
			return fmt.Errorf("starting tracker: %w", err)
		}
	}

	// TODO (LVPN-9654): currently, it should be safe to assume moose got correctly initialized when both worker and the app got started properly
	// however this mechanism of initialization might need to be revisited in the future
	s.isInitialized = true

	if err := s.response(moose.MooseNordvpnappFlushChanges()); err != nil {
		log.Println(internal.WarningPrefix, LogComponentPrefix, "failed to flush changes before setting analytics opt in: %w", err)
	}

	applicationName := "linux-app"
	if snapconf.IsUnderSnap() {
		applicationName = "linux-app-snap"
	}

	if err := s.response(moose.NordvpnappSetContextApplicationNordvpnappName(applicationName)); err != nil {
		return fmt.Errorf("setting application name: %w", err)
	}

	if err := setUserConsentLevelIntoContext(s, consent); err != nil {
		return err
	}

	if err := s.response(moose.NordvpnappSetContextApplicationNordvpnappVersion(s.buildTarget.Version)); err != nil {
		return fmt.Errorf("setting application version: %w", err)
	}

	if err := s.response(moose.NordvpnappSetContextDeviceTimeZone(internal.Timezone())); err != nil {
		return fmt.Errorf("setting moose time zone: %w", err)
	}

	distroVersion, err := sysinfo.GetHostOSPrettyName()
	if err != nil {
		return fmt.Errorf("determining device os 'pretty-name'")
	}
	if err := s.response(moose.NordvpnappSetContextDeviceOs(distroVersion)); err != nil {
		return fmt.Errorf("setting moose device os: %w", err)
	}
	if err := s.response(moose.NordvpnappSetContextDeviceFp(s.deviceID)); err != nil {
		return fmt.Errorf("setting moose device: %w", err)
	}

	dt := deviceTypeToInternalType(sysinfo.GetDeviceType())
	if err := s.response(moose.NordvpnappSetContextDeviceType(dt)); err != nil {
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

	if err := s.response(moose.NordvpnappSetContextDeviceCpuArchitecture(s.buildTarget.Architecture)); err != nil {
		return fmt.Errorf("setting device architecture: %w", err)
	}

	if err := s.handleTokenRenewDateChange(nil, &cfg); err != nil {
		log.Println(internal.WarningPrefix, LogComponentPrefix, "failed to restore token renew date:", err)
	}

	return nil
}

func (s *Subscriber) Stop() error {
	log.Println(internal.DebugPrefix, LogComponentPrefix, "flushing changes")
	if err := s.response(moose.MooseNordvpnappFlushChanges()); err != nil {
		return fmt.Errorf("flushing changes: %w", err)
	}

	if err := s.response(worker.Stop()); err != nil {
		return fmt.Errorf("stopping moose worker: %w", err)
	}

	log.Println(internal.DebugPrefix, LogComponentPrefix, "deinitializing")
	if err := s.response(moose.MooseNordvpnappDeinit()); err != nil {
		return fmt.Errorf("deinitializing: %w", err)
	}

	return nil
}

func (s *Subscriber) NotifyKillswitch(data bool) error {
	return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesKillSwitchEnabledValue(data))
}

func (s *Subscriber) NotifyAccountCheck(any) error {
	return errors.Join(s.fetchSubscriptions(), s.fetchAndSetVpnServiceExpiration())
}

func (s *Subscriber) NotifyAutoconnect(data bool) error {
	return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesAutoConnectEnabledValue(data))
}

func (s *Subscriber) NotifyDefaults(any) error {
	return s.clearSubscriptions()
}

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

func (s *Subscriber) NotifyLANDiscovery(data bool) error {
	return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesLocalNetworkDiscoveryAllowedValue(data))
}

func (s *Subscriber) NotifyVirtualLocation(data bool) error {
	return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesVirtualServerEnabledValue(data))
}

func (s *Subscriber) NotifyPostquantumVpn(data bool) error {
	return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesPostQuantumEnabledValue(data))
}

func (s *Subscriber) NotifyIpv6(data bool) error {
	if err := s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigCurrentStateIpv6EnabledValue(data)); err != nil {
		return err
	}
	return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesIpv6EnabledValue(data))
}

func (s *Subscriber) NotifyLogin(data events.DataAuthorization) error { // regular login, or login after signup
	mooseFn := moose.NordvpnappSendServiceQualityAuthorizationLogin
	if data.EventType == events.LoginSignUp {
		mooseFn = moose.NordvpnappSendServiceQualityAuthorizationRegister
	}

	loginFlowAltered := moose.NordvpnappOptBoolNone
	if data.EventStatus != events.StatusAttempt {
		loginFlowAltered = moose.NordvpnappOptBoolFalse
		if data.IsAlteredFlowOnNordAccount {
			loginFlowAltered = moose.NordvpnappOptBoolTrue
		}
	}

	if err := s.response(mooseFn(
		moose.EventParams{
			EventDuration: int32(data.DurationMs),
			EventStatus:   eventStatusToInternalType(data.EventStatus),
			EventTrigger:  eventTriggerDomainToInternalType(data.EventTrigger),
		},
		loginFlowAltered,
		int32(data.Reason),
		nil,
	)); err != nil {
		return err
	}

	if data.EventStatus == events.StatusSuccess {
		return errors.Join(s.fetchSubscriptions(), s.fetchAndSetVpnServiceExpiration())
	}
	return nil
}

func (s *Subscriber) NotifyLogout(data events.DataAuthorization) error {
	if err := s.response(moose.NordvpnappSendServiceQualityAuthorizationLogout(
		moose.EventParams{
			EventDuration: int32(data.DurationMs),
			EventStatus:   eventStatusToInternalType(data.EventStatus),
			EventTrigger:  eventTriggerDomainToInternalType(data.EventTrigger),
		},
		int32(data.Reason),
		nil,
	)); err != nil {
		return err
	}

	if data.EventStatus == events.StatusSuccess {
		if err := s.unsetTokenRenewDate(); err != nil {
			return err
		}
		return s.clearSubscriptions()
	}
	return nil
}

func (s *Subscriber) NotifyMFA(data bool) error {
	return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesMfaEnabledValue(data))
}

// configChangeHandler defines a handler for a specific config field change.
type configChangeHandler func(prev, curr *config.Config) error

// OnConfigChanged handles config change events and updates moose context accordingly.
// It dispatches to specific handlers based on which config fields have changed.
func (s *Subscriber) OnConfigChanged(e config.DataConfigChange) error {
	if e.Config == nil {
		return nil
	}

	// Add more handlers here as needed
	handlers := []configChangeHandler{
		s.handleTokenRenewDateChange,
	}

	var errs []error
	for _, handler := range handlers {
		if err := handler(e.PreviousConfig, e.Config); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// handleTokenRenewDateChange updates moose context when token renewal date changes.
func (s *Subscriber) handleTokenRenewDateChange(prev, curr *config.Config) error {
	currentDate := getTokenRenewDate(curr)
	previousDate := getTokenRenewDate(prev)

	if currentDate == "" || currentDate == previousDate {
		return nil
	}

	renewDate, err := time.Parse(internal.ServerDateFormat, currentDate)
	if err != nil {
		log.Println(internal.WarningPrefix, LogComponentPrefix,
			fmt.Sprintf("failed to parse token renew date %q: %v", currentDate, err))
		return nil
	}

	return s.setTokenRenewDate(renewDate.Unix())
}

// getTokenRenewDate extracts the token renewal date for the current user from config.
// Returns empty string if config is nil or token data is not available.
func getTokenRenewDate(cfg *config.Config) string {
	if cfg == nil || cfg.TokensData == nil {
		return ""
	}

	tokenData, ok := cfg.TokensData[cfg.AutoConnectData.ID]
	if !ok {
		return ""
	}

	return tokenData.TokenRenewDate
}

// setTokenRenewDate sets the token renewal date in moose context
func (s *Subscriber) setTokenRenewDate(unixTimestamp int64) error {
	return s.response(s.mooseSetTokenRenewDateFunc(int32(unixTimestamp)))
}

func (s *Subscriber) unsetTokenRenewDate() error {
	return s.response(s.mooseUnsetTokenRenewDateFunc())
}

func (s *Subscriber) NotifyUiItemsClick(data events.UiItemsAction) error {
	itemType := moose.NordvpnappUserInterfaceItemTypeButton
	if data.ItemType == "textbox" {
		itemType = moose.NordvpnappUserInterfaceItemTypeTextBox
	}
	return s.response(moose.NordvpnappSendUserInterfaceUiItemsClick(
		moose.UiItemsParams{
			FormReference: data.FormReference,
			ItemName:      data.ItemName,
			ItemType:      itemType,
			ItemValue:     data.ItemValue,
		},
		nil,
	))
}

func (s *Subscriber) NotifyHeartBeat(period time.Duration) error {
	if err := s.response(moose.NordvpnappSendServiceQualityStatusHeartbeat(int32(period.Minutes()), nil)); err != nil {
		return err
	}
	if !s.initialHeartbeatSent {
		s.mux.Lock()
		defer s.mux.Unlock()
		s.initialHeartbeatSent = true
	}
	return nil
}

func (s *Subscriber) NotifyDeviceLocation(insights core.Insights) error {
	if err := s.response(moose.NordvpnappSetContextDeviceLocationCity(insights.City)); err != nil {
		return fmt.Errorf("setting moose device location city: %w", err)
	}
	if err := s.response(moose.NordvpnappSetContextDeviceLocationCountry(insights.CountryCode)); err != nil {
		return fmt.Errorf("setting moose device location country: %w", err)
	}
	if err := s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigCurrentStateIspValue(insights.Isp)); err != nil {
		return fmt.Errorf("setting moose ISP value: %w", err)
	}
	if err := s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigCurrentStateIspAsnValue(strconv.Itoa(insights.IspAsn))); err != nil {
		return fmt.Errorf("setting moose ISP ASN value: %w", err)
	}
	return nil
}

func (s *Subscriber) NotifyNotify(bool) error { return nil }

func (s *Subscriber) NotifyMeshnet(data bool) error {
	if err := s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesMeshnetEnabledValue(data)); err != nil {
		return err
	}
	if s.initialHeartbeatSent {
		// 0 duration indicates that this is not a periodic heart beat
		return s.NotifyHeartBeat(time.Duration(0))
	}
	return nil
}

func (s *Subscriber) NotifyObfuscate(data bool) error {
	return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesObfuscationEnabledValue(data))
}

func (s *Subscriber) NotifyPeerUpdate([]string) error { return nil }

func (s *Subscriber) NotifySelfRemoved(any) error { return nil }

func (s *Subscriber) NotifyThreatProtectionLite(data bool) error {
	if s.connectionStartTime.IsZero() {
		if err := s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigCurrentStateThreatProtectionLiteEnabledValue(data)); err != nil {
			return err
		}
	}
	return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesThreatProtectionLiteEnabledValue(data))
}

func (s *Subscriber) NotifyProtocol(data config.Protocol) error {
	protocol := connectionProtocolToInternalType(data)
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
	if data == config.Technology_UNKNOWN_TECHNOLOGY {
		return errors.New("unknown technology")
	}

	technology := connectionTechnologyToInternalType(data)
	if err := s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigCurrentStateTechnologyValue(technology)); err != nil {
		return err
	}
	return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigUserPreferencesTechnologyValue(technology))
}

func (s *Subscriber) NotifyConnect(data events.DataConnect) error {
	if data.EventStatus == events.StatusSuccess {
		s.mux.Lock()
		s.connectionStartTime = time.Now()
		s.connectionToMeshnetPeer = data.IsMeshnetPeer
		s.mux.Unlock()
	}

	if data.IsMeshnetPeer {
		return s.response(moose.NordvpnappSendServiceQualityServersConnectToMeshnetDevice(
			moose.EventParams{
				EventDuration: int32(data.DurationMs),
				EventStatus:   eventStatusToInternalType(data.EventStatus),
				EventTrigger:  moose.NordvpnappEventTriggerUser,
			},
			-1,
			-1,
			nil,
		))
	}

	if err := s.response(moose.NordvpnappSendServiceQualityServersConnect(
		moose.EventParams{
			EventDuration: int32(data.DurationMs),
			EventStatus:   eventStatusToInternalType(data.EventStatus),
			EventTrigger:  moose.NordvpnappEventTriggerUser,
		},
		moose.TargetConnectionParams{
			TargetServerListSource:    serverListOriginToInternalType(data.ServerFromAPI),
			TargetServerSelectionRule: serverSelectionRuleToInternalType(data.TargetServerSelection),
			TargetServerType:          moose.NordvpnappServerTypeNone,
		},
		moose.TargetConnectionAdditionalParams{
			TargetProtocol:      connectionProtocolToInternalType(data.Protocol),
			TargetServerCity:    data.TargetServerCity,
			TargetServerCountry: data.TargetServerCountryCode,
			TargetServerDomain:  data.TargetServerDomain,
			TargetServerGroup:   data.TargetServerGroup,
			TargetServerIp:      data.TargetServerIP.String(),
			TargetTechnology:    connectionTechnologyToInternalType(data.Technology),
		},
		moose.ConnectionParams{
			ConnectionFunnel:     "",
			VpnConnectionTrigger: moose.NordvpnappVpnConnectionTriggerNone,
		},
		threatProtectionLiteToInternalType(data.ThreatProtectionLite),
		-1,
		data.RecommendationUUID,
		nil,
	)); err != nil {
		return err
	}

	if data.EventStatus == events.StatusSuccess {
		if err := s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigCurrentStateThreatProtectionLiteEnabledValue(data.ThreatProtectionLite)); err != nil {
			return err
		}

		if err := s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigCurrentStateIsOnVpnValue(true)); err != nil {
			return err
		}

		return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigCurrentStateServerCountryValue(data.TargetServerCountryCode))
	}

	return nil

}

func (s *Subscriber) NotifyDisconnect(data events.DataDisconnect) error {
	s.mux.Lock()
	connectionDuration := int32(time.Since(s.connectionStartTime).Seconds())
	if connectionDuration <= 0 {
		connectionDuration = -1
	}
	s.connectionStartTime = time.Time{}
	s.mux.Unlock()

	if s.connectionToMeshnetPeer {
		return s.response(moose.NordvpnappSendServiceQualityServersDisconnectFromMeshnetDevice(
			moose.EventParams{
				EventDuration: int32(data.Duration.Milliseconds()),
				EventStatus:   eventStatusToInternalType(data.EventStatus),
				EventTrigger:  moose.NordvpnappEventTriggerUser,
			},
			connectionDuration, // seconds
			-1,
			nil,
		))
	}

	if data.RecommendationUUID != "" {
		if err := s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigCurrentStateRecommendationUuid(data.RecommendationUUID)); err != nil {
			// We can ignore setting the recommendation Uuid
			// Sending the disconnect event is much more important
			log.Println(internal.WarningPrefix, "Failed to set RecommendationUUID into the moose context ", err)
		}

		defer func() {
			if err := s.response(moose.NordvpnappUnsetContextApplicationNordvpnappConfigCurrentStateRecommendationUuid()); err != nil {
				log.Println(internal.WarningPrefix, "Failed to unset RecommendationUUID into the moose context ", err)
			}
		}()
	}

	if err := s.response(moose.NordvpnappSendServiceQualityServersDisconnect(
		moose.EventParams{
			EventDuration: int32(data.Duration.Milliseconds()),
			EventStatus:   eventStatusToInternalType(data.EventStatus),
			// App should never disconnect from VPN by itself. It has to receive either
			// user command (logout, set defaults) or be shut down.
			EventTrigger: moose.NordvpnappEventTriggerUser,
		},
		moose.TargetConnectionParams{
			TargetServerListSource:    serverListOriginToInternalType(data.ServerFromAPI),
			TargetServerSelectionRule: serverSelectionRuleToInternalType(data.TargetServerSelection),
			TargetServerType:          moose.NordvpnappServerTypeNone,
		},
		moose.ConnectionParams{
			ConnectionFunnel:     "",
			VpnConnectionTrigger: moose.NordvpnappVpnConnectionTriggerNone, // pass proper trigger
		},
		connectionDuration, // seconds
		errToExceptionCode(data.Error),
		nil,
	)); err != nil {
		return err
	}

	if err := s.response(moose.NordvpnappUnsetContextApplicationNordvpnappConfigCurrentStateThreatProtectionLiteEnabledValue()); err != nil {
		return err
	}

	if err := s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigCurrentStateServerCountryValue(UnavailableEventParameterValue)); err != nil {
		return err
	}

	return s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigCurrentStateIsOnVpnValue(false))
}

func (s *Subscriber) NotifyRequestAPI(data events.DataRequestAPI) error {
	if data.Request == nil {
		return fmt.Errorf("request nil")
	}

	//for attempt events response_code shall be set to 0
	responseCode := 0
	if data.Response != nil {
		responseCode = data.Response.StatusCode
	}

	notifierFunc := pickNotifier(data.Request.URL.Path)

	var eventStatus moose.NordvpnappEventStatus
	if data.Error == nil {
		if data.IsAttempt {
			eventStatus = moose.NordvpnappEventStatusAttempt
		} else {
			eventStatus = moose.NordvpnappEventStatusSuccess
		}
	} else {
		eventStatus = moose.NordvpnappEventStatusFailureDueToRuntimeException
	}

	//for attempt events duration shall be set to 0
	duration := int32(0)
	if !data.IsAttempt {
		duration = int32(data.Duration.Milliseconds())
	}

	return s.response(notifierFunc(
		moose.EventParams{
			EventDuration: duration,
			EventStatus:   eventStatus,
			EventTrigger:  moose.NordvpnappEventTriggerApp,
		},
		moose.ApiRequestParams{
			ApiHostName:       data.Request.URL.Host,
			DnsResolutionTime: 0,
			Limits:            data.Limits,
			Offset:            data.Offset,
			RequestFields:     data.RequestFields,
			RequestFilters:    data.RequestFilters,
			ResponseCode:      int32(responseCode),
			ResponseSummary:   "",
			TransferProtocol:  data.Request.Proto,
		},
		nil,
	))
}

// NotifyDebuggerEvent processes a DebuggerEvent to emit a moose debugger log.
// It allows providing a custom JSON payload and context paths for the event.
// For custom context paths, corresponding values must be of any of the following types: bool, float32, int32, int64, string.
// Unsupported types are discarded.
//
// Parameters:
//   - e: The DebuggerEvent containing JSON data and context paths to process
func (s *Subscriber) NotifyDebuggerEvent(e events.DebuggerEvent) error {
	combinedPaths := append([]string{}, e.GeneralContextPaths...)
	key := moose.MooseNordvpnappGetDeveloperContextKey()
	for _, ctx := range e.KeyBasedContextPaths {
		path := fmt.Sprintf("%s.%s", key, ctx.Path)
		switch v := ctx.Value.(type) {
		case bool:
			moose.MooseNordvpnappSetDeveloperEventContextBool(ctx.Path, v)
			combinedPaths = append(combinedPaths, path)
		case float32:
			moose.MooseNordvpnappSetDeveloperEventContextFloat(ctx.Path, v)
			combinedPaths = append(combinedPaths, path)
		//deliberately omitted uint64
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32:
			val := reflect.ValueOf(v).Int()
			if val > math.MaxInt32 {
				moose.MooseNordvpnappSetDeveloperEventContextLong(ctx.Path, int64(val))
			} else {
				moose.MooseNordvpnappSetDeveloperEventContextInt(ctx.Path, int32(val))
			}
			combinedPaths = append(combinedPaths, path)
		case string:
			moose.MooseNordvpnappSetDeveloperEventContextString(ctx.Path, v)
			combinedPaths = append(combinedPaths, path)
		default:
			log.Printf("%s %s Discarding unsupported type (%T) on path: %s\n", internal.WarningPrefix, LogComponentPrefix, ctx.Value, path)
		}
	}
	return s.response(moose.NordvpnappSendDebuggerLoggingLog(e.JsonData, combinedPaths, nil))
}

func (s *Subscriber) NotifyAppStartTime(duration int64) error {
	if duration > math.MaxInt32 || duration < 0 {
		return fmt.Errorf("app start duration overflow")
	}

	if err := s.response(moose.NordvpnappSendServiceQualityStatusAppStart(int32(duration), moose.NordvpnappEventTriggerApp, nil)); err != nil {
		return fmt.Errorf("setting app start time")
	}

	return nil
}

func (s *Subscriber) OnTelemetry(metric telemetry.Metric, value any) error {
	switch metric {
	case telemetry.MetricDesktopEnvironment:
		if value.(string) == "" {
			if err := s.response(moose.NordvpnappUnsetContextDeviceDesktopEnvironment()); err != nil {
				return fmt.Errorf("unsetting desktop-environment: %w", err)
			}
		} else {
			if err := s.response(moose.NordvpnappSetContextDeviceDesktopEnvironment(value.(string))); err != nil {
				return fmt.Errorf("setting desktop-environment: %w", err)
			}
		}

	case telemetry.MetricDisplayProtocol:
		switch value.(telemetrypb.DisplayProtocol) {
		case telemetrypb.DisplayProtocol_DISPLAY_PROTOCOL_UNSPECIFIED:
			if err := s.response(moose.NordvpnappUnsetContextApplicationNordvpnappConfigCurrentStateDisplayProtocol()); err != nil {
				return fmt.Errorf("unsetting display protocol: %w", err)
			}
		case telemetrypb.DisplayProtocol_DISPLAY_PROTOCOL_WAYLAND:
			if err := s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigCurrentStateDisplayProtocol("wayland")); err != nil {
				return fmt.Errorf("setting display protocol: %w", err)
			}
		case telemetrypb.DisplayProtocol_DISPLAY_PROTOCOL_X11:
			if err := s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigCurrentStateDisplayProtocol("x11")); err != nil {
				return fmt.Errorf("setting display protocol: %w", err)
			}
		case telemetrypb.DisplayProtocol_DISPLAY_PROTOCOL_UNKNOWN:
		default:
			if err := s.response(moose.NordvpnappSetContextApplicationNordvpnappConfigCurrentStateDisplayProtocol("unknown")); err != nil {
				return fmt.Errorf("setting display protocol: %w", err)
			}
		}

	default:
		return fmt.Errorf("unsupported metric received (id=%d)", metric)
	}

	return nil
}

func (s *Subscriber) fetchAndSetVpnServiceExpiration() error {
	services, err := s.clientAPI.Services()
	if err != nil {
		return fmt.Errorf("fetching services: %w", err)
	}

	// Will return in YYYY-MM-DD HH:MM:SS format
	expiresAt, err := auth.FindVpnServiceExpiration(services)
	if err != nil {
		return err
	}

	// We must use YYYY-MM format
	expiry, err := time.Parse(internal.ServerDateFormat, expiresAt)
	if err != nil {
		return fmt.Errorf("invalid date format")
	}

	date := expiry.Format(internal.YearMonthDateFormat)

	if err := s.response(moose.NordvpnappSetContextUserNordvpnappSubscriptionCurrentStateServiceExpiresAt(date)); err != nil {
		return err
	}

	return nil
}

func (s *Subscriber) OnFirstOpen() error {
	return s.response(moose.NordvpnappSendServiceQualityStatusFirstOpenApp(-1, nil))
}

func (s *Subscriber) fetchSubscriptions() error {
	cfg, err := s.getConfig()
	if err != nil {
		return err
	}

	if cfg.AnalyticsConsent == config.ConsentUndefined {
		return nil
	}

	payments, err := s.clientAPI.Payments()
	if err != nil {
		return fmt.Errorf("fetching payments: %w", err)
	}

	orders, err := s.clientAPI.Orders()
	if err != nil {
		return fmt.Errorf("fetching orders: %w", err)
	}

	payment, ok := findPayment(payments)
	if !ok {
		return fmt.Errorf("no valid payments found for the user")
	}

	var orderErr error
	order, ok := findOrder(payment, orders)
	if !ok {
		orderErr = fmt.Errorf("no valid order was found for the payment")
	}

	if err := s.setSubscriptions(
		payment,
		order,
		countFunc(payments, isPaymentValid, 2),
	); err != nil {
		return errors.Join(orderErr, fmt.Errorf("setting subscriptions: %w", err))
	}

	return orderErr
}

func findPayment(payments []core.PaymentResponse) (core.Payment, bool) {
	// Sort by CreatedAt descending
	slices.SortFunc(payments, func(a core.PaymentResponse, b core.PaymentResponse) int {
		return -a.Payment.CreatedAt.Compare(b.Payment.CreatedAt)
	})

	// Find first element matching criteria
	index := slices.IndexFunc(payments, isPaymentValid)
	if index < 0 {
		return core.Payment{}, false
	}

	return payments[index].Payment, true
}

func findOrder(p core.Payment, orders []core.Order) (core.Order, bool) {
	// Find order matching the payment
	if p.Subscription.MerchantID != 25 && p.Subscription.MerchantID != 3 {
		return core.Order{}, false
	}
	index := slices.IndexFunc(orders, func(o core.Order) bool {
		var cmpID int
		switch p.Subscription.MerchantID {
		case 3:
			cmpID = o.ID
		case 25:
			cmpID = o.RemoteID
		}
		return p.Payer.OrderID == cmpID
	})
	if index < 0 {
		return core.Order{}, false
	}

	return orders[index], true
}

func isPaymentValid(pr core.PaymentResponse) bool {
	p := pr.Payment
	return p.Status == "done" ||
		p.Status == "chargeback" ||
		p.Status == "refunded" ||
		p.Status == "partially_refunded" ||
		p.Status == "trial"
}

// countFunc returns a number of elements in slice matching criteria
func countFunc[S ~[]E, E any](s S, f func(E) bool, stopAt int) int {
	count := 0
	for _, e := range s {
		if f(e) {
			count++
		}
		if stopAt >= 0 && count >= stopAt {
			return count
		}
	}
	return count
}

func (s *Subscriber) setSubscriptions(
	payment core.Payment,
	order core.Order,
	validPaymentsCount int,
) error {
	var plan core.Plan
	if len(order.Plans) > 0 {
		plan = order.Plans[0]
	}
	for _, fn := range []func() uint32{
		func() uint32 {
			return moose.NordvpnappSetContextUserNordvpnappSubscriptionCurrentStateActivationDate(payment.CreatedAt.Format(internal.YearMonthDateFormat))
		},
		func() uint32 {
			return moose.NordvpnappSetContextUserNordvpnappSubscriptionCurrentStateFrequencyInterval(payment.Subscription.FrequencyInterval)
		},
		func() uint32 {
			return moose.NordvpnappSetContextUserNordvpnappSubscriptionCurrentStateFrequencyUnit(payment.Subscription.FrequencyUnit)
		},
		func() uint32 {
			return moose.NordvpnappSetContextUserNordvpnappSubscriptionCurrentStateIsActive(order.Status == "active")
		},
		func() uint32 {
			return moose.NordvpnappSetContextUserNordvpnappSubscriptionCurrentStateIsNewCustomer(validPaymentsCount == 1)
		},
		func() uint32 {
			return moose.NordvpnappSetContextUserNordvpnappSubscriptionCurrentStateMerchantId(payment.Subscription.MerchantID)
		},
		func() uint32 {
			return moose.NordvpnappSetContextUserNordvpnappSubscriptionCurrentStatePaymentAmount(fmt.Sprintf("%g", payment.Amount))
		},
		func() uint32 {
			return moose.NordvpnappSetContextUserNordvpnappSubscriptionCurrentStatePaymentCurrency(payment.Currency)
		},
		func() uint32 {
			return moose.NordvpnappSetContextUserNordvpnappSubscriptionCurrentStatePaymentProvider(payment.Provider)
		},
		func() uint32 {
			return moose.NordvpnappSetContextUserNordvpnappSubscriptionCurrentStatePaymentStatus(payment.Status)
		},
		func() uint32 {
			if plan.ID != 0 {
				return moose.NordvpnappSetContextUserNordvpnappSubscriptionCurrentStatePlanId(plan.ID)
			}
			return 0
		},
		func() uint32 {
			if plan.Type != "" {
				return moose.NordvpnappSetContextUserNordvpnappSubscriptionCurrentStatePlanType(plan.Type)
			}
			return 0
		},
		func() uint32 {
			return moose.NordvpnappSetContextUserNordvpnappSubscriptionCurrentStateSubscriptionStatus(payment.Subscription.Status)
		},
	} {
		if err := s.response(fn()); err != nil {
			return err
		}
	}
	return nil
}

func (s *Subscriber) clearSubscriptions() error {
	for _, fn := range []func() uint32{
		func() uint32 {
			return moose.NordvpnappUnsetContextUserNordvpnappSubscriptionCurrentStateActivationDate()
		},
		func() uint32 {
			return moose.NordvpnappUnsetContextUserNordvpnappSubscriptionCurrentStateFrequencyInterval()
		},
		func() uint32 {
			return moose.NordvpnappUnsetContextUserNordvpnappSubscriptionCurrentStateFrequencyUnit()
		},
		func() uint32 {
			return moose.NordvpnappUnsetContextUserNordvpnappSubscriptionCurrentStateIsActive()
		},
		func() uint32 {
			return moose.NordvpnappUnsetContextUserNordvpnappSubscriptionCurrentStateIsNewCustomer()
		},
		func() uint32 {
			return moose.NordvpnappUnsetContextUserNordvpnappSubscriptionCurrentStateMerchantId()
		},
		func() uint32 {
			return moose.NordvpnappUnsetContextUserNordvpnappSubscriptionCurrentStatePaymentAmount()
		},
		func() uint32 {
			return moose.NordvpnappUnsetContextUserNordvpnappSubscriptionCurrentStatePaymentCurrency()
		},
		func() uint32 {
			return moose.NordvpnappUnsetContextUserNordvpnappSubscriptionCurrentStatePaymentProvider()
		},
		func() uint32 {
			return moose.NordvpnappUnsetContextUserNordvpnappSubscriptionCurrentStatePaymentStatus()
		},
		func() uint32 {
			return moose.NordvpnappUnsetContextUserNordvpnappSubscriptionCurrentStatePlanId()
		},
		func() uint32 {
			return moose.NordvpnappUnsetContextUserNordvpnappSubscriptionCurrentStatePlanType()
		},
		func() uint32 {
			return moose.NordvpnappUnsetContextUserNordvpnappSubscriptionCurrentStateSubscriptionStatus()
		},
		func() uint32 {
			return moose.NordvpnappUnsetContextUserNordvpnappSubscriptionCurrentStateServiceExpiresAt()
		},
	} {
		if err := s.response(fn()); err != nil {
			return err
		}
	}

	return nil
}

func (s *Subscriber) updateEventDomain() error {
	domainUrl, err := url.Parse(s.domain)
	if err != nil {
		return err
	}
	// TODO: Remove subdomain handling logic as it brings no value after domain rotation removal
	if s.subdomain != "" {
		domainUrl.Host = s.subdomain + "." + domainUrl.Host
	}
	s.currentDomain = domainUrl.String()
	return nil
}

func DrainStart(dbPath string) uint32 {
	return worker.Start(
		dbPath,
		"http://localhost",
		100,
		1000,
		false,
		20,
		false,
	)
}

func errToExceptionCode(err error) int32 {
	if err == nil {
		return -1
	}
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "config"):
		return 1
	case strings.Contains(errStr, "networker"):
		return 2
	}
	return -1
}

// eventTriggerDomainToInternalType converts the domain-specific event trigger type to the internal
// representation
func eventTriggerDomainToInternalType(trigger events.TypeEventTrigger) moose.NordvpnappEventTrigger {
	switch trigger {
	case events.TriggerUser:
		return moose.NordvpnappEventTriggerUser
	case events.TriggerApp:
		fallthrough
	default:
		return moose.NordvpnappEventTriggerApp
	}
}

// eventStatusToInternalType converts the event status type to the internal representation
func eventStatusToInternalType(status events.TypeEventStatus) moose.NordvpnappEventStatus {
	switch status {
	case events.StatusSuccess:
		return moose.NordvpnappEventStatusSuccess
	case events.StatusFailure:
		return moose.NordvpnappEventStatusFailureDueToRuntimeException
	case events.StatusCanceled:
		return moose.NordvpnappEventStatusFailureDueToUserInterrupt
	case events.StatusAttempt:
		fallthrough
	default:
		return moose.NordvpnappEventStatusAttempt
	}
}

// connectionProtocolToInternalType converts the connection protocol to the internal representation
func connectionProtocolToInternalType(proto config.Protocol) moose.NordvpnappVpnConnectionProtocol {
	switch proto {
	case config.Protocol_TCP:
		return moose.NordvpnappVpnConnectionProtocolTcp
	case config.Protocol_UDP:
		return moose.NordvpnappVpnConnectionProtocolUdp
	case config.Protocol_Webtunnel:
		return moose.NordvpnappVpnConnectionProtocolWebtunnel
	case config.Protocol_UNKNOWN_PROTOCOL:
		return moose.NordvpnappVpnConnectionProtocolNone
	default:
		return moose.NordvpnappVpnConnectionProtocolRecommended
	}
}

// connectionTechnologyToInternalType converts connection technology to the internal representation
func connectionTechnologyToInternalType(tech config.Technology) moose.NordvpnappVpnConnectionTechnology {
	switch tech {
	case config.Technology_OPENVPN:
		return moose.NordvpnappVpnConnectionTechnologyOpenvpn
	case config.Technology_NORDLYNX:
		return moose.NordvpnappVpnConnectionTechnologyNordlynx
	case config.Technology_NORDWHISPER:
		return moose.NordvpnappVpnConnectionTechnologyNordwhisper
	case config.Technology_UNKNOWN_TECHNOLOGY:
		return moose.NordvpnappVpnConnectionTechnologyNone
	default:
		return moose.NordvpnappVpnConnectionTechnologyRecommended
	}
}

// serverListOriginToInternalType converts server list origin to the internal representation
func serverListOriginToInternalType(sourceFromApi bool) moose.NordvpnappServerListSource {
	if sourceFromApi {
		return moose.NordvpnappServerListSourceRecommendedByApi
	}

	return moose.NordvpnappServerListSourceLocallyCachedServerList
}

// serverSelectionRuleToInternalType converts server selection rule to the internal representation
func serverSelectionRuleToInternalType(rule config.ServerSelectionRule) moose.NordvpnappServerSelectionRule {
	switch rule {
	case config.ServerSelectionRule_RECOMMENDED:
		return moose.NordvpnappServerSelectionRuleRecommended
	case config.ServerSelectionRule_CITY:
		return moose.NordvpnappServerSelectionRuleCity
	case config.ServerSelectionRule_COUNTRY:
		return moose.NordvpnappServerSelectionRuleCountry
	case config.ServerSelectionRule_SPECIFIC_SERVER:
		return moose.NordvpnappServerSelectionRuleSpecificServer
	case config.ServerSelectionRule_GROUP:
		return moose.NordvpnappServerSelectionRuleSpecialtyServer
	case config.ServerSelectionRule_COUNTRY_WITH_GROUP:
		return moose.NordvpnappServerSelectionRuleSpecialtyServerWithCountry
	case config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP:
		return moose.NordvpnappServerSelectionRuleSpecialtyServerWithSpecificServer
	default:
		return moose.NordvpnappServerSelectionRuleNone
	}
}

// threatProtectionLiteToInternalType converts thread protection lite to the internal representation
func threatProtectionLiteToInternalType(enabled bool) moose.NordvpnappOptBool {
	if enabled {
		return moose.NordvpnappOptBoolTrue
	}

	return moose.NordvpnappOptBoolFalse

}

func deviceTypeToInternalType(deviceType sysinfo.SystemDeviceType) moose.NordvpnappDeviceType {
	var dt moose.NordvpnappDeviceType

	switch deviceType {
	case sysinfo.SystemDeviceTypeDesktop:
		dt = moose.NordvpnappDeviceTypeDesktop
	case sysinfo.SystemDeviceTypeServer:
		dt = moose.NordvpnappDeviceTypeServer
	case sysinfo.SystemDeviceTypeContainer:
		dt = moose.NordvpnappDeviceTypeVirtualMachine
	default:
		dt = moose.NordvpnappDeviceTypeUndefined
	}

	return dt
}
