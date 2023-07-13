package service

import "log"

// CombinedFileshare always tries to use main as primary method, and if it fails fallbacks to backup
type CombinedFileshare struct {
	main               Fileshare
	backup             Fileshare
	enabledThroughMain bool
}

// NewCombinedFileshare creates CombinedFileshare
func NewCombinedFileshare(main, backup Fileshare) *CombinedFileshare {
	return &CombinedFileshare{main: main, backup: backup}
}

// Enable through systemd, or if that fails - fork
func (c *CombinedFileshare) Enable(uid uint32, gid uint32) error {
	err := c.main.Enable(uid, gid)
	if err != nil {
		log.Printf("failed to enable fileshare service using main method (will try backup): %s", err)
		c.enabledThroughMain = false
		return c.backup.Enable(uid, gid)
	}
	c.enabledThroughMain = true
	return nil
}

// Disable using method that was used to Enable
func (c *CombinedFileshare) Disable(uid uint32, gid uint32) error {
	if c.enabledThroughMain {
		return c.main.Disable(uid, gid)
	}
	return c.backup.Disable(uid, gid)
}

// Stop using method that was used to Enable
func (c *CombinedFileshare) Stop(uid uint32, gid uint32) error {
	if c.enabledThroughMain {
		return c.main.Stop(uid, gid)
	}
	return c.backup.Stop(uid, gid)
}
