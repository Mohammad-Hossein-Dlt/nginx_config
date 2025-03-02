package main

import (
	"nginx_configure/tui"
	"os"
)

func main() {

	certBasePath := "/etc/ssl/files"
	// Ensure certificate directory exists.
	err := os.MkdirAll(certBasePath, 0755)
	if err != nil {
		return
	}

	tui.RunApp()
}
