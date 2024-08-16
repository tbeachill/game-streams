package config

// Cloudflare is a struct that holds the credentials for accessing the backup
// Cloudflare R2 Storage Bucket.
type Cloudflare struct {
	// The name of the bucket.
	BucketName string `toml:"bucket_name"`
	// The endpoint URL of the bucket.
	Endpoint string `toml:"endpoint"`
	// The account ID of the owner of the bucket.
	AccountID string `toml:"account_id"`
	// The access key ID for the bucket.
	AccessKeyID string `toml:"access_key_id"`
	// The access key secret for the bucket.
	AccessKeySecret string `toml:"access_key_secret"`
}
