package libtelio

import (
	"context"
	"errors"
	"net/netip"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

type Holder struct {
	mu sync.RWMutex

	current *Libtelio

	prod            bool
	eventPath       string
	fwmark          uint32
	vpnLibCfg       vpn.LibConfigGetter
	appVersion      string
	eventsPublisher *vpn.Events
	ensEnabledFn    func() bool
}

func NewHolder(
	prod bool,
	eventPath string,
	fwmark uint32,
	vpnLibCfg vpn.LibConfigGetter,
	appVersion string,
	eventsPublisher *vpn.Events,
	ensEnabledFn func() bool,
) (*Holder, error) {
	h := &Holder{
		prod:            prod,
		eventPath:       eventPath,
		fwmark:          fwmark,
		vpnLibCfg:       vpnLibCfg,
		appVersion:      appVersion,
		eventsPublisher: eventsPublisher,
		ensEnabledFn:    ensEnabledFn,
	}

	inst, err := New(prod, eventPath, fwmark, vpnLibCfg, appVersion, eventsPublisher, ensEnabledFn())
	h.current = inst
	return h, err
}

func (h *Holder) Recreate() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.current != nil {
		if err := h.current.Stop(); err != nil {
			log.Warn("stopping previous instance before recreate:", err)
		}
	}

	inst, err := New(h.prod, h.eventPath, h.fwmark, h.vpnLibCfg, h.appVersion, h.eventsPublisher, h.ensEnabledFn())
	h.current = inst
	if err != nil {
		log.Error("recreating libtelio instance:", err)
		return err
	}

	log.Info("libtelio instance recreated")
	return nil
}

func (h *Holder) instance() (*Libtelio, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.current == nil {
		return nil, errors.New("libtelio instance is not available")
	}
	return h.current, nil
}

func (h *Holder) Start(ctx context.Context, creds vpn.Credentials, serverData vpn.ServerData) error {
	inst, err := h.instance()
	if err != nil {
		return err
	}
	return inst.Start(ctx, creds, serverData)
}

func (h *Holder) Stop() error {
	inst, err := h.instance()
	if err != nil {
		return err
	}
	return inst.Stop()
}

func (h *Holder) IsActive() bool {
	inst, err := h.instance()
	if err != nil {
		return false
	}
	return inst.IsActive()
}

func (h *Holder) State() vpn.State {
	inst, err := h.instance()
	if err != nil {
		return vpn.ExitedState
	}
	return inst.State()
}

func (h *Holder) Tun() tunnel.T {
	inst, err := h.instance()
	if err != nil {
		return &tunnel.Tunnel{}
	}
	return inst.Tun()
}

func (h *Holder) NetworkChanged() error {
	inst, err := h.instance()
	if err != nil {
		return err
	}
	return inst.NetworkChanged()
}

func (h *Holder) GetConnectionParameters() (vpn.ServerData, bool) {
	inst, err := h.instance()
	if err != nil {
		return vpn.ServerData{}, false
	}
	return inst.GetConnectionParameters()
}

func (h *Holder) Enable(ip netip.Addr, privateKey string) error {
	inst, err := h.instance()
	if err != nil {
		return err
	}
	return inst.Enable(ip, privateKey)
}

func (h *Holder) Disable() error {
	inst, err := h.instance()
	if err != nil {
		return err
	}
	return inst.Disable()
}

func (h *Holder) Refresh(c mesh.MachineMap) error {
	inst, err := h.instance()
	if err != nil {
		return err
	}
	return inst.Refresh(c)
}

func (h *Holder) StatusMap() (map[string]string, error) {
	inst, err := h.instance()
	if err != nil {
		return nil, err
	}
	return inst.StatusMap()
}

func (h *Holder) Private() string {
	inst, err := h.instance()
	if err != nil {
		return ""
	}
	return inst.Private()
}

func (h *Holder) Public(private string) string {
	inst, err := h.instance()
	if err != nil {
		return ""
	}
	return inst.Public(private)
}
