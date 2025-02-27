package common

import (
	"bufio"
	"fmt"
	"github.com/manifoldco/promptui"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FixDpkgLock attempts to kill process 8001 and run dpkg --configure -a.
func FixDpkgLock() error {
	ColoredText("36", "Fix the dpkg lock")
	if err := RunCommand("sudo", "kill", "8001"); err != nil {
		ColoredText("31", "Failed to kill process 8001: "+err.Error())
	}
	return RunCommand("dpkg", "--configure", "-a")
}

// ClearCache mimics cache clearing by removing a file.
func ClearCache() {
	ColoredText("32", "Clear cache")
	_ = os.Remove("management.shc")
	// Note: hash -r, unset BASH_REMATCH, and a bare kill -9 have no direct equivalents in Go.
}

// ColoredText prints the given text in a terminal color (using ANSI escape codes).
func ColoredText(color, text string) {
	_, err := fmt.Fprintf(os.Stderr, "\033[%sm%s\033[0m\n", color, text)
	if err != nil {
		return
	}
}

// SelectMenu displays options to the user and returns the selected option.
func SelectMenu(options []string) string {
	prompt := promptui.Select{
		Label: "Select an Option",
		Items: options,
		Size:  len(options),
	}

	_, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed: %v\n", err)
	}

	//tui.ListUi(
	//	[]tui.NestedItem{
	//		{
	//			Title:       "test1",
	//			Description: "aaa",
	//			Children: []tui.NestedItem{
	//				{Title: "test2", Description: ""},
	//				{Title: "test3", Description: ""},
	//			},
	//			Action: func() {
	//				fmt.Println("Action for Child 1.1 executed!")
	//			},
	//		},
	//		{
	//			Title:       "test4",
	//			Description: "bbb",
	//			Children: []tui.NestedItem{
	//				{Title: "test5", Description: ""},
	//				{Title: "test6", Description: ""},
	//			},
	//		},
	//	},
	//)
	return result
}

// FindKeyByValue searches a map for a value and returns its key.
func FindKeyByValue(assocMap map[string]string, searchValue string) (string, bool) {
	for key, value := range assocMap {
		if value == searchValue {
			return key, true
		}
	}
	return "", false
}

// RunCommand runs an external command with given arguments.
func RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Certificates scans the given certificate base path for certificate files
// with extensions .crt, .pem, or .cer. It returns a map where keys are the
// base names and values are a description string.
func Certificates(certBasePath string) map[string]string {
	assoc := make(map[string]string)
	var certFiles []string
	exists := []string{"*.crt", "*.pem", "*.cer"}
	for _, ext := range exists {
		matches, err := filepath.Glob(filepath.Join(certBasePath, ext))
		if err != nil {
			ColoredText("31", "Error scanning certificates: "+err.Error())
			os.Exit(1)
		}
		certFiles = append(certFiles, matches...)
	}

	if len(certFiles) == 0 {
		ColoredText("93", "No certificates found.")
		os.Exit(1)
	}

	for _, cert := range certFiles {
		certFile := filepath.Base(cert)
		baseName := strings.TrimSuffix(certFile, filepath.Ext(certFile))
		keyPath := filepath.Join(certBasePath, baseName+".key")
		keyFile := baseName + ".key"
		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			keyFile = "N/A"
		}

		// Extract domains from the certificate using openssl.
		domains, err := ExtractDomains(cert)
		if err != nil || len(domains) == 0 {
			domains = []string{"N/A"}
		}
		domainStr := strings.Join(domains, " // ")
		assoc[baseName] = fmt.Sprintf("Cert: %s | Key: %s | Domains: %s", certFile, keyFile, domainStr)
	}
	return assoc
}

// SelectCert calls certificates() and then asks the user to select one.
func SelectCert(certBasePath string) string {
	names := Certificates(certBasePath)
	// Build menu options from the map values.
	var options []string
	for _, detail := range names {
		options = append(options, detail)
	}
	selected := SelectMenu(options)
	// Retrieve the base name (key) corresponding to the selection.
	baseName, found := FindKeyByValue(names, selected)
	if !found {
		ColoredText("31", "Certificate selection failed.")
		os.Exit(1)
	}
	return baseName
}

// ExtractDomains runs openssl to extract the Subject Alternative Names (SAN)
// from a certificate and returns a slice of domain names.
func ExtractDomains(certFile string) ([]string, error) {
	cmd := exec.Command("openssl", "x509", "-in", certFile, "-noout", "-ext", "subjectAltName")
	output, err := cmd.Output()
	if err != nil {
		ColoredText("93", "No domains found in certificate.")
		os.Exit(1)
		return nil, err
	}
	var domains []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "DNS:") {
			// The line might be like: "DNS:example.com, DNS:www.example.com"
			parts := strings.Split(line, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "DNS:") {
					domain := strings.TrimPrefix(part, "DNS:")
					domain = strings.TrimSpace(domain)
					if domain != "" {
						domains = append(domains, domain)
					}
				}
			}
		}
	}
	return domains, nil
}

// PromptInput shows a prompt to the user and returns the entered text.
func PromptInput(prompt string) string {
	ColoredText("32", prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// FileExists returns true if the file at path exists.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
