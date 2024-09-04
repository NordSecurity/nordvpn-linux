package daemon

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/godbus/dbus/v5"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/state"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/meshnet"

	"google.golang.org/grpc/metadata"
)

func (r *RPC) StartJobs(statePublisher *state.StatePublisher) {
	// order of the jobs below matters
	// servers job requires geo info and configs data to create server list
	// TODO what if configs file is deleted just before servers job or disk is full?
	if _, err := r.scheduler.NewJob(gocron.DurationJob(6*time.Hour), gocron.NewTask(JobCountries(r.dm, r.serversAPI)), gocron.WithName("job countries")); err != nil {
		log.Println(internal.WarningPrefix, "job countries schedule error:", err)
	}

	jobInsights, err := r.scheduler.NewJob(gocron.DurationJob(30*time.Minute), gocron.NewTask(JobInsights(r.dm, r.api, r.netw, r.events, false)), gocron.WithName("job insights"))
	if err != nil {
		log.Println(internal.WarningPrefix, "job insights schedule error:", err)
	}

	if _, err := r.scheduler.NewJob(gocron.DurationJob(1*time.Hour), gocron.NewTask(JobServers(r.dm, r.cm, r.serversAPI, true)), gocron.WithName("job servers")); err != nil {
		log.Println(internal.WarningPrefix, "job servers schedule error:", err)
	}
	// TODO if autoconnect runs before servers job, it will return zero servers list

	if _, err := r.scheduler.NewJob(gocron.DurationJob(15*time.Minute), gocron.NewTask(JobServerCheck(r.dm, r.serversAPI, r.netw, r.lastServer)), gocron.WithName("job servers check")); err != nil {
		log.Println(internal.WarningPrefix, "job servers check schedule error:", err)
	}

	if _, err := r.scheduler.NewJob(gocron.DurationJob(24*time.Hour), gocron.NewTask(JobTemplates(r.cdn)), gocron.WithName("job templates")); err != nil {
		log.Println(internal.WarningPrefix, "job templates schedule error:", err)
	}

	if _, err := r.scheduler.NewJob(gocron.DurationJob(3*time.Hour), gocron.NewTask(JobVersionCheck(r.dm, r.repo)), gocron.WithName("job version")); err != nil {
		log.Println(internal.WarningPrefix, "job version schedule error:", err)
	}

	if _, err := r.scheduler.NewJob(gocron.DurationJob(24*time.Hour), gocron.NewTask(JobHeartBeat(1*24*60 /*minutes*/, r.events)), gocron.WithName("job heart beat")); err != nil {
		log.Println(internal.WarningPrefix, "job heart beat schedule error:", err)
	}

	r.scheduler.Start()
	for _, job := range r.scheduler.Jobs() {
		err := job.RunNow()
		if err != nil {
			log.Println(internal.WarningPrefix, job.Name(), "first run error:", err)
		}
	}

	go func() {
		stateChan, _ := statePublisher.AddSubscriber()
		for ev := range stateChan {
			switch ev.(type) {
			case events.DataConnect:
			case events.DataDisconnect:
				last, err := jobInsights.LastRun()
				if err != nil {
					log.Println(internal.WarningPrefix, jobInsights.Name(), "getting last run time error:", err)
				}
				if time.Since(last).Minutes() > 1 {
					err = jobInsights.RunNow()
					if err != nil {
						log.Println(internal.WarningPrefix, jobInsights.Name(), "after event run error:", err)
					}
				}
			default:
			}
		}
	}()
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
		shutdownIsActive := r.systemShutdown.Load() || internal.IsSystemShutdown()
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
	return err == nil
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
			log.Println(internal.ErrorPrefix, "auto-connect failed:", err)
			return err
		}

		server := autoconnectServer{}
		err = r.Connect(&pb.ConnectRequest{ServerTag: cfg.AutoConnectData.ServerTag}, &server)
		if connectErrorCheck(err) && server.err == nil {
			log.Println(internal.InfoPrefix, "auto-connect success")
			r.ConnectionParameters.SetConnectionParameters(pb.ConnectionSource_AUTO,
				ServerParameters{Country: cfg.AutoConnectData.Country,
					City:  cfg.AutoConnectData.City,
					Group: cfg.AutoConnectData.Group})
			return nil
		}
		log.Println(internal.ErrorPrefix, "auto-connect failed, err1:", server.err, "| err2:", err)
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
