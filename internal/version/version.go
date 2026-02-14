package version

// Version is the current GDF version.
// This is set at build time via -ldflags.
var Version = "0.6.0-dev"

// Commit is the git commit hash.
var Commit = "none"

// Date is the build date.
var Date = "unknown"
