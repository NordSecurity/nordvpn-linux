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
