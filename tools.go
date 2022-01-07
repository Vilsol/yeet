//go:build tools
// +build tools

package main

import (
	"github.com/Vilsol/yeet/cmd"
	"github.com/spf13/cobra/doc"

	// Import sub-commands
	_ "github.com/Vilsol/yeet/cmd/serve"
)

//go:generate go run -tags tools tools.go
//go:generate flatc --go -o flat flat/s3.fbs

func main() {
	err := doc.GenMarkdownTree(cmd.RootCMD, "./docs/")
	if err != nil {
		panic(err)
	}
}
