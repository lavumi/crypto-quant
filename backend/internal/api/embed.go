package api

import (
	"embed"
	"io/fs"
)

// Embed frontend static files into the binary
// This will be populated during build with the compiled frontend
//
//go:embed frontend/build/*
var frontendFS embed.FS

// GetFrontendFS returns the embedded frontend filesystem
// The files are located in the "frontend/build" directory within the embedded FS
func GetFrontendFS() (fs.FS, error) {
	return fs.Sub(frontendFS, "frontend/build")
}





