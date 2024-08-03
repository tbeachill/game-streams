package config

// Cloudflare is a struct that holds the Cloudflare configuration values for the bot.
type Cloudflare struct {
	BucketName      string `toml:"bucket_name"`
	AccountID       string `toml:"account_id"`
	AccessKeyID     string `toml:"access_key_id"`
	AccessKeySecret string `toml:"access_key_secret"`
	Endpoint        string `toml:"endpoint"`
}
