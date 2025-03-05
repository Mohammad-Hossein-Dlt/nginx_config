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
)

type Color string

const (
	White Color = "#FCFCFC"
	Gold        = "#FFD700"
	Blue        = "#1E90FF"
	Green       = "#32CD32"
	//Red         = "#E63946"
	Red = "#FF6F61"
)

// FixDpkgLock attempts to kill process 8001 and run dpkg --configure -a.
func FixDpkgLock() error {
	ColoredText("36", "Fix the dpkg lock")
	if err := RunCommand("sudo kill 8001"); err != nil {
		ColoredText("31", "Failed to kill process 8001: "+err.Error())
	}
	return RunCommand("dpkg --configure -a")
}

// ClearCache mimics cache clearing by removing a file.
func ClearCache() {
	ColoredText("32", "Clear cache")
	_ = os.Remove("management.shc")
}

// ColoredText prints the given text in a terminal color (using ANSI escape codes).
func ColoredText(color, text string) {
	tea.Printf("\033[%sm%s\033[0m\n", color, text)
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

type LogData struct {
	Messages []LogItem
}

type LogItem struct {
	Msg   string
	Color Color
}

func CreateLogItems(logs []string, color Color) []LogItem {
	var items []LogItem
	for _, log := range logs {
		items = append(items, LogItem{Msg: log, Color: color})
	}
	return items
}

func CreateSingleLog(msg string, color Color) LogData {
	return LogData{
		Messages: []LogItem{
			{Msg: msg, Color: color},
		},
	}
}

func LogMessage(msg string, color Color) tea.Cmd {
	return func() tea.Msg {
		return LogData{
			Messages: []LogItem{
				{Msg: msg, Color: color},
			},
		}
	}
}

func RunCommandWithLogs(cmd string) tea.Cmd {
	return func() tea.Msg {
		command := exec.Command("bash", "-c", cmd)

		// Get output pipes for both stdout and stderr
		stdout, err := command.StdoutPipe()
		if err != nil {
			return CreateSingleLog(fmt.Sprintf("❌ Error setting up StdoutPipe: %v", err), Red)
		}
		stderr, err2 := command.StderrPipe()
		if err2 != nil {
			return CreateSingleLog(fmt.Sprintf("❌ Error setting up StderrPipe: %v", err2), Red)
		}

		// Start the command
		if err3 := command.Start(); err3 != nil {
			return CreateSingleLog(fmt.Sprintf("❌ Error executing command %s: %v", cmd, err3), Red)
		}

		// Read stdout and stderr

		var data []string

		stdoutScanner := bufio.NewScanner(stdout)
		for stdoutScanner.Scan() {
			data = append(data, stdoutScanner.Text())
		}

		stderrScanner := bufio.NewScanner(stderr)
		for stderrScanner.Scan() {
			data = append(data, stderrScanner.Text())
		}

		_ = command.Wait()

		// Return a final message when done
		return LogData{Messages: CreateLogItems(data, White)}
	}
}

func RunCommand(cmd string) error {
	command := exec.Command("bash", "-c", cmd)
	if err := command.Run(); err != nil {
		return err
	}
	return nil
}

// Certificates scans the given certificate base path for certificate files
// with extensions .crt, .pem, or .cer. It returns a map where keys are the
// base names and values are a description string.
func Certificates(certBasePath string) (map[string]string, LogData) {
	assoc := make(map[string]string)
	var certFiles []string
	exists := []string{"*.crt", "*.pem", "*.cer"}
	for _, ext := range exists {
		matches, err := filepath.Glob(filepath.Join(certBasePath, ext))
		if err != nil {
			return nil, CreateSingleLog(fmt.Sprintf("Error scanning certificates: %v", err), Red)
		}
		certFiles = append(certFiles, matches...)
	}

	if len(certFiles) == 0 {
		return nil, CreateSingleLog("No certificates found.", Gold)
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
	return assoc, LogData{}
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

//----------------------------------------------------------------------------------------------------------------------

// --------------------
// Nginx Management
// --------------------

// getConfigs returns the list of config file names in ConfigsBasePath.
func getConfigs(configsBasePath string) ([]string, error) {
	var configs []string
	err := filepath.Walk(configsBasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".conf") {
			configs = append(configs, info.Name())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(configs) == 0 {
		ColoredText("93", "No configs found.")
		os.Exit(1)
	}
	return configs, nil
}

// deleteConfig deletes a config file from ConfigsBasePath.
func deleteConfig(configsBasePath string, fileName string) {
	if fileName != "" {
		path := filepath.Join(configsBasePath, fileName)
		_ = os.Remove(path)
		ColoredText("32", fmt.Sprintf("Config %s deleted.", fileName))
	} else {
		ColoredText("31", "Cannot delete directory conf.d in path /etc/nginx/conf.d")
	}
}

// editConfig opens a config file in nano.
func editConfig(configsBasePath string, fileName string) {
	path := filepath.Join(configsBasePath, fileName)
	_ = RunCommand(fmt.Sprintf("nano %s", path))
	ColoredText("32", fmt.Sprintf("Config %s edited.", fileName))
}

// --------------------
// Firewall Management
// --------------------

// firewallStatus displays the ufw status.
func firewallStatus() {
	_ = RunCommand("ufw status")
}

// installFirewall installs ufw and opens some default ports.
func installFirewall() {
	ColoredText("32", "Updating package list...")
	_ = RunCommand("apt-get update -y")
	ColoredText("32", "Installing firewall...")
	_ = RunCommand("apt-get install -y ufw")
	if _, err := exec.LookPath("ufw"); err == nil {
		_ = RunCommand("ufw allow 9011/tcp")
		_ = RunCommand("ufw allow 22/tcp")
	}
}

// uninstallFirewall disables ufw, purges it, and removes its directory.
func uninstallFirewall() {
	if _, err := exec.LookPath("ufw"); err == nil {
		ColoredText("32", "Firewall is installed. Purging existing installation and configuration files...")
		ColoredText("32", "Stopping ufw service...")
		_ = RunCommand("ufw disable")
		ColoredText("32", "Purging firewall ufw...")
		_ = RunCommand("apt-get purge -y ufw")
		ColoredText("32", "Auto removing packages...")
		_ = RunCommand("apt-get autoremove -y ufw")
		ColoredText("32", "Removing ufw directory...")
		err := os.RemoveAll("/etc/ufw")
		if err != nil {
			ColoredText("31", "An Error occurred : "+err.Error())
			os.Exit(1)
		}
	}
}

// openingPorts prompts for ports and opens them in ufw.
func openingPorts() {
	reader := bufio.NewReader(os.Stdin)
	ColoredText("32", "Enter ports to open (separated by space):")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	ports := strings.Split(input, " ")
	for _, port := range ports {
		ColoredText("32", fmt.Sprintf("Opening port: %s", port))
		_ = RunCommand(fmt.Sprintf("ufw allow %s/tcp", port))
	}
	ColoredText("32", "Check firewall status")
	_ = RunCommand("ufw status")
}

// --------------------
// Certificate Management
// --------------------

const CertBasePath = "/etc/ssl/files"

// ensureCertBasePath creates CertBasePath if it does not exist.
func ensureCertBasePath() {
	err := os.MkdirAll(CertBasePath, 0755)
	if err != nil {
		ColoredText("31", "An Error occurred : "+err.Error())
		os.Exit(1)
	}

}

// getCertificates scans CertBasePath for certificate files.
func getCertificates() (map[string]string, error) {
	certs := make(map[string]string)
	files, err := os.ReadDir(CertBasePath)
	if err != nil {
		return nil, err
	}
	found := false
	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".crt") ||
			strings.HasSuffix(file.Name(), ".pem") ||
			strings.HasSuffix(file.Name(), ".cer")) {
			found = true
			certFile := file.Name()
			baseName := strings.TrimSuffix(certFile, filepath.Ext(certFile))
			keyPath := filepath.Join(CertBasePath, baseName+".key")
			keyFile := baseName + ".key"
			if _, err := os.Stat(keyPath); os.IsNotExist(err) {
				keyFile = "N/A"
			}
			// Extract domains using openssl (a simple extraction)
			domainsCmd := exec.Command("openssl", "x509", "-in", filepath.Join(CertBasePath, certFile), "-noout", "-ext", "subjectAltName")
			output, err := domainsCmd.Output()
			domains := "N/A"
			if err == nil {
				outStr := string(output)
				parts := strings.Split(outStr, "DNS:")
				if len(parts) > 1 {
					domainPart := strings.Split(parts[1], ",")[0]
					domains = strings.TrimSpace(domainPart)
				}
			}
			certs[filepath.Join(CertBasePath, certFile)] = fmt.Sprintf("Cert: %s | Key: %s | Domains: %s", certFile, keyFile, domains)
		}
	}
	if !found {
		ColoredText("93", "No certificates found.")
		os.Exit(1)
	}
	return certs, nil
}

// certificateInfo displays details of a certificate using openssl.
func certificateInfo(certPath string) {
	certFile := filepath.Base(certPath)
	baseName := strings.TrimSuffix(certFile, filepath.Ext(certFile))
	keyPath := filepath.Join(CertBasePath, baseName+".key")
	keyFile := baseName + ".key"
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		keyFile = "N/A"
	}
	domainsCmd := exec.Command("openssl", "x509", "-in", certPath, "-noout", "-ext", "subjectAltName")
	output, err := domainsCmd.Output()
	domains := "N/A"
	if err == nil {
		outStr := string(output)
		parts := strings.Split(outStr, "DNS:")
		if len(parts) > 1 {
			domainPart := strings.Split(parts[1], ",")[0]
			domains = strings.TrimSpace(domainPart)
		}
	}
	datesCmd := exec.Command("openssl", "x509", "-in", certPath, "-noout", "-dates")
	datesOut, err := datesCmd.Output()
	notBefore, notAfter := "N/A", "N/A"
	if err == nil {
		lines := strings.Split(string(datesOut), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "notBefore=") {
				notBefore = strings.TrimPrefix(line, "notBefore=")
			} else if strings.HasPrefix(line, "notAfter=") {
				notAfter = strings.TrimPrefix(line, "notAfter=")
			}
		}
	}
	info := fmt.Sprintf("Cert: %s\nKey: %s\nDomains: %s\nValid from: %s\nValid to: %s", certFile, keyFile, domains, notBefore, notAfter)
	ColoredText("36", info)
}

