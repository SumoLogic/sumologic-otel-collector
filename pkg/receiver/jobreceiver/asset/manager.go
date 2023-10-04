package asset

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

type Manager struct {
	// Fetcher for downloading assets
	Fetcher Fetcher
	// StoragePath directory where assets will be unpacked and referenced
	StoragePath string
	Logger      *zap.SugaredLogger
}

func (m *Manager) assetPath(asset Spec) string {
	base := m.StoragePath
	if base == "" {
		base = os.TempDir()
	}
	return filepath.Join(base, asset.SHA512)
}

// Validate the assets are valid and that the manager can access its storage
// path
func (m *Manager) Validate(assets []Spec) error {
	for _, asset := range assets {
		if err := asset.Validate(); err != nil {
			return fmt.Errorf("invalid asset %s: %w", asset.Name, err)
		}
		if ref, _, err := m.get(asset); err != nil {
			return fmt.Errorf("could not check for asset at path %s: %w", ref.Path, err)
		}
	}
	return nil
}

// InstallAll runtime assets on the host file system under the StoragePath
// directory.
func (m *Manager) InstallAll(ctx context.Context, all []Spec) ([]Reference, error) {

	type installTuple struct {
		Err error
		Ref Reference
	}

	results := make(chan installTuple)
	ictx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, asset := range all {
		go func(a Spec) {
			var result installTuple
			if ref, exists, err := m.get(a); exists {
				m.Logger.With(
					"asset", result.Ref.Name,
					"path", result.Ref.Path,
				).Info("reusing previously installed runtime asset")
				result.Ref = ref
			} else if err != nil {
				result.Err = fmt.Errorf("failed to access asset storage: %s", err)
			} else {
				result.Ref, result.Err = m.install(ictx, a)
				if result.Err == nil {
					m.Logger.With(
						"asset", result.Ref.Name,
						"path", result.Ref.Path,
					).Info("successfully installed asset")
				}
			}
			results <- result
		}(asset)
	}

	var references []Reference
	for i := 0; i < len(all); i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case r := <-results:
			if r.Err != nil {
				return references, r.Err
			}
			references = append(references, r.Ref)
		}
	}

	return references, nil
}

// get returns true when the manager has the asset locally. Returns an
// error when it was unable to check for the asset in the filesystem, such as
// when the configured storage path is invalid or permissions forbid it.
func (m *Manager) get(asset Spec) (Reference, bool, error) {
	ref := Reference{
		Name: asset.Name,
		Path: m.assetPath(asset),
	}
	if _, err := os.Stat(m.assetPath(asset)); err != nil {
		if os.IsNotExist(err) {
			return ref, false, nil
		}
		return ref, false, err
	}
	return ref, true, nil
}

// install a runtime asset in the host file system inside the  StoragePath
// directory. Uses the Fetcher to download the asset tarball, verifies that
// against the configured asset SHA512 and archives it in a child directory of
// StoragePath named by asset SHA.
func (m *Manager) install(ctx context.Context, asset Spec) (Reference, error) {
	ref := Reference{
		Name: asset.Name,
		Path: m.assetPath(asset),
	}
	f, err := m.Fetcher.Fetch(ctx, asset.URL)
	if err != nil {
		return ref, err
	}

	v := &sha512Verifier{}
	err = v.Verify(f, asset.SHA512)

	if err != nil {
		return ref, err
	}

	e := &archiveExpander{}

	return ref, e.Expand(f, ref.Path)
}
