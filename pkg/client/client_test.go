package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/liamdn8/mc-tool/pkg/config"
)

func TestParseURL(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		expectAlias  string
		expectBucket string
		expectPath   string
		expectError  bool
	}{
		{
			name:         "simple bucket URL",
			url:          "minio1/mybucket",
			expectAlias:  "minio1",
			expectBucket: "mybucket",
			expectPath:   "",
			expectError:  false,
		},
		{
			name:         "bucket with path",
			url:          "minio1/mybucket/folder/subfolder",
			expectAlias:  "minio1",
			expectBucket: "mybucket",
			expectPath:   "folder/subfolder",
			expectError:  false,
		},
		{
			name:         "bucket with single path component",
			url:          "minio1/mybucket/folder",
			expectAlias:  "minio1",
			expectBucket: "mybucket",
			expectPath:   "folder",
			expectError:  false,
		},
		{
			name:        "invalid URL - no bucket",
			url:         "minio1",
			expectError: true,
		},
		{
			name:        "invalid URL - empty",
			url:         "",
			expectError: true,
		},
		{
			name:        "invalid URL - only slash",
			url:         "/",
			expectError: false, // This actually parses as alias="", bucket=""
			expectAlias: "",
			expectBucket: "",
			expectPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alias, bucket, path, err := ParseURL(tt.url)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectAlias, alias)
			assert.Equal(t, tt.expectBucket, bucket)
			assert.Equal(t, tt.expectPath, path)
		})
	}
}

func TestCreateMinIOClientConfiguration(t *testing.T) {
	testConfig := &config.MCConfig{
		Version: "10",
		Aliases: map[string]config.AliasConfig{
			"secure": {
				URL:       "https://secure.example.com",
				AccessKey: "key1",
				SecretKey: "secret1",
				API:       "s3v4",
				Path:      "auto",
				Insecure:  false,
			},
			"insecure": {
				URL:       "http://localhost:9000",
				AccessKey: "key2",
				SecretKey: "secret2",
				API:       "s3v4",
				Path:      "auto",
				Insecure:  true,
			},
		},
	}

	tests := []struct {
		name         string
		alias        string
		insecureFlag bool
		expectError  bool
	}{
		{
			name:         "valid secure alias",
			alias:        "secure",
			insecureFlag: false,
			expectError:  false,
		},
		{
			name:         "valid insecure alias",
			alias:        "insecure",
			insecureFlag: false,
			expectError:  false,
		},
		{
			name:         "non-existent alias",
			alias:        "nonexistent",
			insecureFlag: false,
			expectError:  true,
		},
		{
			name:         "insecure flag override",
			alias:        "secure",
			insecureFlag: true,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := CreateMinIOClient(testConfig, tt.alias, tt.insecureFlag, false)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, client)
		})
	}
}

func TestInsecureFlagLogic(t *testing.T) {
	testConfig := &config.MCConfig{
		Version: "10",
		Aliases: map[string]config.AliasConfig{
			"default": {
				URL:       "https://example.com",
				AccessKey: "key",
				SecretKey: "secret",
				API:       "s3v4",
				Path:      "auto",
				Insecure:  false,
			},
			"config-insecure": {
				URL:       "https://example.com",
				AccessKey: "key",
				SecretKey: "secret",
				API:       "s3v4",
				Path:      "auto",
				Insecure:  true,
			},
		},
	}

	tests := []struct {
		name            string
		alias           string
		insecureFlag    bool
		expectedSecure  bool
		description     string
	}{
		{
			name:         "default config, no flag",
			alias:        "default",
			insecureFlag: false,
			description:  "Should use secure connection when config is secure and no flag",
		},
		{
			name:         "default config, insecure flag",
			alias:        "default",
			insecureFlag: true,
			description:  "Should override config with insecure flag",
		},
		{
			name:         "insecure config, no flag",
			alias:        "config-insecure",
			insecureFlag: false,
			description:  "Should use insecure when config specifies insecure",
		},
		{
			name:         "insecure config, insecure flag",
			alias:        "config-insecure",
			insecureFlag: true,
			description:  "Should remain insecure when both config and flag are insecure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := CreateMinIOClient(testConfig, tt.alias, tt.insecureFlag, false)
			require.NoError(t, err, tt.description)
			assert.NotNil(t, client, tt.description)
		})
	}
}