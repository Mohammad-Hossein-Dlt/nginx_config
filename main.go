package main

import (
	"bufio"
	"fmt"
	"github.com/manifoldco/promptui"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// coloredText prints the given text in a terminal color (using ANSI escape codes).
func coloredText(color, text string) {
	_, err := fmt.Fprintf(os.Stderr, "\033[%sm%s\033[0m\n", color, text)
	if err != nil {
		return
	}
}

// selectMenu displays options to the user and returns the selected option.
func selectMenu(options []string) string {
	prompt := promptui.Select{
		Label: "Select an Option",
		Items: options,
	}

	_, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed: %v\n", err)
	}

	return result
}

// findKeyByValue searches a map for a value and returns its key.
func findKeyByValue(assocMap map[string]string, searchValue string) (string, bool) {
	for key, value := range assocMap {
		if value == searchValue {
			return key, true
		}
	}
	return "", false
}

// runCommand runs an external command with given arguments.
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// certificates scans the given certificate base path for certificate files
// with extensions .crt, .pem, or .cer. It returns a map where keys are the
// base names and values are a description string.
func certificates(certBasePath string) map[string]string {
	assoc := make(map[string]string)
	var certFiles []string
	exists := []string{"*.crt", "*.pem", "*.cer"}
	for _, ext := range exists {
		matches, err := filepath.Glob(filepath.Join(certBasePath, ext))
		if err != nil {
			coloredText("31", "Error scanning certificates: "+err.Error())
			os.Exit(1)
		}
		certFiles = append(certFiles, matches...)
	}

	if len(certFiles) == 0 {
		coloredText("93", "No certificates found.")
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
		domains, err := extractDomains(cert)
		if err != nil || len(domains) == 0 {
			domains = []string{"N/A"}
		}
		domainStr := strings.Join(domains, " // ")
		assoc[baseName] = fmt.Sprintf("Cert: %s | Key: %s | Domains: %s", certFile, keyFile, domainStr)
	}
	return assoc
}

// selectCert calls certificates() and then asks the user to select one.
func selectCert(certBasePath string) string {
	names := certificates(certBasePath)
	// Build menu options from the map values.
	var options []string
	for _, detail := range names {
		options = append(options, detail)
	}
	selected := selectMenu(options)
	// Retrieve the base name (key) corresponding to the selection.
	baseName, found := findKeyByValue(names, selected)
	if !found {
		coloredText("31", "Certificate selection failed.")
		os.Exit(1)
	}
	return baseName
}

