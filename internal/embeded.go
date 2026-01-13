package internal

import "embed"

//go:embed static/*
var StaticFS embed.FS
