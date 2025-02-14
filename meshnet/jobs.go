package meshnet

import (
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	interfaces "github.com/NordSecurity/nordvpn-linux/meshnet/interfaces"
	"github.com/NordSecurity/nordvpn-linux/meshnet/jobs"
)

func (s *Server) StartJobs(
	meshnetStatusPublisher events.PublishSubcriber[bool],
	cfg config.Config,
) {
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

	// monitors the meshnet status and starts/stops the meshnet map refreshing job
	meshnetStatusPublisher.Subscribe(func(enabled bool) error {
		return jobs.ConfigureMeshnetMapRefresher(enabled, s.scheduler, s, s, internal.MeshnetMapUpdateInterval)
	})

	if cfg.Mesh {
		jobs.ConfigureMeshnetMapRefresher(true, s.scheduler, s, s, internal.MeshnetMapUpdateInterval)
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
	if !j.meshChecker.IsMeshnetOn() {
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
	meshChecker        interfaces.MeshnetChecker
	rulesController    rulesController
	processChecker     processChecker
}

type rulesController interface {
	ForbidFileshare() error
	PermitFileshare() error
}

type processChecker interface {
	isFileshareRunning() bool
}
