package asset

import "errors"

// Spec for asset fetching
type Spec struct {
	// Name is the name of the asset
	Name string `mapstructure:"name"`
	// Path is the absolute path to where the asset should be installed
	Path string `mapstructure:"path"`
	// Url is the remote address used for fetching the asset
	URL string `mapstructure:"url"`
	// SHA512 is the hash of the asset tarball
	SHA512 string `mapstructure:"sha512"`
}

// Validate checks an asset ID is valid, but does not attempt to fetch
// the asset or verify the integrity of its hash
func (a Spec) Validate() error {
	if a.Name == "" {
		return errors.New("asset name cannot be empty")
	}
	return nil
}
