package system

import (
	"path/filepath"

	"github.com/mudler/LocalAI/pkg/xsysinfo"
	"github.com/mudler/xlog"
)

type Backend struct {
	BackendsPath       string
	BackendsSystemPath string
}

type Model struct {
	ModelsPath string
}

type SystemState struct {
	GPUVendor string
	Backend   Backend
	Model     Model
	VRAM      uint64
	// StagingPath holds in-flight downloads before extraction; empty
	// falls back to the OS temp directory.
	StagingPath string

	systemCapabilities string

	// Backend image fallback tag configuration
	BackendImagesReleaseTag string
	BackendImagesBranchTag  string
	BackendDevSuffix        string
	// PreferDevelopmentBackends prefers compatible runnable installed development
	// builds for generic aliases and installs the development image as the primary
	// backend URI (the released image becomes a fallback) rather than only using
	// development as a download fallback when the released image is missing.
	// This resolver preference is fixed at startup, not hot-switched by settings.
	PreferDevelopmentBackends bool

	// RequireBackendIntegrity is the strict-integrity switch bound to
	// --require-backend-integrity / LOCALAI_REQUIRE_BACKEND_INTEGRITY. The
	// backend install path takes it as a call argument; the gallery index
	// fetch reads it from here because the listing code it runs under never
	// sees the application config.
	RequireBackendIntegrity bool

	// DisableBackendInstall refuses every backend install and upgrade —
	// gallery, OCI/URL/path, automatic — bound to --disable-backend-install /
	// LOCALAI_DISABLE_BACKEND_INSTALL. Backends then come only from what is
	// already in BackendsPath or from BackendsSystemPath (packaged installs).
	DisableBackendInstall bool
}

type SystemStateOptions func(*SystemState)

func WithBackendPath(path string) SystemStateOptions {
	return func(s *SystemState) {
		s.Backend.BackendsPath = path
	}
}

func WithBackendSystemPath(path string) SystemStateOptions {
	return func(s *SystemState) {
		s.Backend.BackendsSystemPath = path
	}
}

func WithStagingPath(path string) SystemStateOptions {
	return func(s *SystemState) {
		s.StagingPath = path
	}
}

// BackendStagingPath returns the directory for in-flight backend downloads:
// the explicit StagingPath when configured, otherwise a "downloading"
// subdirectory inside BackendsPath so downloads stage on the same writable
// volume as installed backends without creating a root-level //downloading
// directory when basepath is "/".
func (s *SystemState) BackendStagingPath() string {
	if s.StagingPath != "" {
		return filepath.Clean(s.StagingPath)
	}
	if s.Backend.BackendsPath != "" {
		return filepath.Join(s.Backend.BackendsPath, "downloading")
	}
	return ""
}

func WithModelPath(path string) SystemStateOptions {
	return func(s *SystemState) {
		s.Model.ModelsPath = path
	}
}

func WithBackendImagesReleaseTag(tag string) SystemStateOptions {
	return func(s *SystemState) {
		s.BackendImagesReleaseTag = tag
	}
}

func WithBackendImagesBranchTag(tag string) SystemStateOptions {
	return func(s *SystemState) {
		s.BackendImagesBranchTag = tag
	}
}

func WithBackendDevSuffix(suffix string) SystemStateOptions {
	return func(s *SystemState) {
		s.BackendDevSuffix = suffix
	}
}

func WithRequireBackendIntegrity(require bool) SystemStateOptions {
	return func(s *SystemState) {
		s.RequireBackendIntegrity = require
	}
}

func WithDisableBackendInstall(disable bool) SystemStateOptions {
	return func(s *SystemState) {
		s.DisableBackendInstall = disable
	}
}

func WithPreferDevelopmentBackends(prefer bool) SystemStateOptions {
	return func(s *SystemState) {
		s.PreferDevelopmentBackends = prefer
	}
}

func GetSystemState(opts ...SystemStateOptions) (*SystemState, error) {
	state := &SystemState{}
	for _, opt := range opts {
		opt(state)
	}

	// Detection is best-effort here, we don't want to fail if it fails
	state.GPUVendor, _ = xsysinfo.DetectGPUVendor()
	xlog.Debug("GPU vendor", "gpuVendor", state.GPUVendor)
	state.VRAM, _ = xsysinfo.TotalAvailableVRAM()
	xlog.Debug("Total available VRAM", "vram", state.VRAM)

	state.getSystemCapabilities()

	return state, nil
}
