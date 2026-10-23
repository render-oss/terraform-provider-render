package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-render/internal/client"
)

const testImageDigest = "sha256:1fe6fa21bcb65be49241d28f919392179db011fadf8a77ee3bfb2c314924dd4c"

func TestParseImageReference(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		reference  string
		repository string
		tag        string
		digest     string
	}{
		"bare repository": {
			reference:  "nginx",
			repository: "nginx",
		},
		"normal tag": {
			reference:  "docker.io/library/nginx:latest",
			repository: "docker.io/library/nginx",
			tag:        "latest",
		},
		"digest with algorithm colon": {
			reference:  "ghcr.io/acme/app@" + testImageDigest,
			repository: "ghcr.io/acme/app",
			digest:     testImageDigest,
		},
		"registry port with tag": {
			reference:  "registry.example.com:5000/acme/app:release",
			repository: "registry.example.com:5000/acme/app",
			tag:        "release",
		},
		"registry port with digest": {
			reference:  "registry.example.com:5000/acme/app@" + testImageDigest,
			repository: "registry.example.com:5000/acme/app",
			digest:     testImageDigest,
		},
		"nested repository path": {
			reference:  "registry.example.com/acme/platform/app:release",
			repository: "registry.example.com/acme/platform/app",
			tag:        "release",
		},
		"non-sha256 digest algorithm is preserved": {
			reference:  "registry.example.com/acme/app@sha512:abc123",
			repository: "registry.example.com/acme/app",
			digest:     "sha512:abc123",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := parseImageReference(test.reference)
			require.NoError(t, err)
			assert.Equal(t, test.repository, got.repository)
			assert.Equal(t, test.tag, got.tag)
			assert.Equal(t, test.digest, got.digest)
		})
	}
}

func TestImageReferenceRoundTrip(t *testing.T) {
	t.Parallel()

	references := []string{
		"nginx",
		"docker.io/library/nginx:latest",
		"ghcr.io/acme/app@" + testImageDigest,
		"registry.example.com:5000/acme/app:release",
		"registry.example.com:5000/acme/app@" + testImageDigest,
		"registry.example.com/acme/platform/app:release",
		"registry.example.com/acme/app@sha512:abc123",
	}

	for _, reference := range references {
		t.Run(reference, func(t *testing.T) {
			t.Parallel()

			parsed, err := parseImageReference(reference)
			require.NoError(t, err)
			assert.Equal(t, reference, parsed.String())
		})
	}
}

func TestParseImageReferenceRejectsMalformedSeparators(t *testing.T) {
	t.Parallel()

	for _, reference := range []string{"", "@sha256:abc123", "repo@", ":tag", "repo:"} {
		t.Run(reference, func(t *testing.T) {
			t.Parallel()

			_, err := parseImageReference(reference)
			require.Error(t, err)
		})
	}
}

func TestImageURLForURLAndReference(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		repository string
		tag        string
		digest     string
		want       string
	}{
		"bare repository": {
			repository: "nginx",
			want:       "nginx",
		},
		"tag": {
			repository: "docker.io/library/nginx",
			tag:        "latest",
			want:       "docker.io/library/nginx:latest",
		},
		"digest": {
			repository: "ghcr.io/acme/app",
			digest:     testImageDigest,
			want:       "ghcr.io/acme/app@" + testImageDigest,
		},
		"historical tag and digest state prefers digest": {
			repository: "ghcr.io/acme/app",
			tag:        "historical-bad-tag",
			digest:     testImageDigest,
			want:       "ghcr.io/acme/app@" + testImageDigest,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.want, ImageURLForURLAndReference(test.repository, test.tag, test.digest))
		})
	}
}

func TestImageRuntimeSource_DigestReference(t *testing.T) {
	const (
		repository = "ghcr.io/acme/app"
	)
	imagePath := repository + "@" + testImageDigest

	image, err := ImageRuntimeSource(
		&client.Service{
			ImagePath:          &imagePath,
			RegistryCredential: &client.RegistryCredentialSummary{Id: "rgc-123"},
		},
		client.EnvSpecificDetails{},
	)
	require.NoError(t, err)

	assert.Equal(t, repository, image.ImageURL.ValueString())
	assert.Equal(t, testImageDigest, image.Digest.ValueString())
	assert.True(t, image.Tag.IsNull(), "a digest reference must not populate tag")
	assert.Equal(t, "rgc-123", image.RegistryCredentialID.ValueString())
}

func TestImageRuntimeSource_TagReference(t *testing.T) {
	imagePath := "registry.example.com:5000/acme/app:release"

	image, err := ImageRuntimeSource(
		&client.Service{ImagePath: &imagePath},
		client.EnvSpecificDetails{},
	)
	require.NoError(t, err)

	assert.Equal(t, "registry.example.com:5000/acme/app", image.ImageURL.ValueString())
	assert.Equal(t, "release", image.Tag.ValueString())
	assert.True(t, image.Digest.IsNull())
	assert.True(t, image.RegistryCredentialID.IsNull())
}

func TestImageRuntimeSource_BareReference(t *testing.T) {
	imagePath := "nginx"

	image, err := ImageRuntimeSource(
		&client.Service{ImagePath: &imagePath},
		client.EnvSpecificDetails{},
	)
	require.NoError(t, err)

	assert.Equal(t, "nginx", image.ImageURL.ValueString())
	assert.True(t, image.Tag.IsNull())
	assert.True(t, image.Digest.IsNull())
	assert.True(t, image.RegistryCredentialID.IsNull())
}

func TestImageRuntimeSource_RejectsMalformedReference(t *testing.T) {
	imagePath := "ghcr.io/acme/app@"

	_, err := ImageRuntimeSource(
		&client.Service{ImagePath: &imagePath},
		client.EnvSpecificDetails{},
	)
	require.Error(t, err)
}
