package swagger

import "embed"

//go:embed swagger.yaml
var Docs embed.FS
