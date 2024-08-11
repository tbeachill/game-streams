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
		logs.LogError("SCHED", "could not open database file", "err", err)
	} else {
		defer file.Close()
		_, err = bucket.Client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(bucket.Name),
			Key:    aws.String(objectKey),
			Body:   file,
		})
		if err != nil {
			logs.LogError("SCHED", "could not upload database file", "err", err)
		} else {
			logs.LogInfo("SCHED", "database file uploaded successfully", false)
		}
	}
}

// CleanUp removes backup files older than 28 days from the bucket.
func (bucket Bucket) CleanUp() {
	objects, err := bucket.Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket.Name),
	})
	if err != nil {
		logs.LogError("SCHED", "could not list objects in bucket", "err", err)
		return
	}
	for _, object := range objects.Contents {
		if object.LastModified.Before(time.Now().AddDate(0, 0, -28)) {
			_, err := bucket.Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
				Bucket: aws.String(bucket.Name),
				Key:    object.Key,
			})
			if err != nil {
				logs.LogError("SCHED", "could not delete object",
					"obj", *object.Key,
					"err", err)
			} else {
				logs.LogInfo("SCHED", "deleted object", false, "key", *object.Key)
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
		logs.LogError("SCHED", "could not list objects in bucket", "err", err)
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
		logs.LogInfo("SCHED", "backup not supported on windows", false)
		return
	}

	err := Encrypt()
	if err != nil {
		logs.LogError("SCHED", "backup failed: could not encrypt database", "err", err)
		return
	}

	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.eu.r2.cloudflarestorage.com", config.Values.Cloudflare.AccountID),
		}, nil
	})
	cfg, err := awsconf.LoadDefaultConfig(context.TODO(),
		awsconf.WithEndpointResolverWithOptions(r2Resolver),
		awsconf.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.Values.Cloudflare.AccessKeyID, config.Values.Cloudflare.AccessKeySecret, "")),
		awsconf.WithRegion("auto"),
	)
	if err != nil {
		logs.LogError("SCHED", "backup failed: could not load default config", "err", err)
		return
	}
	bucket := Bucket{
		Name:      config.Values.Cloudflare.BucketName,
		AccountID: config.Values.Cloudflare.AccountID,
		KeyID:     config.Values.Cloudflare.AccessKeyID,
		KeySecret: config.Values.Cloudflare.AccessKeySecret,
		Client:    s3.NewFromConfig(cfg),
	}
	if bucket.TodaysBackupExists() {
		logs.LogInfo("SCHED", "backup already exists for today", false)
		return
	}
	bucket.UploadFile(config.Values.Files.EncryptedDatabase)
	bucket.CleanUp()
}
