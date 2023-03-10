package meshnet

import (
	"log"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (s *Server) StartJobs() {
	if _, err := s.scheduler.Every(2).Hours().Do(JobRefreshMeshnet(s)); err != nil {
		log.Println(internal.WarningPrefix, "starting job refresh meshnet", err)
	}
	s.scheduler.RunAll()
	s.scheduler.StartBlocking()
}

func JobRefreshMeshnet(s *Server) func() error {
	return func() error {
		// ignore what is returned, try to do it here as light as possible
		_, _ = s.RefreshMeshnet(nil, nil)
		return nil
	}
}
