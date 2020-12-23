// Read about tools here: https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
// +build tools

package tools

import (
	_ "github.com/ory/go-acc"
	_ "github.com/quasilyte/go-consistent"
	_ "mvdan.cc/gofumpt/gofumports"
)
