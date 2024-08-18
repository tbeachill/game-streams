/*
backup.go contains functions to backup the database to a Cloudflare R2 bucket.
*/
package backup

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconf "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"gamestreams/config"
	"gamestreams/logs"
)

// Bucket represents a Cloudflare R2 bucket.
type Bucket struct {
	Name      string
	AccountID string
	KeyID     string
	KeySecret string
	Client    *s3.Client
}

// UploadFile uploads a file to the bucket. The current date is appended to the file name.
func (bucket Bucket) UploadFile(filePath string) {
	file, err := os.Open(filePath)
	currentDate := time.Now().UTC().Format("2006-01-02")
	fileName := strings.Split(filePath, "/")[len(strings.Split(filePath, "/"))-1]
	objectKey := fmt.Sprintf("%s_%s", fileName, currentDate)

	if err != nil {
		logs.LogError("BCKUP", "could not open database file", "err", err)
	} else {
		defer file.Close()
		_, err = bucket.Client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(bucket.Name),
			Key:    aws.String(objectKey),
			Body:   file,
		})
		if err != nil {
			logs.LogError("BCKUP", "could not upload database file", "err", err)
		} else {
			logs.LogInfo("BCKUP", "database file uploaded successfully", false)
		}
	}
}

// CleanUp removes backup files older than the number of days specified in config.toml.
func (bucket Bucket) CleanUp() {
	objects, err := bucket.Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket.Name),
	})
	if err != nil {
		logs.LogError("BCKUP", "could not list objects in bucket", "err", err)
		return
	}
	for _, object := range objects.Contents {
		if object.LastModified.Before(time.Now().AddDate(0, 0, -config.Values.Backup.DaysToKeep)) {
			_, err := bucket.Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
				Bucket: aws.String(bucket.Name),
				Key:    object.Key,
			})
			if err != nil {
				logs.LogError("BCKUP", "could not delete object",
					"obj", *object.Key,
					"err", err)
			} else {
				logs.LogInfo("BCKUP", "deleted object", false, "key", *object.Key)
			}
		}
	}
}

// BackupDB wraps the other functions in this package to create a backup of the database.
func BackupDB() {
	if runtime.GOOS == "windows" {
		logs.LogInfo("BCKUP", "backup not supported on windows", false)
		return
	}

	err := Encrypt()
	if err != nil {
		logs.LogError("BCKUP", "backup failed: could not encrypt database", "err", err)
		return
	}
	// Create a new endpoint resolver that resolves to the R2 endpoint.
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.%s", config.Values.Cloudflare.AccountID,
				config.Values.Cloudflare.Endpoint),
		}, nil
	})
	// Load the default configuration with the R2 endpoint resolver and Cloudflare credentials.
	cfg, err := awsconf.LoadDefaultConfig(context.TODO(),
		awsconf.WithEndpointResolverWithOptions(r2Resolver),
		awsconf.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.Values.Cloudflare.AccessKeyID, config.Values.Cloudflare.AccessKeySecret, "")),
		awsconf.WithRegion("auto"),
	)
	if err != nil {
		logs.LogError("BCKUP", "backup failed: could not load default config", "err", err)
		return
	}
	bucket := Bucket{
		Name:      config.Values.Cloudflare.BucketName,
		AccountID: config.Values.Cloudflare.AccountID,
		KeyID:     config.Values.Cloudflare.AccessKeyID,
		KeySecret: config.Values.Cloudflare.AccessKeySecret,
		Client:    s3.NewFromConfig(cfg),
	}
	bucket.UploadFile(config.Values.Files.EncryptedDatabase)
	bucket.CleanUp()
}
