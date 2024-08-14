package meshnet

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-co-op/gocron/v2"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (s *Server) StartJobs() {
	if _, err := s.scheduler.NewJob(gocron.DurationJob(2*time.Hour), gocron.NewTask(JobRefreshMeshnet(s)), gocron.WithName("job refresh meshnet")); err != nil {
		log.Println(internal.WarningPrefix, "job refresh meshnet schedule error:", err)
	}

	// TODO: find a better place for this job
	// TODO: verify 1 sec is fine;
	if _, err := s.scheduler.NewJob(gocron.DurationJob(1*time.Second), gocron.NewTask(JobMonitorFileshareProcess(s)), gocron.WithName("job monitor fileshare process")); err != nil {
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

func JobMonitorFileshareProcess(s *Server) func() error {
	return func() error {
		log.Println(internal.DebugPrefix, "monitoring fileshare process")

		procDirs, err := os.ReadDir("/proc")
		if err != nil {
			log.Printf(internal.ErrorPrefix+" error reading /proc directory: %v\n", err)
			return fmt.Errorf("error while monitoring fileshare process: %w", err)
		}

		if !isProcessRunning(internal.FileshareBinaryPath, procDirs) {
			// TODO: disable port
			// TODO: stop monitoring after the process is gone
		}

		return nil
	}
}

// TODO: move it (maybe to internal)
func isProcessRunning(executablePath string, procDirs []os.DirEntry) bool {
	for _, dir := range procDirs {
		if _, err := strconv.Atoi(dir.Name()); err != nil {
			continue
		}
		exePath := filepath.Join("/proc", dir.Name(), "exe")
		resolvedPath, err := os.Readlink(exePath)
		if err == nil && resolvedPath == executablePath {
			return true
		}
	}
	return false
}
