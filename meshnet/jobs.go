package meshnet

import (
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (s *Server) StartJobs() {
	if _, err := s.scheduler.NewJob(gocron.DurationJob(2*time.Hour), gocron.NewTask(JobRefreshMeshnet(s)), gocron.WithName("job refresh meshnet")); err != nil {
		log.Println(internal.WarningPrefix, "job refresh meshnet schedule error:", err)
	}

	if _, err := s.scheduler.NewJob(
		gocron.DurationJob(5*time.Second),
		gocron.NewTask(JobMonitorFileshareProcess(s)),
		gocron.WithName("job monitor fileshare process")); err != nil {
		log.Println(internal.WarningPrefix, "job monitor fileshare process schedule error:", err)
	}

	s.scheduler.Start()
	for _, job := range s.scheduler.Jobs() {
		err := job.RunNow()
		if err != nil {
			log.Println(internal.WarningPrefix, job.Name(), "first run error:", err)
		}
	}
}

func JobRefreshMeshnet(s *Server) func() error {
	return func() error {
		// ignore what is returned, try to do it here as light as possible
		_, _ = s.RefreshMeshnet(nil, nil)
		return nil
	}
}

func JobMonitorFileshareProcess(s *Server) func() error {
	oldState := false
	return func() error {
		newState := internal.IsProcessRunning(internal.FileshareBinaryPath)
		if newState == oldState {
			// only state change triggers the modifications
			return nil
		}

		log.Println(internal.InfoPrefix, "fileshare change to running", newState)
		peers, err := s.listPeers()
		if err != nil {
			return err
		}

		isFileshareUp := newState
		for _, peer := range peers {
			if !isFileshareUp {
				s.netw.BlockFileshare(UniqueAddress{UID: peer.PublicKey, Address: peer.Address})
			} else {
				s.netw.AllowFileshare(UniqueAddress{UID: peer.PublicKey, Address: peer.Address})
			}
		}
		oldState = newState

		return nil
	}
}
