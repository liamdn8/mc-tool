package validation

import (
	"context"
	"fmt"
	"strings"

	"github.com/minio/minio-go/v7"
)

// CheckBucketConfiguration performs comprehensive bucket configuration validation
func CheckBucketConfiguration(ctx context.Context, client *minio.Client, bucketName string) error {
	fmt.Printf("Checking bucket: %s\n\n", bucketName)

	// Check if bucket exists
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("bucket %s does not exist", bucketName)
	}

	fmt.Println("✅ Bucket exists")

	// Check versioning configuration
	checkVersioningConfig(ctx, client, bucketName)

	// Check notification configuration
	checkNotificationConfig(ctx, client, bucketName)

	// Check lifecycle configuration
	checkLifecycleConfig(ctx, client, bucketName)

	// Check encryption configuration
	checkEncryptionConfig(ctx, client, bucketName)

	// Check bucket policy
	checkBucketPolicyConfig(ctx, client, bucketName)

	return nil
}

func checkVersioningConfig(ctx context.Context, client *minio.Client, bucketName string) {
	versioningConfig, err := client.GetBucketVersioning(ctx, bucketName)
	if err != nil {
		fmt.Printf("❌ Versioning: Failed to retrieve configuration - %v\n", err)
		return
	}

	if versioningConfig.Status == "Enabled" {
		fmt.Println("✅ Versioning: Enabled")
	} else {
		fmt.Println("⚠️  Versioning: Disabled - Consider enabling for data protection")
	}
}

func checkNotificationConfig(ctx context.Context, client *minio.Client, bucketName string) {
	notification, err := client.GetBucketNotification(ctx, bucketName)
	if err != nil {
		fmt.Printf("❌ Event Notifications: Failed to retrieve configuration - %v\n", err)
		return
	}

	totalConfigs := len(notification.LambdaConfigs) + len(notification.TopicConfigs) + len(notification.QueueConfigs)

	if totalConfigs == 0 {
		fmt.Println("➖ Event Notifications: Not configured")
	} else {
		fmt.Printf("✅ Event Notifications: %d configurations found\n", totalConfigs)
		if len(notification.LambdaConfigs) > 0 {
			fmt.Printf("   - Lambda configurations: %d\n", len(notification.LambdaConfigs))
		}
		if len(notification.TopicConfigs) > 0 {
			fmt.Printf("   - Topic configurations: %d\n", len(notification.TopicConfigs))
		}
		if len(notification.QueueConfigs) > 0 {
			fmt.Printf("   - Queue configurations: %d\n", len(notification.QueueConfigs))
		}
	}
}

func checkLifecycleConfig(ctx context.Context, client *minio.Client, bucketName string) {
	lifecycle, err := client.GetBucketLifecycle(ctx, bucketName)
	if err != nil {
		fmt.Println("➖ Object Lifecycle: Not configured")
		return
	}

	fmt.Printf("✅ Object Lifecycle: %d rules configured\n", len(lifecycle.Rules))

	hasIncompleteUploadRule := false

	for _, rule := range lifecycle.Rules {
		fmt.Printf("   - Rule '%s': %s\n", rule.ID, rule.Status)
		if rule.AbortIncompleteMultipartUpload.DaysAfterInitiation > 0 {
			hasIncompleteUploadRule = true
		}
	}

	if !hasIncompleteUploadRule {
		fmt.Println("   💡 Consider adding rules to abort incomplete multipart uploads")
	}
}

func checkEncryptionConfig(ctx context.Context, client *minio.Client, bucketName string) {
	encryption, err := client.GetBucketEncryption(ctx, bucketName)
	if err != nil {
		fmt.Println("⚠️  Server-side Encryption: Not configured - Consider enabling for data security")
		return
	}

	if len(encryption.Rules) > 0 {
		rule := encryption.Rules[0]
		fmt.Printf("✅ Server-side Encryption: %s configured\n", rule.Apply.SSEAlgorithm)
	}
}

func checkBucketPolicyConfig(ctx context.Context, client *minio.Client, bucketName string) {
	policy, err := client.GetBucketPolicy(ctx, bucketName)
	if err != nil {
		fmt.Println("➖ Bucket Policy: Not configured")
		return
	}

	fmt.Println("✅ Bucket Policy: Configured")

	// Basic policy analysis
	if strings.Contains(policy, `"s3:*"`) {
		fmt.Println("   ⚠️  Warning: Policy contains wildcard actions - review for security")
	}
	if strings.Contains(policy, `"Resource": "*"`) {
		fmt.Println("   ⚠️  Warning: Policy contains wildcard resources - review for security")
	}
}