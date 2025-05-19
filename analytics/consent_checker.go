package analytics

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/analytics/pb"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// XXX: add docs
type ConsentChecker struct {
	cm config.Manager
}

func NewConsentChecker(cfgManager config.Manager) *ConsentChecker {
	return &ConsentChecker{cm: cfgManager}
}

func (cc *ConsentChecker) StreamInterceptor(
	srv any,
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
) error {
	if isCallAllowed(info) {
		return nil
	}
	if err := cc.checkConsent(); err != nil {
		return err
	}
	return nil
}

func (cc *ConsentChecker) checkConsent() error {
	wasConsentGiven, err := cc.wasConsentGiven()
	if err != nil {
		return err
	}

	if wasConsentGiven {
		return nil
	}

	st := status.New(codes.FailedPrecondition, internal.MissingConsentMsg)
	ds, err := st.WithDetails(&pb.ErrMissingConsent{})
	if err != nil {
		return st.Err()
	}
	return ds.Err()
}

func (cc *ConsentChecker) wasConsentGiven() (bool, error) {
	var cfg config.Config
	if err := cc.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return false, err
	}
	return cfg.Analytics.Get(), nil
}

func (cc *ConsentChecker) UnaryInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
) (any, error) {
	if isCallAllowed(info) {
		return nil, nil
	}
	if err := cc.checkConsent(); err != nil {
		return nil, err
	}
	return nil, nil
}

func isCallAllowed(info any) bool {
	var fullMethod string
	switch i := info.(type) {
	case *grpc.UnaryServerInfo:
		fullMethod = i.FullMethod
	case *grpc.StreamServerInfo:
		fullMethod = i.FullMethod
	default:
		return false
	}

	return fullMethod == "/pb.Daemon/SetAnalytics"
}

func IsMissingConsent(err error) bool {
	errMsg := strings.ToLower(fmt.Sprintf("%v", err))
	return strings.Contains(errMsg, internal.MissingConsentMsg)
}
