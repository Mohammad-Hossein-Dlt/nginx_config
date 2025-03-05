package main

import (
	"nginx_configure/common"
	"nginx_configure/tui"
	"os"
)

func main() {
	// Check if running as root.

	if os.Geteuid() != 0 {
		common.ColoredText("31", "Please run as root (sudo).")
		os.Exit(1)
	}

	certBasePath := "/etc/ssl/files"
	// Ensure certificate directory exists.
	err := os.MkdirAll(certBasePath, 0755)
	if err != nil {
		return
	}

	tui.RunApp()
}
