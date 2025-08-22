package provider_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"terraform-provider-render/internal/provider"

	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func TestNewRateLimitHTTPClient(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		var called bool
		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			called = true
		}))

		client := provider.NewRateLimitHTTPClient(svr.Client(), rate.NewLimiter(rate.Limit(1), 1), time.Sleep)

		req, err := http.NewRequest(http.MethodGet, svr.URL, nil)
		require.NoError(t, err)

		_, err = client.Do(req)
		require.NoError(t, err)
		require.True(t, called)
	})

	t.Run("waits until retry after", func(t *testing.T) {
		var callCount int
		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				w.Header().Add("Retry-After", "10")
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}

			w.WriteHeader(http.StatusOK)
		}))

		fakeSleep := func(d time.Duration) {
			require.Equal(t, 10*time.Second, d)
		}

		client := provider.NewRateLimitHTTPClient(svr.Client(), rate.NewLimiter(rate.Limit(10), 1), fakeSleep)

		req, err := http.NewRequest(http.MethodGet, svr.URL, nil)
		require.NoError(t, err)

		_, err = client.Do(req)
		require.NoError(t, err)
		require.Equal(t, callCount, 2)
	})

	t.Run("waits until retry after, then backs off", func(t *testing.T) {
		var callCount int
		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount <= 2 {
				w.Header().Add("Retry-After", "10")
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}

			w.WriteHeader(http.StatusOK)
		}))

		var sleepCallCount int
		fakeSleep := func(d time.Duration) {
			sleepCallCount++
			switch sleepCallCount {
			case 1:
				require.Equal(t, 10*time.Second, d)
			case 2:
				require.Equal(t, 1*time.Second, d)
			default:
				require.Fail(t, "unexpected call to sleep")
			}
		}

		client := provider.NewRateLimitHTTPClient(svr.Client(), rate.NewLimiter(rate.Limit(10), 1), fakeSleep)

		req, err := http.NewRequest(http.MethodGet, svr.URL, nil)
		require.NoError(t, err)

		_, err = client.Do(req)
		require.NoError(t, err)
		require.Equal(t, callCount, 3)
		require.Equal(t, sleepCallCount, 2)
	})
}
