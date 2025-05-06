package meshnet

import (
	"log"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (s *Server) StartJobs() {
	if _, err := s.scheduler.NewJob(gocron.DurationJob(5*time.Minute), gocron.NewTask(JobRefreshMeshMap(s)), gocron.WithName("job refresh mesh map")); err != nil {
		log.Println(internal.WarningPrefix, "job refresh meshnet map schedule error:", err)
	}
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

func JobRefreshMeshMap(s *Server) func() error {
	return func() error {
		// Ignore anything as this is just needed to issue a mesh map update if it is old
		// enough.
		_, _, _, _ = s.fetchPeers()
		return nil
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
		rulesController:    s.netw,
		processChecker:     defaultProcessChecker{},
	}
	return job.run
}

func (j *monitorFileshareProcessJob) run() error {
	j.mu.Lock()
	defer j.mu.Unlock()

	if !j.meshChecker.isMeshOn() {
		if j.isFileshareAllowed {
			if err := j.rulesController.ForbidFileshare(); err == nil {
				j.isFileshareAllowed = false
			}
		}
		return nil
	}

	if j.processChecker.isFileshareRunning() {
		j.rulesController.PermitFileshare()
		j.isFileshareAllowed = true
	} else {
		j.rulesController.ForbidFileshare()
		j.isFileshareAllowed = false
	}

	return nil
}

type defaultProcessChecker struct{}

func (defaultProcessChecker) isFileshareRunning() bool {
	return internal.IsProcessRunning(internal.FileshareBinaryPath)
}

type monitorFileshareProcessJob struct {
	isFileshareAllowed bool
	meshChecker        meshChecker
	rulesController    rulesController
	processChecker     processChecker
	mu                 sync.Mutex
}

type meshChecker interface {
	isMeshOn() bool
}

type rulesController interface {
	ForbidFileshare() error
	PermitFileshare() error
}

type processChecker interface {
	isFileshareRunning() bool
}
