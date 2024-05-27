package daemon

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
	"github.com/godbus/dbus/v5"

	"google.golang.org/grpc/metadata"
)

func (r *RPC) StartJobs() {
	// order of the jobs below matters
	// servers job requires geo info and configs data to create server list
	// TODO what if configs file is deleted just before servers job or disk is full?
	if _, err := r.scheduler.Every(6).Hours().Do(JobCountries(r.dm, r.api)); err != nil {
		log.Println(internal.WarningPrefix, "job countries", err)
	}

	if _, err := r.scheduler.Every(30).Minutes().Do(JobInsights(r.dm, r.api, r.netw, false)); err != nil {
		log.Println(internal.WarningPrefix, "job insights", err)
	}

	if _, err := r.scheduler.Every(1).Hour().Do(JobServers(r.dm, r.cm, r.api, true)); err != nil {
		log.Println(internal.WarningPrefix, "job servers", err)
	}
	// TODO if autoconnect runs before servers job, it will return zero servers list

	if _, err := r.scheduler.Every(15).Minutes().Do(JobServerCheck(r.dm, r.api, r.netw, r.lastServer)); err != nil {
		log.Println(internal.WarningPrefix, "job servers", err)
	}

	if _, err := r.scheduler.Every(1).Day().Do(JobTemplates(r.cdn)); err != nil {
		log.Println(internal.WarningPrefix, "job templates", err)
	}

	if _, err := r.scheduler.Every(3).Hours().Do(JobVersionCheck(r.dm, r.repo)); err != nil {
		log.Println(internal.WarningPrefix, "job version", err)
	}

	if _, err := r.scheduler.Every(1).Day().Do(JobHeartBeat(1*24*60 /*minutes*/, r.events)); err != nil {
		log.Println(internal.WarningPrefix, "job heart beat", err)
	}

	r.scheduler.RunAll()
	r.scheduler.StartBlocking()
}

func (r *RPC) StartKillSwitch() {
	var cfg config.Config
	err := r.cm.Load(&cfg)
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
		return
	}

	if cfg.KillSwitch {
		allowlist := cfg.AutoConnectData.Allowlist
		if cfg.LanDiscovery {
			allowlist = addLANPermissions(allowlist)
		}
		if err := r.netw.SetKillSwitch(allowlist); err != nil {
			log.Println(internal.ErrorPrefix, "starting killswitch:", err)
			return
		}
		return
	}
}

func (r *RPC) StopKillSwitch() error {
	var cfg config.Config
	err := r.cm.Load(&cfg)
	if err != nil {
		return fmt.Errorf("loading daemon config: %w", err)
	}

	if cfg.KillSwitch {
		// do not unset killswitch rules if system is in shutdown or reboot
		systemd := internal.IsSystemd()
		shutdownIsActive := (systemd && r.systemShutdown.Load()) ||
			(!systemd && internal.IsSystemShutdown())
		if shutdownIsActive {
			log.Println(internal.InfoPrefix, "detected system reboot - do not remove killswitch protection.")
			return nil
		}
		if err := r.netw.UnsetKillSwitch(); err != nil {
			return fmt.Errorf("unsetting killswitch: %w", err)
		}
	}
	return nil
}

// StartSystemShutdownMonitor to be run on separate goroutine
func (r *RPC) StartSystemShutdownMonitor() {
	// get connection to system dbus
	conn, err := dbus.SystemBus()
	if err != nil {
		log.Println(internal.ErrorPrefix, "getting system dbus:", err)
		return
	}
	defer conn.Close()

	// register dbus signal monitor
	err = conn.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.systemd1.Manager"),
		dbus.WithMatchObjectPath("/org/freedesktop/systemd1"),
		dbus.WithMatchMember("JobNew"),
	)
	if err != nil {
		log.Println(internal.ErrorPrefix, "registering dbus signal monitor:", err)
		return
	}

	/* expected signal example:
	signal time=1716379735.997938 sender=:1.3 -> destination=(null destination) serial=610 path=/org/freedesktop/systemd1; interface=org.freedesktop.systemd1.Manager; member=JobNew
	   uint32 1541
	   object path "/org/freedesktop/systemd1/job/1541"
	   string "reboot.target"
	*/

	log.Println(internal.InfoPrefix, "dbus monitor started, waiting for signals...")

	dbusSignalCh := make(chan *dbus.Signal, 1)
	conn.Signal(dbusSignalCh)
	for signal := range dbusSignalCh {
		if isSystemShutdownSignal(signal) {
			log.Println(internal.InfoPrefix, "got dbus signal - shutdown detected!")
			r.systemShutdown.Store(true)
			return
		}
	}
}

