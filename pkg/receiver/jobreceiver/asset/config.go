package asset

import (
	"errors"
	"fmt"
	"net/url"
)

const (
	binDir     = "bin"
	libDir     = "lib"
	includeDir = "include"
)

// Spec for asset fetching
type Spec struct {
	// Name is the name of the asset
	Name string `mapstructure:"name"`
	// Url is the remote address used for fetching the asset
	URL string `mapstructure:"url"`
	// SHA512 is the hash of the asset tarball
	SHA512 string `mapstructure:"sha512"`
}

// Validate checks an asset ID is valid, but does not attempt to fetch
// the asset or verify the integrity of its hash
func (a *Spec) Validate() error {
	if a.Name == "" {
		return errors.New("asset name cannot be empty")
	}
	if a.SHA512 == "" {
		return errors.New("asset sha cannot be empty")
	}
	if _, err := url.Parse(a.URL); err != nil {
		return fmt.Errorf("could not parse url %s: %s", a.URL, err)
	}
	return nil
}
