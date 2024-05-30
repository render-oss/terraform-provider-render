package testhelpers

import (
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
	"terraform-provider-render/internal/provider"
	"terraform-provider-render/internal/provider/common"
)

func SetupRecordingProvider(t *testing.T, casetteName string) map[string]func() (tfprotov6.ProviderServer, error) {
	return SetupRecordingProviderConfigureWait(t, casetteName, false)
}

var emailRegex = regexp.MustCompile(`[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}`)

func scrubString(i *cassette.Interaction, from, to string) {
	i.Request.URL = strings.ReplaceAll(i.Request.URL, from, to)
	i.Request.Body = strings.ReplaceAll(i.Request.Body, from, to)
	i.Response.Body = strings.ReplaceAll(i.Response.Body, from, to)
}

func scrubeRegex(i *cassette.Interaction, re *regexp.Regexp, to string) {
	i.Request.URL = re.ReplaceAllString(i.Request.URL, to)
	i.Request.Body = re.ReplaceAllString(i.Request.Body, to)
	i.Response.Body = re.ReplaceAllString(i.Response.Body, to)
}

func SetupRecordingProviderConfigureWait(t *testing.T, casetteName string, waitForComplete bool) map[string]func() (tfprotov6.ProviderServer, error) {
	mode := recorder.ModeRecordOnce
	updateRecordings := os.Getenv("UPDATE_RECORDINGS")
	if updateRecordings == "true" {
		mode = recorder.ModeRecordOnly
	}

	r, err := recorder.NewWithOptions(
		&recorder.Options{
			CassetteName:       "testdata/" + casetteName,
			Mode:               mode,
			SkipRequestLatency: true,
			RealTransport:      http.DefaultTransport,
		},
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, r.Stop())
	})

	replaceAuthHeader := func(i *cassette.Interaction) error {
		i.Request.Headers.Set("Authorization", "some-api-key")
		return nil
	}
	r.AddHook(replaceAuthHeader, recorder.AfterCaptureHook)

	replaceHost := func(i *cassette.Interaction) error {
		i.Request.Host = "https://api.testing.render.com/v1"
		u, err := url.Parse(i.Request.URL)
		if err != nil {
			return err
		}
		u.Host = "api.testing.render.com"
		i.Request.URL = u.String()

		return nil
	}
	r.AddHook(replaceHost, recorder.AfterCaptureHook)

	replaceOwnerID := func(i *cassette.Interaction) error {
		scrubString(i, os.Getenv("RENDER_OWNER_ID"), "some-owner-id")
		return nil
	}
	r.AddHook(replaceOwnerID, recorder.AfterCaptureHook)

	replaceEmail := func(i *cassette.Interaction) error {
		scrubeRegex(i, emailRegex, "email@example.com")
		return nil
	}
	r.AddHook(replaceEmail, recorder.AfterCaptureHook)

	replacePostgresConnectionInfo := func(i *cassette.Interaction) error {
		scrubeRegex(i, regexp.MustCompile(`PGPASSWORD=[^ ]+`), "PGPASSWORD=thirtytwocharacterpasswooooooord")
		scrubeRegex(i, regexp.MustCompile(`(postgres://[^:]+:)[^@]+(@)`), `$1thirtytwocharacterpasswooooooord$2`)
		scrubeRegex(i, regexp.MustCompile(`"password":"[^"]+"`), `"password":"thirtytwocharacterpasswooooooord"`)
		return nil
	}
	r.AddHook(replacePostgresConnectionInfo, recorder.AfterCaptureHook)

	providerOpts := []provider.ConfigFunc{
		provider.WithHTTPClient(r.GetDefaultClient()),
		provider.WithWaitForDeployCompletion(waitForComplete),
	}

	// we need valid credentials if we are recording
	if r.IsRecording() {
		t.Log("Recording interactions for " + casetteName)
		require.NotZero(t, os.Getenv("RENDER_HOST"), "RENDER_HOST must be set when recording")
		require.NotZero(t, os.Getenv("RENDER_OWNER_ID"), "RENDER_OWNER_ID must be set when recording")
		require.NotZero(t, os.Getenv("RENDER_API_KEY"), "RENDER_API_KEY must be set when recording")
	} else {
		providerOpts = append(providerOpts,
			provider.WithHost("https://api.testing.render.com/v1"),
			provider.WithOwnerID("some-owner-id"),
			provider.WithAPIKey("some-api-key"),
			provider.WithPoller(&common.TestPoller),
		)
	}

	return map[string]func() (tfprotov6.ProviderServer, error){
		"render": providerserver.NewProtocol6WithError(
			provider.New(
				"test",
				providerOpts...,
			)()),
	}
}
