package config

import (
	"crypto/rand"
	"fmt"
	"log"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/google/uuid"
)

type MachineIDGetter interface {
	GetMachineID() uuid.UUID
}

type FileReader func(fileName string) ([]byte, error)
type HostNameReader func() (name string, err error)

type GeneratorFn func() ([]byte, error)

type MachineID struct {
	hostNameReader HostNameReader
	fileReader     FileReader
}

func NewMachineID(fileReader FileReader, hostNameReader HostNameReader) MachineID {
	return MachineID{
		hostNameReader: hostNameReader,
		fileReader:     fileReader,
	}
}

func (getter MachineID) GetMachineID() uuid.UUID {
	id, err := getter.generateID()
	if err == nil {
		return id
	}

	// create a random UUID
	log.Println(internal.ErrorPrefix, "failed to generate machine ID", err)
	id, err = uuid.NewRandom()
	if err == nil {
		return id
	}

	// Fallback to manually generating a UUID
	log.Println(internal.ErrorPrefix, "failed to generate random UUID", err)
	var fallbackUUID uuid.UUID
	_, err = rand.Read(fallbackUUID[:])
	if err != nil {
		log.Println(internal.ErrorPrefix, "rand failed, retry to generate uuid", err)
		return uuid.New()
	}

	// Set version (4) and variant bits according to RFC 4122
	fallbackUUID[6] = (fallbackUUID[6] & 0x0F) | 0x40 // Version 4
	fallbackUUID[8] = (fallbackUUID[8] & 0x3F) | 0x80 // Variant (10xx)

	// ensure that the random generate UUID is correct
	return uuid.MustParse(fallbackUUID.String())
}

func (getter MachineID) generateID() (uuid.UUID, error) {
	machineId, err := getter.calculateMachineID()
	if err != nil {
		return machineId, fmt.Errorf("failed to generate machine ID %w", err)
	}

	// order is important, changing the order would create new machine ID
	generators := []GeneratorFn{
		getter.readCPUSerial,
		getter.readProductUUID,
		getter.readMotherboardSerialNumber,
	}

	for _, generatorFn := range generators {
		if val, err := generatorFn(); err == nil {
			machineId = uuid.NewSHA1(machineId, val)
		}
	}

	return machineId, nil
}

// get the machine UUID from one of the files and combine it with the hostname
func (getter MachineID) calculateMachineID() (uuid.UUID, error) {
	sourceFiles := []string{
		"/etc/machine-id",
		"/var/lib/dbus/machine-id",
		"/sys/class/dmi/id/product_uuid",
	}

	hostname, err := getter.hostNameReader()
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to read hostname %w", err)
	}

	for _, fileName := range sourceFiles {
		data, err := getter.fileReader(fileName)
		if err == nil && len(data) > 0 {
			machineUUID, err := uuid.Parse(string(data))
			if err == nil {
				return uuid.NewSHA1(machineUUID, []byte(hostname)), nil
			}
		}
	}
	return uuid.UUID{}, fmt.Errorf("failed to get device UUID")
}

func (getter MachineID) readCPUSerial() ([]byte, error) {
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

func (getter MachineID) readMotherboardSerialNumber() ([]byte, error) {
	return getter.fileReader("/sys/class/dmi/id/board_serial")
}

func (getter MachineID) readProductUUID() ([]byte, error) {
	return getter.fileReader("/sys/class/dmi/id/product_uuid")
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
