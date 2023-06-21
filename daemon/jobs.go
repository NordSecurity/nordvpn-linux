package daemon

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

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
		if err := r.netw.SetKillSwitch(cfg.AutoConnectData.Whitelist); err != nil {
			log.Println(internal.ErrorPrefix, "starting killswitch:", err)
			return
		}
		return
	}
	return
}

func (r *RPC) StopKillSwitch() error {
	var cfg config.Config
	err := r.cm.Load(&cfg)
	if err != nil {
		return fmt.Errorf("loading daemon config: %w", err)
	}

	if cfg.KillSwitch {
		if err := r.netw.UnsetKillSwitch(); err != nil {
			return fmt.Errorf("unsetting killswitch: %w", err)
		}
	}
	return nil
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

func (r *RPC) StartAutoConnect() {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return
	}

	if !cfg.AutoConnect {
		return
	}

	udp := []int64{}
	for port := range cfg.AutoConnectData.Whitelist.Ports.UDP {
		udp = append(udp, port)
	}

	tcp := []int64{}
	for port := range cfg.AutoConnectData.Whitelist.Ports.TCP {
		tcp = append(tcp, port)
	}

	subnets := []string{}
	for subnet := range cfg.AutoConnectData.Whitelist.Subnets {
		subnets = append(subnets, subnet)
	}

	server := autoconnectServer{}
	if err := r.Connect(&pb.ConnectRequest{
		ServerTag: cfg.AutoConnectData.ServerTag,
	}, &server); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	if server.err != nil {
		log.Println(internal.ErrorPrefix, server.err)
	}
	return
}
