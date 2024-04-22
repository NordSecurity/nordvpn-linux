// snapconf package contains the code required when code is run under snapd as snap package.
// It is named snapconf because snap directory name under root is reserved by snapctl.
package snapconf

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/snapconf/pb"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	EnvSnapName       = "SNAP_NAME"
	EnvSnapRealHome   = "SNAP_REAL_HOME" // from snapd version 2.46
	EnvSnapUserCommon = "SNAP_USER_COMMON"
	EnvSnapUserData   = "SNAP_USER_DATA"
)

// Interface defines a snap interface as described in
// https://snapcraft.io/docs/supported-interfaces
type Interface string

const (
	InterfaceNetwork         Interface = "network"
	InterfaceNetworkBind     Interface = "network-bind"
	InterfaceNetworkControl  Interface = "network-control"
	InterfaceFirewallControl Interface = "firewall-control"
	InterfaceNetworkObserve  Interface = "network-observe"
	InterfaceHome            Interface = "home"
)

// IsUnderSnap defines whether the current process is executed under snapd
func IsUnderSnap() bool {
	return os.Getenv(EnvSnapName) != ""
}

// ConnChecker is a gRPC middleware which checks whether all necessary snap interfaces are
// connected to the package and returns a corresponding error message to the client so it can
// inform users on manual actions needed.
// NOTE: It is solely designed for UX purposes and not security. Security is handled by the
// AppArmor under the snapd.
type ConnChecker struct {
	requirements    []Interface
	recommendations []Interface
	publisherErr    events.Publisher[error]
}

// NewConnChecker is a constructor for the [ConnChecker]. It constructs it with a set of hardcoded
// pre-defined requirement list. It is assumed that constructor is called once in the beginning of
// the process and it defines whether it makes sense to suggest snap to recommend process restart
// on specific interface connections.
// Parameters:
//   - requirements - list of requirements used in this process
//   - recommendations - list of requirements to be recommended via gRPC in case of a checker
//     error. It may be useful if multiple services are running under the same snap and they
//     require different snap connections. but for smooth UX user is recommended to connect
//     everything at once. E. g. nordvpnd + nordfileshared
//   - publisherErr - publisher for error reporting
func NewConnChecker(
	requirements []Interface,
	recommendations []Interface,
	publisherErr events.Publisher[error],
) *ConnChecker {
	connectedInterfaces, err := getConnectedInterfaces()
	if err != nil && publisherErr != nil {
		publisherErr.Publish(fmt.Errorf(
			"listing connecting snap interfaces for restart suggestion: %w", err,
		))
	} else {
		err := storeInitialConnections(connectedInterfaces)
		if err != nil && publisherErr != nil {
			publisherErr.Publish(fmt.Errorf("storing initial connections: %w", err))
		}
	}
	return &ConnChecker{
		requirements:    requirements,
		recommendations: recommendations,
		publisherErr:    publisherErr,
	}
}

func (c *ConnChecker) StreamInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
) error {
	if err := c.permissionCheck(); err != nil {
		return err
	}
	return nil
}

func (c *ConnChecker) UnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
) (interface{}, error) {
	if err := c.permissionCheck(); err != nil {
		return nil, err
	}
	return nil, nil
}

func storeInitialConnections(connections []Interface) error {
	// Sometimes if process is started without network-control interface connected, socket()
	// syscall fails when setting the WireGuard tunnel configuration even after connecting the
	// interface. Therefore, process restart is the only known workaround
	connDir := filepath.Join(os.Getenv("SNAP_COMMON"), "connections")
	if err := os.MkdirAll(
		connDir,
		internal.PermUserRWX,
	); err != nil {
		return fmt.Errorf("creating directory for initial snap connections: %w", err)
	}

	for _, iface := range []Interface{InterfaceNetworkControl, InterfaceFirewallControl} {
		if err := os.WriteFile(
			filepath.Join(connDir, string(iface)),
			boolToBytes(slices.Contains(connections, iface)),
			internal.PermUserRW,
		); err != nil {
			return fmt.Errorf("saving initial connection for %s: %w", iface, err)
		}
	}
	return nil
}

func (c *ConnChecker) permissionCheck() error {
	connectedInterfaces, err := getConnectedInterfaces()
	if err != nil {
		// If listing interfaces fails, it is OK to log error and try to execute gRPC
		// method anyways. It may worsen UX due to generic errors for the user but misuse
		// of permission checker will not disturb the whole application
		c.publisherErr.Publish(fmt.Errorf("listing connected snap interfaces: %w", err))
		return nil
	}
	if !containsAll(connectedInterfaces, c.requirements) {
		missingConnections := sub(c.recommendations, connectedInterfaces)
		st := status.New(codes.PermissionDenied, "Snap permissions required")
		ds, err := st.WithDetails(
			&pb.ErrMissingConnections{
				MissingConnections: convertStringSlice[Interface, string](
					missingConnections,
				),
			},
		)
		if err != nil {
			return st.Err()
		}
		return ds.Err()
	}
	return nil
}

// getConnectedInterfaces returns list of connected snap interfaces for the current snap.
func getConnectedInterfaces() ([]Interface, error) {
	out, err := exec.Command("snapctl", "is-connected", "--list").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf(
			"executing `snapctl is-connected --list`: %w: %s", err, string(out),
		)
	}
	return convertStringSlice[string, Interface](
		strings.Split(strings.TrimSpace(string(out)), "\n"),
	), nil
}

// convertStringSlice and copies all elements of a given slice to a new slice of a given type.
func convertStringSlice[T1 ~string, T2 ~string](s []T1) []T2 {
	res := make([]T2, len(s))
	for i, e := range s {
		res[i] = T2(e)
	}
	return res
}

// sub returns list of elements which are present in s1 but not in s2.
func sub[E comparable](s1 []E, s2 []E) []E {
	out := []E{}
	for _, e := range s1 {
		if !slices.Contains(s2, e) {
			out = append(out, e)
		}
	}
	return out
}

// containsAll returns true if  s1 contains all of the elements of s2, regardless of order.
// Otherwise, returns false.
func containsAll[E comparable](s1 []E, s2 []E) bool {
	if len(s1) < len(s2) {
		return false
	}
	for _, e := range s2 {
		if !slices.Contains(s1, e) {
			return false
		}
	}
	return true
}

// boolToBytes converts true to 1 and false to 0.
func boolToBytes(val bool) []byte {
	if val {
		return []byte{'1'}
	}
	return []byte{'0'}
}

func RealUserHomeDir() string {
	dir := os.Getenv(EnvSnapRealHome)
	if dir != "" {
		return dir
	}

	// for snapd before version 2.46 try to "guess" the home dir based on SNAP_USER_DATA variable
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	dir = os.Getenv(EnvSnapUserData)
	if homeDir == dir {
		// For non-classic snaps, HOME environment variable is re-written to SNAP_USER_DATA
		// Typical value: /home/_user_name_/snap/_snap_name_/_snap_revision_
		for i := 0; i < 3; i++ {
			dir = filepath.Dir(dir)
		}
		return dir
	}

	return ""
}
