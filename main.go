package main

import (
	"github.com/Vilsol/yeet/cmd"

	// Import sub-commands
	_ "github.com/Vilsol/yeet/cmd/serve"
)

func main() {
	cmd.Execute()
}
