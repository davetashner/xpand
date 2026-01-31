package main

import (
	"fmt"
	"os"
)

// Version information (set by build)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("xpand %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	fmt.Println("xpand - Expand application intent into explicit Kubernetes resources")
	fmt.Println()
	fmt.Println("Status: Pre-alpha (design phase)")
	fmt.Println("See: https://github.com/davetashner/xpand")
	fmt.Println()
	fmt.Println("Commands will be implemented after ADRs are decided.")
}
