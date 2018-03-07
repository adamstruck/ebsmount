package main

import (
	"fmt"
	"os"

	"github.com/adamstruck/ebsmount/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
}
