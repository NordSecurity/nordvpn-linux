package remote

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/google/uuid"
)

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
	value := int(num%defaultMaxGroup) + 1
	return value
}
