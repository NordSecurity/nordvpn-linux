package cli

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

// itsUsMsg is hardcoded here in a different place than the original
// one just because the original changes the tests would fail and this
// should be updated manually
const itsUsMsg = "It's not you, it's us. We're having trouble with " +
	"our servers. If the issue persists, please contact " +
	"our customer support."

func TestReadForConfirmation(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name   string
		input  io.Reader
		output bool
	}{
		{
			name:   "no input",
			input:  strings.NewReader(""),
			output: true,
		},
		{
			name:   "a newline",
			input:  strings.NewReader("\n"),
			output: true,
		},
		{
			name:   "a number",
			input:  strings.NewReader("5"),
			output: true,
		},
		{
			name:   "predicting Anton's input",
			input:  strings.NewReader("\\"),
			output: true,
		},
		{
			name:  "lowercase n",
			input: strings.NewReader("n"),
		},
		{
			name:  "uppercase n",
			input: strings.NewReader("N"),
		},
		{
			name:   "lowercase y",
			input:  strings.NewReader("y"),
			output: true,
		},
		{
			name:   "uppercase y",
			input:  strings.NewReader("Y"),
			output: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, readForConfirmation(test.input, test.name), test.output)
		})
	}
}

func TestRespondToInviteResponseToError(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name  string
		resp  *pb.RespondToInviteResponse
		email string
		err   error
	}{
		{
			name:  "unknown",
			email: "a@b.c",
			err:   errors.New(itsUsMsg),
		},
		{
			name: "service response code",
			resp: &pb.RespondToInviteResponse{
				Response: &pb.RespondToInviteResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			},
			err: internal.ErrNotLoggedIn,
		},
		{
			name: "service response code",
			resp: &pb.RespondToInviteResponse{
				Response: &pb.RespondToInviteResponse_RespondToInviteErrorCode{
					RespondToInviteErrorCode: pb.RespondToInviteErrorCode_NO_SUCH_INVITATION,
				},
			},
			email: "b@c.d",
			err:   errors.New("no invitation from 'b@c.d' was found"),
		},
		{
			name: "No error",
			resp: &pb.RespondToInviteResponse{
				Response: &pb.RespondToInviteResponse_Empty{
					Empty: &pb.Empty{},
				},
			},
			email: "d@e.f",
			err:   nil,
		},
	}
	for _, tt := range tests {
		assert.Equal(
			t,
			tt.err,
			respondToInviteResponseToError(
				tt.resp,
				tt.email,
			),
		)
	}
}

func TestRespondToInviteErrorCodeToError(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name  string
		code  pb.RespondToInviteErrorCode
		email string
		err   error
	}{
		{
			name:  "no invitation",
			code:  pb.RespondToInviteErrorCode_NO_SUCH_INVITATION,
			email: "a@b.c",
			err: errors.New(
				"no invitation from 'a@b.c' was found",
			),
		},
		{
			name:  "unknown error",
			code:  pb.RespondToInviteErrorCode(3),
			email: "a@b.d",
			err:   errors.New(itsUsMsg),
		},
	}
	for _, tt := range tests {
		assert.Equal(
			t,
			tt.err,
			respondToInviteErrorCodeToError(
				tt.code,
				tt.email,
			),
		)
	}
}
