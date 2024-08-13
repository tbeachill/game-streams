/*
restore.go contains the functions to restore the database from the most recent backup in the bucket.
*/
package backup

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconf "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"gamestreams/config"
	"gamestreams/logs"
)

// DownloadFile downloads the most recent file from the bucket.
func (bucket Bucket) DownloadFile() {
	fmt.Println("Downloading file...")

	// Get the list of objects in the bucket.
	listObjectsOutput, err := bucket.Client.ListObjects(context.TODO(), &s3.ListObjectsInput{
		Bucket: &bucket.Name,
	})
	if err != nil {
		logs.LogError("RESTO", "could not list objects in bucket", "err", err)
		return
	}

	// Find the most recent object.
	var mostRecent time.Time
	var mostRecentKey string
	for _, object := range listObjectsOutput.Contents {
		if object.LastModified.After(mostRecent) {
			mostRecent = *object.LastModified
			mostRecentKey = *object.Key
		}
	}

	// Download the most recent object.
	getObjectOutput, err := bucket.Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &bucket.Name,
		Key:    &mostRecentKey,
	})
	if err != nil {
		logs.LogError("RESTO", "could not download object from bucket", "err", err)
		return
	}

	// Write the object to file.
	file, err := os.Create(config.Values.Files.EncryptedDatabase)
	if err != nil {
		logs.LogError("RESTO", "could not create file", "err", err)
		return
	}
	defer file.Close()

	_, err = file.ReadFrom(getObjectOutput.Body)
	if err != nil {
		logs.LogError("RESTO", "could not write to file", "err", err)
		return
	}
	logs.LogInfo("RESTO", "file downloaded successfully", false)

}

// RestoreDB wraps the download and decrypt functions to restore the database.
func RestoreDB() {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.eu.r2.cloudflarestorage.com", config.Values.Cloudflare.AccountID),
		}, nil
	})

	// Load the default config with the custom resolver and credentials.
	cfg, err := awsconf.LoadDefaultConfig(context.TODO(),
		awsconf.WithEndpointResolverWithOptions(r2Resolver),
		awsconf.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.Values.Cloudflare.AccessKeyID, config.Values.Cloudflare.AccessKeySecret, "")),
		awsconf.WithRegion("auto"),
	)
	if err != nil {
		logs.LogError("RESTO", "restore failed: could not load default config", "err", err)
		return
	}

	bucket := Bucket{
		Name:      config.Values.Cloudflare.BucketName,
		AccountID: config.Values.Cloudflare.AccountID,
		KeyID:     config.Values.Cloudflare.AccessKeyID,
		KeySecret: config.Values.Cloudflare.AccessKeySecret,
		Client:    s3.NewFromConfig(cfg),
	}

	bucket.Client = s3.NewFromConfig(cfg)
	bucket.DownloadFile()
	decryptErr := Decrypt()
	if decryptErr != nil {
		logs.LogError("RESTO", "restore failed: could not decrypt database", "err", decryptErr)
		return
	}
	err = os.Remove(config.Values.Files.EncryptedDatabase)
	if err != nil {
		logs.LogError("RESTO", "could not delete encrypted database file", "err", err)
	}
	logs.LogInfo("RESTO", "database restored", false)
}
