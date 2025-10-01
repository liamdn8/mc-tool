package client

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/liamdn8/mc-tool/pkg/config"
)

// CreateMinIOClient creates a MinIO client for the specified alias
func CreateMinIOClient(cfg *config.MCConfig, alias string, insecure bool, verbose bool) (*minio.Client, error) {
	aliasConfig, exists := cfg.Aliases[alias]
	if !exists {
		return nil, fmt.Errorf("alias '%s' not found in MC configuration", alias)
	}

	// Parse URL to determine if HTTPS is used
	useSSL := strings.HasPrefix(aliasConfig.URL, "https://")
	endpoint := strings.TrimPrefix(strings.TrimPrefix(aliasConfig.URL, "https://"), "http://")

	// Determine if we should skip certificate verification
	// Priority: command line flag > config setting > default (false)
	skipVerify := insecure || aliasConfig.Insecure

	// Create credentials
	creds := credentials.NewStaticV4(aliasConfig.AccessKey, aliasConfig.SecretKey, "")

	// Create transport with TLS configuration
	transport := &http.Transport{}
	if useSSL {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: skipVerify,
		}
	}

	// Create MinIO client
	client, err := minio.New(endpoint, &minio.Options{
		Creds:     creds,
		Secure:    useSSL,
		Transport: transport,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %v", err)
	}

	if verbose {
		fmt.Printf("Connected to %s (SSL: %v, Skip Verify: %v)\n", aliasConfig.URL, useSSL, skipVerify)
	}

	return client, nil
}

// ParseURL parses a MinIO URL into alias, bucket, and path components
func ParseURL(url string) (alias, bucket, path string, err error) {
	parts := strings.SplitN(url, "/", 3)
	if len(parts) < 2 {
		return "", "", "", fmt.Errorf("invalid URL format: %s (expected alias/bucket[/path])", url)
	}

	alias = parts[0]
	bucket = parts[1]
	if len(parts) > 2 {
		path = parts[2]
	}

	return alias, bucket, path, nil
}