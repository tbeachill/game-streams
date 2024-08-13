package config

// Documents is a struct that holds the URLs of the documents.
type Documents struct {
	TermsOfService string `toml:"terms_of_service"`
	PrivacyPolicy  string `toml:"privacy_policy"`
	Changelog      string `toml:"changelog"`
}
