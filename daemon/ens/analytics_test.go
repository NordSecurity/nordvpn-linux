package ens

import (
	"encoding/json"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func contextValueByPath(paths []events.ContextValue, path string) any {
	for _, cv := range paths {
		if cv.Path == path {
			return cv.Value
		}
	}
	return nil
}

func TestVPNConnectionErrorEvent_ToDebuggerEvent(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		code     events.VPNConnectionError
		wantCode string
		wantDesc string
	}{
		{
			name:     "unknown",
			code:     events.VPNConnectionErrorUnknown,
			wantCode: "unknown",
			wantDesc: "unknown error",
		},
		{
			name:     "connection limit reached",
			code:     events.VPNConnectionErrorConnectionLimitReached,
			wantCode: "connection_limit_reached",
			wantDesc: "connection limit reached",
		},
		{
			name:     "server maintenance",
			code:     events.VPNConnectionErrorServerMaintenance,
			wantCode: "server_maintenance",
			wantDesc: "server maintenance",
		},
		{
			name:     "unauthenticated",
			code:     events.VPNConnectionErrorUnauthenticated,
			wantCode: "unauthenticated",
			wantDesc: "unauthenticated",
		},
		{
			name:     "superseded",
			code:     events.VPNConnectionErrorSuperseded,
			wantCode: "superseded",
			wantDesc: "superseded by newer connection",
		},
		{
			name:     "unrecognized code falls back to default",
			code:     events.VPNConnectionError(999),
			wantCode: "unrecognized",
			wantDesc: "unrecognized error code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			debuggerEvent := newVPNConnectionErrorEvent(tt.code).ToDebuggerEvent()

			require.NotNil(t, debuggerEvent)
			require.NotEmpty(t, debuggerEvent.JsonData)

			var decoded vpnConnectionErrorEvent
			require.NoError(t, json.Unmarshal([]byte(debuggerEvent.JsonData), &decoded))

			assert.Equal(t, "nordvpn-linux", decoded.Namespace)
			assert.Equal(t, "ens", decoded.Subscope)
			assert.Equal(t, "ens_connection_error", decoded.Event)
			assert.Equal(t, tt.wantCode, decoded.Code)
			assert.Equal(t, tt.wantDesc, decoded.Description)

			assert.Equal(t, tt.wantCode, contextValueByPath(debuggerEvent.KeyBasedContextPaths, "ens.code"))
			assert.NotEmpty(t, debuggerEvent.GeneralContextPaths)
		})
	}
}
