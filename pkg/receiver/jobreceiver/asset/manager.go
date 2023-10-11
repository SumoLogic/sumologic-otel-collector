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
//
// Loops through the provided asset Specs and ensures they are installed at the
// Manager's StoragePath. Returns the first error encountered in the install.
func (m *Manager) InstallAll(ctx context.Context, all []Spec) ([]Reference, error) {
	references := make([]Reference, 0, len(all))

	for _, asset := range all {
		if ref, exists, err := m.get(asset); exists {
			m.Logger.With(
				"asset", ref.Name,
				"path", ref.Path,
			).Info("reusing previously installed runtime asset")
			references = append(references, ref)
			continue
		} else if err != nil {
			return references, fmt.Errorf("failed to access asset storage: %s", err)
		}
		ref, err := m.install(ctx, asset)
		if err != nil {
			return references, fmt.Errorf("failed to retrieve runtime asset %s: %s", asset.Name, err)
		}
		references = append(references, ref)
		m.Logger.With(
			"asset", asset.Name,
			"path", ref.Path,
		).Info("successfully installed asset")
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
