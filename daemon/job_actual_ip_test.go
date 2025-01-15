package daemon

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/events"

	"github.com/stretchr/testify/assert"
)

type mockInsights struct {
	insightsFunc insightFunc
}

func (m *mockInsights) Insights() (*core.Insights, error) {
	return nil, errors.New("not implemented")
}

func (m *mockInsights) InsightsViaTunnel() (*core.Insights, error) {
	return m.insightsFunc()
}

type insightFunc = func() (*core.Insights, error)

func TestInsightsIPUntilSuccess(t *testing.T) {
	tests := []struct {
		name        string
		ctxTimeout  time.Duration
		insightsAPI func() func() (*core.Insights, error)
		expectedIP  string
		expectedErr string
	}{
		{
			name:       "Successful IP retrieval",
			ctxTimeout: time.Second,
			insightsAPI: func() insightFunc {
				return func() (*core.Insights, error) {
					return &core.Insights{IP: "192.168.1.1"}, nil
				}
			},
			expectedIP:  "192.168.1.1",
			expectedErr: "",
		},
		{
			name:       "API returns error",
			ctxTimeout: time.Millisecond * 100,
			insightsAPI: func() insightFunc {
				return func() (*core.Insights, error) {
					return nil, errors.New("API error")
				}
			},
			expectedIP: "invalid IP",
			// deadline exceeded and not API error, because when an API errors occurs, it should retry until the context is cancelled
			expectedErr: context.DeadlineExceeded.Error(),
		},
		{
			name:       "Context canceled",
			ctxTimeout: 0,
			insightsAPI: func() insightFunc {
				return func() (*core.Insights, error) {
					time.Sleep(10 * time.Millisecond)
					return &core.Insights{IP: "192.168.1.1"}, nil
				}
			},
			expectedIP:  "invalid IP",
			expectedErr: context.DeadlineExceeded.Error(),
		},
		{
			name:       "Invalid IP format",
			ctxTimeout: time.Second,
			insightsAPI: func() insightFunc {
				return func() (*core.Insights, error) {
					return &core.Insights{IP: netip.Addr{}.String()}, nil
				}
			},
			expectedIP: "invalid IP",
			// deadline exceeded and not invalid ip error, because when an invalid ip error occurs, it should retry until the context is cancelled
			expectedErr: context.DeadlineExceeded.Error(),
		},
		{
			name:       "Successful IP retrieval on third attempt",
			ctxTimeout: time.Second,
			insightsAPI: func() insightFunc {
				callCount := 0
				return func() (*core.Insights, error) {
					return func() (*core.Insights, error) {
						callCount++
						if callCount < 3 {
							return nil, errors.New("temporary error")
						}
						return &core.Insights{IP: "192.168.1.10"}, nil
					}()
				}
			},
			expectedIP:  "192.168.1.10",
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.ctxTimeout)
			defer cancel()

			insightFunc := tt.insightsAPI()
			api := &mockInsights{
				insightsFunc: insightFunc,
			}

			ip, err := insightsIPUntilSuccess(ctx, api, func(i int) time.Duration {
				return time.Millisecond * 10
			})

			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.expectedErr)
			}
			assert.Equal(t, tt.expectedIP, ip.String())
		})
	}
}

func TestUpdateActualIP(t *testing.T) {
	tests := []struct {
		name        string
		isConnected bool
		insightsAPI func() (*core.Insights, error)
		expectedIP  string
		expectedErr error
	}{
		{
			name:        "Successful IP update",
			isConnected: true,
			insightsAPI: func() (*core.Insights, error) {
				return &core.Insights{IP: "192.168.1.2"}, nil
			},
			expectedIP:  "192.168.1.2",
			expectedErr: nil,
		},
		{
			name:        "Not connected",
			isConnected: false,
			insightsAPI: func() (*core.Insights, error) {
				return &core.Insights{IP: "192.168.1.2"}, nil
			},
			expectedIP:  "invalid IP",
			expectedErr: nil,
		},
		{
			name:        "Error from Insights API",
			isConnected: true,
			insightsAPI: func() (*core.Insights, error) {
				return nil, errors.New("API error")
			},
			expectedIP:  "invalid IP",
			expectedErr: context.DeadlineExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			dm := NewDataManager("", "", "", "", events.NewDataUpdateEvents())
			api := &mockInsights{
				insightsFunc: tt.insightsAPI,
			}

			err := updateActualIP(dm, api, ctx, tt.isConnected)

			assert.ErrorIs(t, err, tt.expectedErr)

			dm.mu.Lock()
			actualIP := dm.actualIP
			dm.mu.Unlock()

			assert.Equal(t, actualIP.String(), tt.expectedIP)
		})
	}
}
