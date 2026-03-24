//go:build embed_frontend

package main

import (
	"embed"
	"io/fs"
	"log"
)

// frontend_dist is copied from web/frontend/dist by the Makefile before building.
//
//go:embed all:frontend_dist
var embeddedFrontend embed.FS

// frontendFS exposes the embedded frontend to the router.
var frontendFS fs.FS

func init() {
	sub, err := fs.Sub(embeddedFrontend, "frontend_dist")
	if err != nil {
		log.Fatalf("Failed to access embedded frontend: %v", err)
	}
	frontendFS = sub
}
