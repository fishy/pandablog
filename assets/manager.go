package assets

import "embed"

//go:embed css/* cactus/* webmention.js/*
var Assets embed.FS
