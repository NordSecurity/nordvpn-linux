package service

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/norduser/process"
)

type NorduserSnap struct{}

func NewNorduserSnapService() NorduserSnap {
	return NorduserSnap{}
}

func (n NorduserSnap) Enable(uint32, uint32, string) error {
	return nil
}

func (n NorduserSnap) Disable(uid uint32) error {
	if err := process.NewNorduserGRPCProcessManager(uid).StopProcess(true); err != nil {
		return fmt.Errorf("stopping norduser process: %w", err)
	}
	return nil
}

func (n NorduserSnap) Stop(uid uint32) error {
	if err := process.NewNorduserGRPCProcessManager(uid).StopProcess(false); err != nil {
		return fmt.Errorf("stopping norduser process: %w", err)
	}
	return nil
}

func (n NorduserSnap) stopAll(disable bool) {
	// #nosec G204 -- arg values are constant
	output, err := exec.Command("ps", "-C", internal.Norduserd, "-o", "uid=").CombinedOutput()
	if err != nil {
		log.Println("Failed to list running norduser instances: ", err)
	}

	uids := string(output)
	if uids == "" {
		return
	}
	uids = strings.Trim(uids, "\n")

	for _, uid := range strings.Split(uids, "\n") {
		uidInt, err := strconv.Atoi(strings.TrimSpace(uid))
		if err != nil {
			log.Printf("Invalid unix user id, failed to convert from string: %s", uid)
			continue
		}

		if err := process.NewNorduserGRPCProcessManager(uint32(uidInt)).StopProcess(disable); err != nil {
			log.Println("Failed to stop norduser for uid: ", uid)
		}
	}
}

func (n NorduserSnap) StopAll() {
	n.stopAll(false)
}

func (n NorduserSnap) DisableAll() {
	n.stopAll(true)
}
