package remote

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/google/uuid"
)

const DefaultMaxGroup = 100

type Group struct {
	value int
}

// NewGroup creates a new Group instance based on a provided UUID and maximum group size.
// It computes a group value by hashing the UUID and deriving a number between 1 and provided max (inclusive).
//
// Parameters:
//   - uuid: The UUID used as the basis for group assignment
//   - max: The maximum possible group value (must be positive and not exceed DefaultMaxGroup)
//
// Returns:
//   - *Group: A new Group instance with the computed value
//   - error: An error if max is less than 1 or greater than DefaultMaxGroup
func NewGroup(uuid uuid.UUID, max int) (*Group, error) {
	if max < 1 {
		return nil, fmt.Errorf("max value must be positive, got %d", max)
	}

	if max > DefaultMaxGroup {
		return nil, fmt.Errorf("max value must not exceed %d, got %d", DefaultMaxGroup, max)
	}

	hash := sha256.Sum256(uuid[:])
	num := binary.BigEndian.Uint32(hash[:])
	value := int(num)%max + 1

	return &Group{
		value: value,
	}, nil
}
