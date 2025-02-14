package jobs

import (
	"log"
	"time"

	"github.com/NordSecurity/nordvpn-linux/internal"
	meshnet "github.com/NordSecurity/nordvpn-linux/meshnet/interfaces"
	"github.com/go-co-op/gocron/v2"
)

// Starts or stops the meshnet map refresh job
func ConfigureMeshnetMapRefresher(
	enabled bool,
	scheduler gocron.Scheduler,
	meshnetChecker meshnet.MeshnetChecker,
	fetcher meshnet.MeshnetFetcher,
	interval time.Duration,
) error {
	if enabled {
		job, err := scheduler.NewJob(
			gocron.DurationJob(interval),
			gocron.NewTask(JobRefreshMeshnetMap(meshnetChecker, fetcher)),
			gocron.WithName("refresh meshnet map"),
			gocron.WithTags(internal.MeshnetMapJobTag),
		)
		if err != nil {
			log.Println(internal.ErrorPrefix, "job refresh meshnet schedule error:", err)
			return err
		}

		log.Println(internal.DebugPrefix, "meshnet map refresh job scheduled")

		if err := job.RunNow(); err != nil {
			log.Println(internal.ErrorPrefix, "failed to run meshnet map refresh job", err)
			return err
		}
	} else {
		scheduler.RemoveByTags(internal.MeshnetMapJobTag)
		log.Println(internal.DebugPrefix, "stop meshnet map refresh job")
	}
	return nil
}

func JobRefreshMeshnetMap(
	meshnetChecker meshnet.MeshnetChecker,
	fetcher meshnet.MeshnetFetcher,
) func() {
	return func() {
		if !meshnetChecker.IsMeshnetOn() {
			log.Println(internal.InfoPrefix, "updating meshnet map called when meshnet is not enabled")
			return
		}

		_, err := fetcher.RefreshMeshnetMap(nil)
		if err != nil {
			log.Println(internal.ErrorPrefix, "job update meshnet map failed", err)
		}
	}
}
