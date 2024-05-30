package testhelpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/require"
)

func StaticResponse(body any) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Add("Content-Type", "application/json")

		switch v := body.(type) {
		case string:
			_, _ = resp.Write([]byte(v))
		case []byte:
			_, _ = resp.Write(v)
		default:
			res, err := json.Marshal(body)
			if err != nil {
				resp.WriteHeader(http.StatusInternalServerError)
				_, _ = resp.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err)))
			}

			_, _ = resp.Write(res)
		}
	}
}

func ListResponse(responses ...any) http.HandlerFunc {
	callCount := -1
	return func(resp http.ResponseWriter, req *http.Request) {
		callCount++
		if callCount < len(responses) {
			StaticResponse(responses[callCount])(resp, req)
			return
		}

		StaticResponse([]struct{}{})(resp, req)
	}
}

func WithRequestBodySnapshot(t *testing.T, v any, handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		require.NoError(t, json.NewDecoder(req.Body).Decode(&v))
		bs, err := json.MarshalIndent(v, "", "\t")
		require.NoError(t, err)
		cupaloy.New(cupaloy.SnapshotFileExtension(".json")).SnapshotT(t, string(bs))

		handlerFunc(resp, req)
	}
}

func NewMockRenderAPI(responses map[string]http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		for pathPattern, handler := range responses {
			if matched, err := regexp.MatchString("^"+pathPattern+"$", req.URL.Path); matched && err == nil {
				resp.Header().Add("Content-Type", "application/json")
				handler(resp, req)
				return
			}
		}

		http.NotFound(resp, req)
	}))
}
