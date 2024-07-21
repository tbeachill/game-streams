package utils

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

type Bucket struct {
	Name      string
	AccountID string
	KeyID     string
	KeySecret string
	Client    *s3.Client
}

// UploadFile uploads a file to the bucket.
func (bucket Bucket) UploadFile(filePath string) {
	file, err := os.Open(filePath)
	currentDate := time.Now().UTC().Format("2006-01-02")
	fileName := strings.Split(filePath, "/")[len(strings.Split(filePath, "/"))-1]
	objectKey := fmt.Sprintf("%s_%s", fileName, currentDate)

	if err != nil {
		LogError("SCHED", "could not open database file", "err", err)
	} else {
		defer file.Close()
		_, err = bucket.Client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(bucket.Name),
			Key:    aws.String(objectKey),
			Body:   file,
		})
		if err != nil {
			LogError("SCHED", "could not upload database file", "err", err)
		} else {
			LogInfo("SCHED", "database file uploaded successfully", false)
		}
	}
}

// CleanUp removes backup files older than 14 days from the bucket.
func (bucket Bucket) CleanUp() {
	objects, err := bucket.Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket.Name),
	})
	if err != nil {
		LogError("SCHED", "could not list objects in bucket", "err", err)
		return
	}
	for _, object := range objects.Contents {
		if object.LastModified.Before(time.Now().AddDate(0, 0, -14)) {
			_, err := bucket.Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
				Bucket: aws.String(bucket.Name),
				Key:    object.Key,
			})
			if err != nil {
				LogError("SCHED", "could not delete object",
					"obj", *object.Key,
					"err", err)
			} else {
				LogInfo("SCHED", "deleted object", false, "key", *object.Key)
			}
		}
	}
}

// TodaysBackupExists checks if todays backup already exists in the bucket.
func (bucket Bucket) TodaysBackupExists() bool {
	objects, err := bucket.Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket.Name),
	})
	if err != nil {
		LogError("SCHED", "could not list objects in bucket", "err", err)
		return false
	}
	currentDate := time.Now().UTC().Format("2006-01-02")
	for _, object := range objects.Contents {
		if strings.Contains(*object.Key, currentDate) {
			return true
		}
	}
	return false
}

// BackupDB uploads the database file to the Cloudflare bucket.
func BackupDB() {
	if runtime.GOOS == "windows" {
		LogInfo("SCHED", "backup not supported on windows", false)
		return
	}
	godotenv.Load(".env")
	var bucketName = os.Getenv("CF_BUCKET_NAME")
	var accountId = os.Getenv("CF_ACCOUNT_ID")
	var accessKeyId = os.Getenv("CF_ACCESS_KEY_ID")
	var accessKeySecret = os.Getenv("CF_ACCESS_KEY_SECRET")

	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.eu.r2.cloudflarestorage.com", accountId),
		}, nil
	})
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		LogError("SCHED", "backup failed: could not load default config", "err", err)
		return
	}
	bucket := Bucket{
		Name:      bucketName,
		AccountID: accountId,
		KeyID:     accessKeyId,
		KeySecret: accessKeySecret,
		Client:    s3.NewFromConfig(cfg),
	}
	if bucket.TodaysBackupExists() {
		LogInfo("SCHED", "backup already exists for today", false)
		return
	}
	bucket.UploadFile(Files.DB)
	bucket.CleanUp()
}
