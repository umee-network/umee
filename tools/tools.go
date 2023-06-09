//go:build tools
// +build tools

// This file uses the recommended method for tracking developer tools in a Go
// module.
//
// REF: https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
package tools

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/mgechev/revive"
	_ "mvdan.cc/gofumpt"

	// unnamed import of statik for swagger UI support
	_ "github.com/umee-network/umee/v5/swagger/statik"
)
