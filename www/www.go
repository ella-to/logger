package www

import "embed"

//go:embed dist/*
var Files embed.FS
