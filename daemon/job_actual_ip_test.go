package daemon

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/events"
)

// func TestJobActualIP(t *testing.T) {

// }

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
		expectedErr error
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
			expectedErr: nil,
		},
		{
			name:       "API returns error",
			ctxTimeout: time.Second,
			insightsAPI: func() insightFunc {
				return func() (*core.Insights, error) {
					return nil, errors.New("API error")
				}
			},
			expectedIP:  "",
			expectedErr: context.DeadlineExceeded,
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
			expectedIP:  "",
			expectedErr: context.DeadlineExceeded,
		},
		{
			name:       "Invalid IP format",
			ctxTimeout: time.Second,
			insightsAPI: func() insightFunc {
				return func() (*core.Insights, error) {
					return &core.Insights{IP: "invalid-ip"}, nil
				}
			},
			expectedIP:  "",
			expectedErr: errors.New("invalid IP format"),
		},
		{
			name:       "Successful IP retrieval on third attempt",
			ctxTimeout: 2 * time.Second,
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
			expectedErr: nil,
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

			if (err != nil) != (tt.expectedErr != nil) {
				t.Fatalf("unexpected error: got %v, want %v", err, tt.expectedErr)
			}

			if err == nil && ip.String() != tt.expectedIP {
				t.Errorf("unexpected IP: got %v, want %v", ip, tt.expectedIP)
			}
		})
	}
}

func TestJobActualIP(t *testing.T) {
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

			job := JobActualIP(dm, api)
			err := job(ctx, tt.isConnected)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("unexpected error: got %v, want %v", err, tt.expectedErr)
			}

			dm.mu.Lock()
			actualIP := dm.actualIP
			dm.mu.Unlock()

			if actualIP.String() != tt.expectedIP {
				t.Errorf("unexpected IP: got %v, want %v", actualIP, tt.expectedIP)
			}
		})
	}
}