// getCert prompts the user to enter certificate content using nano and writes it to file.
func getCert(name string) (string, error) {
	tmpFile, err := os.CreateTemp("", "cert_*.crt")
	if err != nil {
		ColoredText("31", "An Error occurred : "+err.Error())
		os.Exit(1)
	}
	tmpFilePath := tmpFile.Name()
	err2 := tmpFile.Close()
	if err2 != nil {
		ColoredText("31", "An Error occurred : "+err2.Error())
		os.Exit(1)
	}
	ColoredText("36", "Please enter your certificate content in nano. Save and exit when done.")
	if err := RunCommand(fmt.Sprintf("nano %s", tmpFilePath)); err != nil {
		return "", err
	}
	content, err := os.ReadFile(tmpFilePath)
	if err != nil {
		return "", err
	}
	err3 := os.Remove(tmpFilePath)
	if err3 != nil {
		ColoredText("31", "An Error occurred : "+err3.Error())
		os.Exit(1)
	}
	destPath := filepath.Join(CertBasePath, name+".crt")
	if err := os.WriteFile(destPath, content, 0644); err != nil {
		return "", err
	}
	return destPath, nil
}

// getKey prompts the user to enter private key content using nano and writes it to file.
func getKey(name string) (string, error) {
	tmpFile, err := os.CreateTemp("", "key_*.key")
	if err != nil {
		return "", err
	}
	tmpFilePath := tmpFile.Name()
	err2 := tmpFile.Close()
	if err2 != nil {
		ColoredText("31", "An Error occurred : "+err2.Error())
		os.Exit(1)
	}
	ColoredText("36", "Please enter your private key content in nano. Save and exit when done.")
	if err := RunCommand(fmt.Sprintf("nano %s", tmpFilePath)); err != nil {
		return "", err
	}
	content, err := os.ReadFile(tmpFilePath)
	if err != nil {
		return "", err
	}
	err3 := os.Remove(tmpFilePath)
	if err3 != nil {
		ColoredText("31", "An Error occurred : "+err3.Error())
		os.Exit(1)
	}
	destPath := filepath.Join(CertBasePath, name+".key")
	if err := os.WriteFile(destPath, content, 0644); err != nil {
		return "", err
	}
	return destPath, nil
}

