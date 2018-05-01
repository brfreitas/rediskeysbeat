package main

import (
	"os"

	"github.com/brfreitas/rediskeysbeat/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
