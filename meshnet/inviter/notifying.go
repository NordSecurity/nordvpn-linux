package inviter

import (
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/google/uuid"
)

type NotifyingInviter struct {
	inner mesh.Inviter
	sub   events.PublishSubcriber[[]string]
}

func NewNotifyingInviter(inner mesh.Inviter, sub events.PublishSubcriber[[]string]) *NotifyingInviter {
	return &NotifyingInviter{
		inner: inner,
		sub:   sub,
	}
}

func (i *NotifyingInviter) Invite(
	token string,
	self uuid.UUID,
	email string,
	doIAllowInbound bool,
	doIAllowRouting bool,
	doIAllowLocalNetwork bool,
	doIAllowFileshare bool,
) error {
	return i.inner.Invite(
		token,
		self,
		email,
		doIAllowInbound,
		doIAllowRouting,
		doIAllowLocalNetwork,
		doIAllowFileshare,
	)
}

func (i *NotifyingInviter) Sent(token string, self uuid.UUID) (mesh.Invitations, error) {
	return i.inner.Sent(token, self)
}

func (i *NotifyingInviter) Received(token string, self uuid.UUID) (mesh.Invitations, error) {
	return i.inner.Received(token, self)
}

func (i *NotifyingInviter) Accept(
	token string,
	self uuid.UUID,
	invitation uuid.UUID,
	doIAllowInbound bool,
	doIAllowRouting bool,
	doIAllowLocalNetwork bool,
	doIAllowFileshare bool,
) error {
	if err := i.inner.Accept(
		token,
		self,
		invitation,
		doIAllowInbound,
		doIAllowRouting,
		doIAllowLocalNetwork,
		doIAllowFileshare,
	); err != nil {
		return err
	}
	i.sub.Publish([]string{invitation.String()})
	return nil
}

func (i *NotifyingInviter) Reject(token string, self uuid.UUID, invitation uuid.UUID) error {
	return i.inner.Reject(token, self, invitation)
}

func (i *NotifyingInviter) Revoke(token string, self uuid.UUID, invitation uuid.UUID) error {
	return i.inner.Revoke(token, self, invitation)
}
