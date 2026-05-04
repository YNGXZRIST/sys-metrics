// Package internal exposes the embedded static filesystem for the server.
package internal

import "embed"

//go:embed static/*

// StaticFS is the embedded filesystem with static assets (HTML/CSS) for the server index page.
var StaticFS embed.FS
