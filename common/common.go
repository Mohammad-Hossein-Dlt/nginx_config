package common

import (
	"bufio"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Color string

const (
	White Color = "#FCFCFC"
	Gold        = "#FFD700"
	Blue        = "#1E90FF"
	Green       = "#32CD32"
	Red         = "#FF4500"
)

// FixDpkgLock attempts to kill process 8001 and run dpkg --configure -a.
//func FixDpkgLock() error {
//	ColoredText("36", "Fix the dpkg lock")
//	if err := RunCommand("sudo", "kill", "8001"); err != nil {
//		ColoredText("31", "Failed to kill process 8001: "+err.Error())
//	}
//	return RunCommand("dpkg", "--configure", "-a")
//}

// ClearCache mimics cache clearing by removing a file.
func ClearCache() {
	ColoredText("32", "Clear cache")
	_ = os.Remove("management.shc")
	// Note: hash -r, unset BASH_REMATCH, and a bare kill -9 have no direct equivalents in Go.
}

// ColoredText prints the given text in a terminal color (using ANSI escape codes).
func ColoredText(color, text string) {
	tea.Printf("\033[%sm%s\033[0m\n", color, text)
	//if err != nil {
	//	return
	//}
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

func GetValues(assocMap map[string]string) []string {
	var assoc []string
	for _, value := range assocMap {
		assoc = append(assoc, value)
	}
	return assoc
}

func ExtractKeys(assocMap map[string]string) []string {
	var assoc []string
	for k, _ := range assocMap {
		assoc = append(assoc, k)
	}
	return assoc
}

func GetKeyByIndex(assocMap map[string]string, index int) string {
	keys := ExtractKeys(assocMap)

	for i, k := range keys {
		if i == index {
			return k
		}
	}
	return ""
}

type LogMsg struct {
	Msg   string
	Color Color
}

func sendLog(ch chan<- LogMsg, pipe io.ReadCloser, isError bool) {
	newScanner := bufio.NewScanner(pipe)
	for newScanner.Scan() {
		ch <- LogMsg{Msg: newScanner.Text(), Color: White}
	}
}

// RunCommand runs an external command with given arguments.
func RunCommand(cmdStr string, ch chan<- LogMsg) {
	cmd := exec.Command("bash", "-c", cmdStr)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	_ = cmd.Start()

	go sendLog(ch, stdout, false)
	go sendLog(ch, stderr, true)

	_ = cmd.Wait()
}

// Certificates scans the given certificate base path for certificate files
// with extensions .crt, .pem, or .cer. It returns a map where keys are the
// base names and values are a description string.
func Certificates(certBasePath string) (map[string]string, LogMsg) {
	assoc := make(map[string]string)
	var certFiles []string
	exists := []string{"*.crt", "*.pem", "*.cer"}
	for _, ext := range exists {
		matches, err := filepath.Glob(filepath.Join(certBasePath, ext))
		if err != nil {
			return nil, LogMsg{Msg: "Error scanning certificates: " + err.Error(), Color: Red}
		}
		certFiles = append(certFiles, matches...)
	}

	if len(certFiles) == 0 {
		return nil, LogMsg{Msg: "No certificates found.", Color: Gold}
		//os.Exit(1)
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
		domainStr := strings.Join(domains, ", ")
		assoc[baseName] = fmt.Sprintf("Cert: %s | Key: %s | Domains: %s", certFile, keyFile, domainStr)
	}
	return assoc, LogMsg{Msg: ""}
}

func ExtractDomains(certPath string) ([]string, error) {
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		return nil, fmt.Errorf("error decoding PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing certificate: %v", err)
	}

	var domains []string

	domains = append(domains, cert.DNSNames...)

	return domains, nil
}

// FileExists returns true if the file at path exists.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
