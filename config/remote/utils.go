package remote

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/google/uuid"
)

// defaultMaxGroup represents the maximum value for a rollout group,
// effectively making the value to be in range of 1-100 (inclusive) to reflect percentage-based groups.
const defaultMaxGroup = 100

// GenerateRolloutGroup creates a new RolloutGroup instance based on a provided UUID.
// It computes a group value by hashing the UUID and deriving a number between 1 and defaultMaxGroup (inclusive).
//
// Parameters:
//   - uuid: The UUID used as the basis for group assignment
//
// Returns:
//   - RolloutGroup: A new RolloutGroup instance with the computed value
func GenerateRolloutGroup(uuid uuid.UUID) int {
	hash := sha256.Sum256(uuid[:])
	num := binary.BigEndian.Uint32(hash[:])
	value := int(num%uint32(defaultMaxGroup)) + 1
	return value
}
