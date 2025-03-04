package common

import (
	"bufio"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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

//func RunCommand(cmd string) LogMsg {
//
//	command := exec.Command("bash", "-c", cmd)
//
//	stdout, _ := command.StdoutPipe()
//	stderr, _ := command.StderrPipe()
//
//	if err := command.Start(); err != nil {
//		//tea.Println(err)
//		return LogMsg{Log: fmt.Sprintf("❌ Error executing %s: %v", cmd, err)}
//	}
//
//	logChan := make(chan string)
//
//	go func() {
//		scanner := bufio.NewScanner(stdout)
//		for scanner.Scan() {
//			logChan <- scanner.Text()
//		}
//		close(logChan)
//	}()
//
//	go func() {
//		scanner := bufio.NewScanner(stderr)
//		for scanner.Scan() {
//			logChan <- scanner.Text()
//		}
//	}()
//
//	for log := range logChan {
//		time.Sleep(100 * time.Millisecond)
//		return LogMsg{Log: log}
//	}
//
//	_ = command.Wait()
//	return LogMsg{Log: fmt.Sprintf("✅ Command `%s` executed.", cmd)}
//
//}

func RunCommand(cmd string, args ...string) tea.Cmd {
	return func() tea.Msg {
		command := exec.Command("bash", "-c", cmd)

		// Get output pipes for both stdout and stderr
		stdout, err := command.StdoutPipe()
		if err != nil {
			return LogMsg{Msg: fmt.Sprintf("❌ Error setting up StdoutPipe: %v", err)}
		}
		stderr, err := command.StderrPipe()
		if err != nil {
			return LogMsg{Msg: fmt.Sprintf("❌ Error setting up StderrPipe: %v", err)}
		}

		// Start the command
		if err := command.Start(); err != nil {
			return LogMsg{Msg: fmt.Sprintf("❌ Error executing command %s: %v", cmd, err)}
		}

		// Use a channel to capture all output logs
		logChan := make(chan string)

		// Read stdout and stderr
		go func() {
			stdoutScanner := bufio.NewScanner(stdout)
			for stdoutScanner.Scan() {
				logChan <- stdoutScanner.Text()
			}
		}()

		go func() {
			stderrScanner := bufio.NewScanner(stderr)
			for stderrScanner.Scan() {
				logChan <- stderrScanner.Text()
			}
		}()

		// Send the logs to the UI (live)
		for log := range logChan {
			time.Sleep(100 * time.Millisecond) // Slight delay for smoother output
			return LogMsg{Msg: log}
		}

		// Wait for command to finish execution
		_ = command.Wait()
		close(logChan)

		// Return final message when done
		return LogMsg{Msg: fmt.Sprintf("✅ Command `%s` executed successfully.", cmd)}
	}
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
