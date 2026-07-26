package analyzer

// WalkOptions configures resource limits for directory traversal.
// Zero values are replaced by safe defaults via Normalize().
type WalkOptions struct {
	// MaxFiles is the maximum number of files to collect before stopping.
	// Default: 50000.
	MaxFiles int
	// MaxFileSize is the maximum single file size in bytes; larger files are skipped.
	// Default: 2MB (2 * 1024 * 1024).
	MaxFileSize int64
	// MaxDepth is the maximum directory depth to traverse (root = 0).
	// Default: 30.
	MaxDepth int
	// FollowSymlinks controls whether symbolic links are followed.
	// Default: false (symlinks are skipped to avoid cycles).
	FollowSymlinks bool
}

// DefaultWalkOptions returns production-safe defaults.
func DefaultWalkOptions() WalkOptions {
	return WalkOptions{
		MaxFiles:       50000,
		MaxFileSize:    2 * 1024 * 1024, // 2MB
		MaxDepth:       30,
		FollowSymlinks: false,
	}
}

// Normalize fills zero fields with defaults and clamps invalid values.
func (o WalkOptions) Normalize() WalkOptions {
	d := DefaultWalkOptions()
	if o.MaxFiles <= 0 {
		o.MaxFiles = d.MaxFiles
	}
	if o.MaxFileSize <= 0 {
		o.MaxFileSize = d.MaxFileSize
	}
	if o.MaxDepth <= 0 {
		o.MaxDepth = d.MaxDepth
	}
	return o
}

// WalkDiagnostics reports resource limit hits during traversal.
type WalkDiagnostics struct {
	// Truncated is true when MaxFiles was reached before full traversal.
	Truncated bool
	// SkippedLarge is the count of files skipped due to MaxFileSize.
	SkippedLarge int
	// SkippedSymlinks is the count of symlinks skipped.
	SkippedSymlinks int
	// MaxDepthReached is true when traversal hit the depth limit.
	MaxDepthReached bool
}
