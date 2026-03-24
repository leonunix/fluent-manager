//go:build !embed_frontend

package main

import "io/fs"

// frontendFS is nil in API-only mode — no static files served.
var frontendFS fs.FS