func isSystemShutdownSignal(sig *dbus.Signal) bool {
	if sig.Name == "org.freedesktop.systemd1.Manager.JobNew" {
		for _, bodyItem := range sig.Body {
			str, ok := bodyItem.(string)
			if !ok { // skip non string items
				continue
			}
			switch strings.ToLower(str) {
			case "poweroff.target", "halt.target", "reboot.target":
				return true
			}
		}
	}
	return false
}

type autoconnectServer struct {
	err error
}

func (autoconnectServer) SetHeader(metadata.MD) error  { return nil }
func (autoconnectServer) SendHeader(metadata.MD) error { return nil }
func (autoconnectServer) SetTrailer(metadata.MD)       {}
func (autoconnectServer) Context() context.Context     { return nil }
func (autoconnectServer) SendMsg(m interface{}) error  { return nil }
func (autoconnectServer) RecvMsg(m interface{}) error  { return nil }
func (a *autoconnectServer) Send(data *pb.Payload) error {
	switch data.GetType() {
	case internal.CodeFailure:
		a.err = errors.New("autoconnect failure")
	}
	return nil
}

type GetTimeoutFunc func(tries int) time.Duration

func connectErrorCheck(err error) bool {
	return err == nil ||
		errors.Is(err, internal.ErrNotLoggedIn)
}

// StartAutoConnect connect to VPN server if autoconnect is enabled
func (r *RPC) StartAutoConnect(timeoutFn GetTimeoutFunc) error {
	tries := 1
	for {
		if r.netw.IsVPNActive() {
			log.Println(internal.InfoPrefix, "auto-connect success (already connected)")
			return nil
		}

		var cfg config.Config
		err := r.cm.Load(&cfg)
		if err != nil {
			log.Println(internal.ErrorPrefix, "auto-connect failed with error:", err)
			return err
		}

		server := autoconnectServer{}
		err = r.Connect(&pb.ConnectRequest{ServerTag: cfg.AutoConnectData.ServerTag}, &server)
		if connectErrorCheck(err) && server.err == nil {
			log.Println(internal.InfoPrefix, "auto-connect success")
			return nil
		}
		log.Println(internal.ErrorPrefix, "err1:", server.err, "| err2:", err)
		tryAfterDuration := timeoutFn(tries)
		tries++
		log.Println(internal.WarningPrefix, "will retry(", tries, ") auto-connect after:", tryAfterDuration)
		<-time.After(tryAfterDuration)
	}
}

func meshErrorCheck(err error) bool {
	return err == nil ||
		errors.Is(err, meshnet.ErrNotLoggedIn) ||
		errors.Is(err, meshnet.ErrConfigLoad) ||
		errors.Is(err, meshnet.ErrMeshnetNotEnabled)
}

// StartAutoMeshnet enable meshnet if it was enabled before
func (r *RPC) StartAutoMeshnet(meshService *meshnet.Server, timeoutFn GetTimeoutFunc) error {
	tries := 1
	for {
		if r.netw.IsMeshnetActive() {
			log.Println(internal.InfoPrefix, "auto-enable mesh success (already enabled)")
			return nil
		}

		err := meshService.StartMeshnet()
		if meshErrorCheck(err) {
			if err != nil {
				log.Println(internal.ErrorPrefix, "auto-enable mesh failed with error:", err)
				return err
			} else {
				log.Println(internal.InfoPrefix, "auto-enable mesh success")
				return nil
			}
		}
		log.Println(internal.ErrorPrefix, "err1:", err)
		tryAfterDuration := timeoutFn(tries)
		tries++
		log.Println(internal.WarningPrefix, "will retry(", tries, ") enable mesh after:", tryAfterDuration)
		<-time.After(tryAfterDuration)
	}
}
