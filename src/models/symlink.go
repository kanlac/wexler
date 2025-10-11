package models

// SymlinkInfo captures metadata about a symlink that Mindful needs to manage.
type SymlinkInfo struct {
	LinkPath    string `json:"link_path"`    // The path of the symlink (project-relative when possible)
	TargetPath  string `json:"target_path"`  // The target path of the symlink (project-relative when possible)
	IsValid     bool   `json:"is_valid"`     // True when an existing symlink already points to the target
	IsDirectory bool   `json:"is_directory"` // Indicates whether the target is a directory symlink
}