// deleteCertificate removes the certificate and its key.
func deleteCertificate(certPath string) {
	certFile := filepath.Base(certPath)
	baseName := strings.TrimSuffix(certFile, filepath.Ext(certFile))
	ColoredText("32", fmt.Sprintf("Removing ssl certificate '%s'", baseName))
	err := os.Remove(filepath.Join(CertBasePath, baseName+".crt"))
	if err != nil {
		ColoredText("31", "An Error occurred : "+err.Error())
		os.Exit(1)
	}
	err2 := os.Remove(filepath.Join(CertBasePath, baseName+".key"))
	if err2 != nil {
		ColoredText("31", "An Error occurred : "+err2.Error())
		os.Exit(1)
	}
}

// deleteAllCertificates removes all certificate and key files from CertBasePath.
func deleteAllCertificates() {
	ColoredText("32", "Removing all ssl certificates...")
	if files, err := filepath.Glob(filepath.Join(CertBasePath, "*.crt")); err == nil {
		for _, f := range files {
			err := os.Remove(f)
			if err != nil {
				ColoredText("31", "An Error occurred : "+err.Error())
				os.Exit(1)
			}
		}
	}
	if files, err := filepath.Glob(filepath.Join(CertBasePath, "*.key")); err == nil {
		for _, f := range files {
			err := os.Remove(f)
			if err != nil {
				ColoredText("31", "An Error occurred : "+err.Error())
				os.Exit(1)
			}
		}
	}
}

// --------------------
// Install/Uninstall Requirements
// --------------------

// installRequirements installs both nginx and ufw.
func installRequirements() {
	ColoredText("32", "Updating package list...")
	_ = RunCommand("apt-get update -y")
	ColoredText("32", "Installing nginx...")
	_ = RunCommand("apt-get install nginx -y")
	ColoredText("32", "Installing firewall...")
	_ = RunCommand("apt-get install -y ufw")
}

// uninstallEverything calls the uninstallation routines.
func uninstallEverything() {
	uninstallFirewall()
	deleteAllCertificates()
}

// confirmPrompt asks for a yes/no confirmation.
func confirmPrompt(message string) bool {
	ColoredText("94", message+" (yes/y to confirm, no/n to cancel):")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))
	return input == "yes" || input == "y"
}
