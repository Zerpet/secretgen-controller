// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"carvel.dev/secretgen-controller/internal/cli/cmd"
)

// Version is set via -ldflags at build time.
var Version = "develop"

func main() {
	cmd.Execute(Version)
}
