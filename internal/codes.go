package internal

import "fmt"

const (
	// Success
	CodeSuccess          int64 = 1000
	CodeConnecting       int64 = 1001
	CodeConnected        int64 = 1002
	CodeDisconnected     int64 = 1003
	CodeInteraction      int64 = 1004
	CodeProxyNone        int64 = 1005
	CodeSuccessWithArg   int64 = 1006
	CodeSuccessWithoutAC int64 = 1007

	// Warning
	CodeNothingToDo      int64 = 2000
	CodeVPNRunning       int64 = 2002
	CodeVPNNotRunning    int64 = 2003
	CodeUFWDisabled      int64 = 2004
	CodeTokenInvalidated int64 = 2005

	// Error
	CodeFailure      int64 = 3000
	CodeUnauthorized int64 = 3001
	CodeFormatError  int64 = 3003
	// CodeConfigError is returned when config loading and/or saving fails.
	CodeConfigError                    int64 = 3004
	CodeEmptyPayloadError              int64 = 3005
	CodeOffline                        int64 = 3007
	CodeAccountExpired                 int64 = 3008
	CodeVPNMisconfig                   int64 = 3010
	CodeDaemonOffline                  int64 = 3013
	CodeGatewayError                   int64 = 3014
	CodeOutdated                       int64 = 3015
	CodeDependencyError                int64 = 3017
	CodeNoNewDataError                 int64 = 3019
	CodeNoService                      int64 = 3020
	CodeExpiredRenewToken              int64 = 3021
	CodeTokenRenewError                int64 = 3022
	CodeKillSwitchError                int64 = 3023
	CodeBadRequest                     int64 = 3024
	CodeConflict                       int64 = 3025
	CodeInternalError                  int64 = 3026
	CodeOpenVPNAccountExpired          int64 = 3031
	CodeServerUnavailable              int64 = 3032
	CodeTagNonexisting                 int64 = 3033
	CodeDoubleGroupError               int64 = 3034
	CodeTokenLoginFailure              int64 = 3035
	CodeGroupNonexisting               int64 = 3036
	CodeAutoConnectServerNotObfuscated int64 = 3037
	CodeAutoConnectServerObfuscated    int64 = 3038
	CodeTokenInvalid                   int64 = 3039
	CodePrivateSubnetLANDiscovery      int64 = 3040
	CodeDedicatedIPRenewError          int64 = 3041
	CodeDedicatedIPNoServer            int64 = 3042
	CodeDedicatedIPServiceButNoServers int64 = 3043
	CodeAllowlistInvalidSubnet         int64 = 3044
	CodeAllowlistSubnetNoop            int64 = 3045
	CodeAllowlistPortOutOfRange        int64 = 3046
	CodeAllowlistPortNoop              int64 = 3047
	CodePqAndMeshnetSimultaneously     int64 = 3048
	CodePqWithoutNordlynx              int64 = 3049
	CodeFeatureHidden                  int64 = 3050
	CodeTechnologyDisabled             int64 = 3051
	CodeNotInNordVPNGroup              int64 = 3052
	CodeConsentMissing                 int64 = 3052
	CodeExpiredAccessToken             int64 = 3053
	CodeRevokedAccessToken             int64 = 3054
)

type ErrorWithCode struct {
	Code int64
}

func NewErrorWithCode(code int64) error {
	return &ErrorWithCode{Code: code}
}

func (e *ErrorWithCode) Error() string {
	return fmt.Sprintf("Error with code %d", e.Code)
}
