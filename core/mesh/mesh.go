// Package mesh implements mesh related data structure conversions.
package mesh

import (
	"github.com/google/uuid"
)

// Registry defines a set of operations used to interact with the rest of the mesh.
type Registry interface {
	// Register Self to mesh network.
	Register(token string, self Machine) (*Machine, error)
	// Update already registered peer.
	Update(token string, id uuid.UUID, info MachineUpdateRequest) error
	// Configure interaction with specific peer.
	Configure(
		token string,
		id uuid.UUID,
		peerID uuid.UUID,
		peerUpdateInfo PeerUpdateRequest,
	) error
	// Unregister Peer from the mesh network.
	Unregister(token string, self uuid.UUID) error
	// List given peer's neighbours in the mesh network.
	List(token string, self uuid.UUID) (MachinePeers, error)
	Map(token string, self uuid.UUID) (*MachineMap, error)
	// Unpair invited peer.
	Unpair(token string, self uuid.UUID, peer uuid.UUID) error
	// NotifyNewTransfer notifies a device about a new incoming transfer (outgoing from this
	// device perspective)
	NotifyNewTransfer(
		token string,
		self uuid.UUID,
		peer uuid.UUID,
		fileName string,
		fileCount int,
		transferID string,
	) error
}

// Inviter defines a set of operations for managing personal mesh network.
type Inviter interface {
	// Invite to mesh network.
	Invite(
		token string,
		self uuid.UUID,
		email string,
		doIAllowInbound bool,
		doIAllowRouting bool,
		doIAllowLocalNetwork bool,
		doIAllowFileshare bool,
	) error
	// Sent invitations to other users.
	Sent(token string, self uuid.UUID) (Invitations, error)
	// Received invitations from other users.
	Received(token string, self uuid.UUID) (Invitations, error)
	// Accept an invitation.
	Accept(
		token string,
		self uuid.UUID,
		invitation uuid.UUID,
		doIAllowInbound bool,
		doIAllowRouting bool,
		doIAllowLocalNetwork bool,
		doIAllowFileshare bool,
	) error
	// Reject an invitation.
	Reject(token string, self uuid.UUID, invitation uuid.UUID) error
	// Revoke an invitation.
	Revoke(token string, self uuid.UUID, invitation uuid.UUID) error
}

// Invitations to join other mesh networks.
type Invitations []Invitation
