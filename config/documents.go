package config

// Documents is a struct that holds the URLs of the documents.
type Documents struct {
	// The URL to the terms of service.
	TermsOfService string `toml:"terms_of_service"`
	// The URL to the privacy policy.
	PrivacyPolicy string `toml:"privacy_policy"`
	// The URL to the changelog.
	Changelog string `toml:"changelog"`
}
