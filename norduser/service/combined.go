package service

import (
	"fmt"
	"log"
	"sync"
)

type processType int

const (
	systemd processType = iota
	fork
)

type Combined struct {
	mu               sync.Mutex
	systemd          SystemdNorduser
	fork             ForkNorduser
	uidToProcessType map[uint32]processType
}

func NewNorduserService() Combined {
	return Combined{
		uidToProcessType: make(map[uint32]processType),
	}
}

func (c *Combined) Enable(uid uint32, gid uint32) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.systemd.Enable(uid)
	if err == nil {
		c.uidToProcessType[uid] = systemd
		return nil
	}

	log.Printf("failed to enable norduserd via systemd: %s, will fallback to fork implementation", err)
	if err := c.fork.Enable(uid, gid); err != nil {
		return fmt.Errorf("enabling norduserd via fork: %w", err)
	}

	c.uidToProcessType[uid] = fork
	return nil
}

func (c *Combined) Disable(uid uint32) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch c.uidToProcessType[uid] {
	case systemd:
		if err := c.systemd.Disable(uid); err != nil {
			return fmt.Errorf("disabling systemd norduserd: %w", err)
		}
	case fork:
		if err := c.fork.Stop(uid); err != nil {
			return fmt.Errorf("stopping fork norduserd: %w", err)
		}
	}

	delete(c.uidToProcessType, uid)

	return nil
}

func (c *Combined) Stop(uid uint32) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch c.uidToProcessType[uid] {
	case systemd:
		if err := c.systemd.Stop(uid); err != nil {
			return fmt.Errorf("stopping systemd norduserd: %w", err)
		}
	case fork:
		if err := c.fork.Stop(uid); err != nil {
			return fmt.Errorf("stopping fork norduserd: %w", err)
		}
	}

	delete(c.uidToProcessType, uid)

	return nil
}
