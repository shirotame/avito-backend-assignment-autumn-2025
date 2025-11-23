package api

import "embed"

//go:embed openapi.yml
var SpecFile embed.FS