// extractDomains runs openssl to extract the Subject Alternative Names (SAN)
// from a certificate and returns a slice of domain names.
func extractDomains(certFile string) ([]string, error) {
	cmd := exec.Command("openssl", "x509", "-in", certFile, "-noout", "-ext", "subjectAltName")
	output, err := cmd.Output()
	if err != nil {
		coloredText("93", "No domains found in certificate.")
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

// promptInput shows a prompt to the user and returns the entered text.
func promptInput(prompt string) string {
	coloredText("32", prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// fileExists returns true if the file at path exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func main() {
	// Check if running as root.
	//if os.Geteuid() != 0 {
	//	coloredText("31", "Please run as root (sudo).")
	//	os.Exit(1)
	//}

	////////////////////////////////////////
	// Initialization and Fixes
	////////////////////////////////////////

	coloredText("36", "Fix the dpkg lock")
	// Kill process with PID 8001.
	_ = runCommand("kill", "8001")
	_ = runCommand("dpkg", "--configure", "-a")

	coloredText("32", "Clear cache")
	// Simulate cache clearing.
	_ = os.Remove("management.shc")
	// (BASH_REMATCH unset and kill -9 with no PID are shell specifics and are omitted.)

	////////////////////////////////////////
	// Get Certification & Setup Inputs
	////////////////////////////////////////

	certification := selectMenu([]string{"SSL", "No SSL"})
	setup := selectMenu([]string{"Default", "Websocket"})
	coloredText("36", certification)
	coloredText("36", setup)

	////////////////////////////////////////
	// Certificate Section (if needed)
	////////////////////////////////////////

	certBasePath := "/etc/ssl/files"
	// Ensure certificate directory exists.
	err := os.MkdirAll(certBasePath, 0755)
	if err != nil {
		return
	}

	// Variables for later use.
	var selectedDomain, serverIP, certPath, keyPath string
	var httpPort, httpsPort string

	if certification == "SSL" {
		// Let the user select a certificate.
		certBaseName := selectCert(certBasePath)
		certFullPath := filepath.Join(certBasePath, certBaseName+".crt")
		// Extract domains from the selected certificate.
		domains, _ := extractDomains(certFullPath)
		selectedDomain = selectMenu(domains)
		certPath = certFullPath
		keyFullPath := filepath.Join(certBasePath, certBaseName+".key")
		if !fileExists(keyFullPath) {
			coloredText("93", "Key file not found for certificate.")
			os.Exit(1)
		}
		keyPath = keyFullPath

		httpPort = promptInput("Please enter http port (80 is default):")
		if httpPort == "" {
			httpPort = "80"
		}
		httpsPort = promptInput("Please enter https port (443 is default):")
		if httpsPort == "" {
			httpsPort = "443"
		}
	} else { // No SSL
		serverIP = promptInput("Please enter the ip of this server:")
		selectedDomain = serverIP
		httpPort = promptInput("Please enter http port (80 is default):")
		if httpPort == "" {
			httpPort = "80"
		}
	}

	////////////////////////////////////////
	// Nginx Configuration
	////////////////////////////////////////

	configsBasePath := "/etc/nginx/conf.d"
	// List previous configuration files.
	coloredText("36", "Please enter a unique name for config file. Previous configs are shown below:")
	existingConfigs, _ := filepath.Glob(filepath.Join(configsBasePath, "*.conf"))
	for _, cfg := range existingConfigs {
		fmt.Println(cfg)
	}
	configName := promptInput("")
	configFilePath := filepath.Join(configsBasePath, configName+".conf")
	if fileExists(configFilePath) {
		coloredText("93", "The name you entered already exists.")
		os.Exit(1)
	}

	upstreamIPs := promptInput("Please enter the list of upstream IP addresses (space separated):")
	upstreamArray := strings.Fields(upstreamIPs)
	var upstreamConf strings.Builder
	upstreamConf.WriteString(fmt.Sprintf("upstream %s {", configName))
	if setup == "Websocket" {
		upstreamConf.WriteString("\n    ip_hash;")
	}
	for _, ip := range upstreamArray {
		upstreamConf.WriteString(fmt.Sprintf("\n    server %s;", ip))
	}
	upstreamConf.WriteString("\n}")

	var configContent string
	// Build configuration based on the chosen options.
	if certification == "SSL" && setup == "Default" {
		config := `
			# Define an upstream block for the backend server(s)
			%s
			
			# HTTP block: Redirect all HTTP traffic to HTTPS
			server {
				listen %s;
				server_name %s;
				return 301 https://$host$request_uri;
			}
			
			# HTTPS block: SSL configuration and reverse proxy settings
			server {
				listen %s ssl;
				server_name %s;
			
				ssl_certificate %s;
				ssl_certificate_key %s;
			
				ssl_protocols TLSv1.2 TLSv1.3;
				ssl_ciphers HIGH:!aNULL:!MD5;
			
				location / {
					proxy_pass http://%s;
			
					proxy_set_header Host $host;
			
					proxy_set_header X-Real-IP $remote_addr;
					proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
					proxy_set_header X-Forwarded-Proto $scheme;
				}
			}
		`
		configContent = fmt.Sprintf(config, upstreamConf.String(), httpPort, selectedDomain, httpsPort, selectedDomain, certPath, keyPath, configName)
	} else if certification == "SSL" && setup == "Websocket" {
		config := `
			# Define an upstream block for the backend server(s)
			%s
			
			# HTTP block: Redirect all HTTP traffic to HTTPS
			server {
			listen %s;
			server_name %s;
			return 301 https://$host$request_uri;
			}
			
			# HTTPS block: SSL configuration and reverse proxy settings
			server {
			listen %s ssl;
			server_name %s;
			
			ssl_certificate %s;
			ssl_certificate_key %s;
			
			ssl_protocols TLSv1.2 TLSv1.3;
			ssl_ciphers HIGH:!aNULL:!MD5;
			
			location / {
			proxy_pass http://%s;
			
			proxy_http_version 1.1;
			proxy_set_header Upgrade $http_upgrade;
			proxy_set_header Connection "upgrade";
			
			proxy_set_header Host $host;
			
			proxy_set_header X-Real-IP $remote_addr;
			proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
			proxy_set_header X-Forwarded-Proto $scheme;
			}
			}
		`
		configContent = fmt.Sprintf(config, upstreamConf.String(), httpPort, selectedDomain, httpsPort, selectedDomain, certPath, keyPath, configName)
	} else if certification == "No SSL" && setup == "Default" {
		config := `
			%s
			
			server {
				listen %s;
				server_name %s;
			
				location / {
					proxy_pass http://%s;
					proxy_set_header Host $host;
					proxy_set_header X-Real-IP $remote_addr;
					proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
					proxy_set_header X-Forwarded-Proto $scheme;
				}
			}
		`
		configContent = fmt.Sprintf(config, upstreamConf.String(), httpPort, selectedDomain, configName)
	} else if certification == "No SSL" && setup == "Websocket" {
		config := `
			%s
			
			server {
				listen %s;
				server_name %s;
			
				location / {
					proxy_pass http://%s;
			
					proxy_http_version 1.1;
					proxy_set_header Upgrade $http_upgrade;
					proxy_set_header Connection "upgrade";
			
					proxy_set_header Host $host;
			
					proxy_set_header X-Real-IP $remote_addr;
					proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
					proxy_set_header X-Forwarded-Proto $scheme;
				}
			}
		`
		configContent = fmt.Sprintf(config, upstreamConf.String(), httpPort, selectedDomain, configName)
	}

	// Write the configuration file.
	err = os.WriteFile(configFilePath, []byte(configContent), 0644)
	if err != nil {
		coloredText("31", "Error writing config file: "+err.Error())
		os.Exit(1)
	}
	coloredText("32", "Creating configuration file for load balancer and reverse proxy:")

	// Remove default configuration files if they exist.
	defaultFiles := []string{"/etc/nginx/sites-enabled/default", "/etc/nginx/conf.d/default.conf"}
	for _, df := range defaultFiles {
		if fileExists(df) {
			coloredText("32", "Removing default configuration at "+df)
			_ = os.Remove(df)
		}
	}

	// Test the nginx configuration.
	coloredText("32", "Testing nginx configuration...")
	if err := runCommand("nginx", "-t"); err != nil {
		coloredText("31", "Error in nginx configuration. Please check the config files.")
		os.Exit(1)
	}

	// Reload and enable nginx.
	coloredText("32", "Reloading nginx...")
	_ = runCommand("systemctl", "reload", "nginx")
	coloredText("32", "Enabling nginx service to automatically start after reboot...")
	_ = runCommand("systemctl", "enable", "nginx")

	coloredText("36", "Reverse proxy and Load balancer installation and configuration completed successfully.")

	////////////////////////////////////////
	// Firewall (ufw) Setup
	////////////////////////////////////////

	coloredText("32", "Allowing SSH on port 22 and web traffic on ports 80, 443...")
	_ = runCommand("ufw", "allow", "9011/tcp")
	_ = runCommand("ufw", "allow", "22/tcp")
	_ = runCommand("ufw", "allow", "80/tcp")
	_ = runCommand("ufw", "allow", "443/tcp")

	// Enable ufw (this may prompt for confirmation).
	_ = runCommand("ufw", "--force", "enable")

	coloredText("36", "All is done.")
}
