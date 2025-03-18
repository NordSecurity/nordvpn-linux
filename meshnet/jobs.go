package meshnet

import (
	"context"
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/meshnet/pb"
)

func (s *Server) StartJobs() {
	if _, err := s.scheduler.NewJob(gocron.DurationJob(5*time.Minute), gocron.NewTask(JobRefreshMeshMap(s)), gocron.WithName("job refresh mesh map")); err != nil {
		log.Println(internal.WarningPrefix, "job refresh meshnet map schedule error:", err)
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
		resp, err := s.RefreshMeshnet(context.Background(), &pb.Empty{})
		if err == nil {
			if resp, ok := resp.Response.(*pb.MeshnetResponse_ServiceError); ok {
				// Retry after possible failure on the backend server
				if resp.ServiceError == pb.ServiceErrorCode_API_FAILURE {
					_, _ = s.RefreshMeshnet(context.Background(), &pb.Empty{})
				}
			}
		}
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
