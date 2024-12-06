package config

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/google/uuid"
)

type MachineIDGetter interface {
	GetMachineID() uuid.UUID
}

type FileReader func(fileName string) ([]byte, error)
type HostNameReader func() (name string, err error)

type generatorFn func() ([]byte, error)

// Used to create a bitmask to identify how the machine ID was generated
const (
	isSnapEnv     = 1
	etcMachineID  = 1 << 1
	dbusMachineID = 1 << 2
	dmiProductID  = 1 << 3
	cpuSerial     = 1 << 4
	boardSerial   = 1 << 5
	productUUID   = 1 << 6
)

type MachineID struct {
	hostNameReader HostNameReader
	fileReader     FileReader

	sync.Mutex
	machineID    *uuid.UUID // store computed device ID
	usedInfoMask int16      // stores the bitmask to identify how the ID was generate
}

func NewMachineID(fileReader FileReader, hostNameReader HostNameReader) *MachineID {
	return &MachineID{
		hostNameReader: hostNameReader,
		fileReader:     fileReader,
		machineID:      nil,
		usedInfoMask:   -1,
	}
}

func (getter *MachineID) GetMachineID() (ret uuid.UUID) {
	getter.Lock()
	defer getter.Unlock()

	return getter.getMachineID()
}

// Return the bitmask to identify what information was used to generate the machine ID
func (getter *MachineID) GetUsedInformationMask() int16 {
	getter.Lock()
	defer getter.Unlock()

	if getter.machineID != nil {
		getter.getMachineID()
	}

	return getter.usedInfoMask
}

func (getter *MachineID) getMachineID() (ret uuid.UUID) {
	if getter.machineID != nil {
		return *getter.machineID
	}

	defer func() {
		if getter.machineID == nil {
			getter.machineID = &ret
			if IsUnderSnap() {
				getter.usedInfoMask |= isSnapEnv
			}
		}
	}()

	getter.usedInfoMask = 0

	id, err := getter.generateID()
	if err == nil {
		return id
	}

	// random UUID was used to generate, reset the mask to 0
	getter.usedInfoMask = 0

	// create a random UUID
	log.Println(internal.ErrorPrefix, "failed to generate machine ID", err)
	id, err = uuid.NewRandom()
	if err == nil {
		return id
	}

	// Fallback to manually generating a UUID
	log.Println(internal.ErrorPrefix, "failed to generate random UUID", err)
	return getter.fallbackGenerateUUID()
}

func (getter *MachineID) generateID() (uuid.UUID, error) {
	machineId, err := getter.calculateMachineID()
	if err != nil {
		return machineId, fmt.Errorf("failed to generate machine ID %w", err)
	}

	// order is important, changing the order would create new machine ID
	generators := []struct {
		mask int
		generatorFn
	}{
		{cpuSerial, getter.readCPUSerial},
		{productUUID, getter.readProductUUID},
		{boardSerial, getter.readMotherboardSerialNumber},
	}

	for _, pair := range generators {
		if val, err := pair.generatorFn(); err == nil {
			machineId = uuid.NewSHA1(machineId, val)
			getter.usedInfoMask |= int16(pair.mask)
		}
	}

	return machineId, nil
}

// get the machine UUID from one of the files and combine it with the hostname
func (getter *MachineID) calculateMachineID() (uuid.UUID, error) {
	sourceFiles := []struct {
		mask     int
		fileName string
	}{
		{etcMachineID, "/etc/machine-id"},
		{dbusMachineID, "/var/lib/dbus/machine-id"},
		{dmiProductID, "/sys/class/dmi/id/product_uuid"},
	}

	hostname, err := getter.hostNameReader()
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to read hostname %w", err)
	}

	for _, pair := range sourceFiles {
		data, err := getter.fileReader(pair.fileName)
		if err == nil && len(data) > 0 {
			value := strings.Trim(string(data), "\n")
			if value == "" {
				continue
			}

			machineUUID, err := uuid.Parse(value)
			if err == nil {
				getter.usedInfoMask |= int16(pair.mask)

				return uuid.NewSHA1(machineUUID, []byte(hostname)), nil
			} else {
				log.Println("failed to parse", pair.fileName, err)
			}
		} else {
			log.Println("failed to read", pair.fileName, err)
		}
	}
	return uuid.UUID{}, fmt.Errorf("failed to get device UUID")
}

func (getter *MachineID) readCPUSerial() ([]byte, error) {
	data, err := getter.fileReader("/proc/cpuinfo")
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("cpuinfo file is empty")
	}

	val, err := getValueForKey(string(data), "Serial", ":")

	if err != nil {
		return nil, fmt.Errorf("no serial number found for CPU %w", err)
	}

	return []byte(val), nil
}

func (getter *MachineID) readMotherboardSerialNumber() ([]byte, error) {
	return getter.fileReader("/sys/class/dmi/id/board_serial")
}

func (getter *MachineID) readProductUUID() ([]byte, error) {
	return getter.fileReader("/sys/class/dmi/id/product_uuid")
}

// fallback to generate a UUID using random data, when uuid.New fails
func (getter *MachineID) fallbackGenerateUUID() uuid.UUID {
	var id uuid.UUID
	// randomize the content
	_, err := rand.Read(id[:])
	if err != nil {
		log.Println(internal.ErrorPrefix, "rand failed, retry to generate uuid", err)
		return uuid.New()
	}

	// Set version (4) and variant bits according to RFC 4122
	id[6] = (id[6] & 0x0F) | 0x40 // Version 4
	id[8] = (id[8] & 0x3F) | 0x80 // Variant (10xx)

	// ensure that the random generate UUID is correct
	return uuid.MustParse(id.String())
}

func getValueForKey(fileContent string, key string, delim string) (string, error) {
	trimmedKey := strings.TrimSpace(key)
	for _, line := range strings.Split(fileContent, "\n") {
		parts := strings.SplitN(line, delim, 2)
		if len(parts) != 2 {
			continue
		}
		lineKey := strings.TrimSpace(parts[0])
		if lineKey == trimmedKey {
			return strings.TrimSpace(parts[1]), nil
		}
	}

	return "", fmt.Errorf("key %s not found", key)
}

// duplicate to avoid circular dependencies
func IsUnderSnap() bool {
	return os.Getenv("SNAP_NAME") != ""
}
