package main

import (
	"os"

	"github.com/maxgio92/apkover/internal/output"
	"github.com/maxgio92/apkover/pkg/cmd"
)

func main() {
	if err := cmd.NewCmd().Execute(); err != nil {
		output.PrettyError(err)
		os.Exit(1)
	}
}
