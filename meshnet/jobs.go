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
		gocron.DurationJob(1*time.Second),
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
	job := monitorFileshareProcessJob{
		isFileshareAllowed: false,
		meshChecker:        s,
		networker:          s.netw,
	}
	return job.run
}

type monitorFileshareProcessJob struct {
	isFileshareAllowed bool
	meshChecker        meshChecker
	networker          Networker
}

type meshChecker interface {
	isMeshOn() bool
}

func (j *monitorFileshareProcessJob) run() error {
	if !j.meshChecker.isMeshOn() {
		if j.isFileshareAllowed {
			if err := j.networker.ForbidFileshare(); err == nil {
				j.isFileshareAllowed = false
			}
		}
		return nil
	}

	if internal.IsProcessRunning(internal.FileshareBinaryPath) {
		j.networker.PermitFileshare()
		j.isFileshareAllowed = true
	} else {
		j.networker.ForbidFileshare()
		j.isFileshareAllowed = false
	}

	return nil
}
