package configure

import (
	"fmt"
	"nginx_configure/common"
	"os"
	"path/filepath"
	"strings"
)

func Configure() {
	// Check if running as root.
	//if os.Geteuid() != 0 {
	//	common.ColoredText("31", "Please run as root (sudo).")
	//	os.Exit(1)
	//}

	////////////////////////////////////////
	// Get Certification & Setup Inputs
	////////////////////////////////////////

	certification := common.SelectMenu([]string{"SSL", "No SSL"})
	setup := common.SelectMenu([]string{"Default", "Websocket"})

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
		certBaseName := common.SelectCert(certBasePath)
		certFullPath := filepath.Join(certBasePath, certBaseName+".crt")
		domains, _ := common.ExtractDomains(certFullPath)
		selectedDomain = common.SelectMenu(domains)
		certPath = certFullPath
		keyFullPath := filepath.Join(certBasePath, certBaseName+".key")
		if !common.FileExists(keyFullPath) {
			common.ColoredText("93", "Key file not found for certificate.")
			os.Exit(1)
		}
		keyPath = keyFullPath

		httpPort = common.PromptInput("Please enter http port (80 is default):")
		if httpPort == "" {
			httpPort = "80"
		}
		httpsPort = common.PromptInput("Please enter https port (443 is default):")
		if httpsPort == "" {
			httpsPort = "443"
		}
	} else { // No SSL
		serverIP = common.PromptInput("Please enter the ip of this server:")
		selectedDomain = serverIP
		httpPort = common.PromptInput("Please enter http port (80 is default):")
		if httpPort == "" {
			httpPort = "80"
		}
	}

	////////////////////////////////////////
	// Nginx Configuration
	////////////////////////////////////////

	configsBasePath := "/etc/nginx/conf.d"
	// List previous configuration files.
	common.ColoredText("36", "Please enter a unique name for config file. Previous configs are shown below:")
	existingConfigs, _ := filepath.Glob(filepath.Join(configsBasePath, "*.conf"))
	for _, cfg := range existingConfigs {
		fmt.Println(cfg)
	}
	configName := common.PromptInput("")
	configFilePath := filepath.Join(configsBasePath, configName+".conf")
	if common.FileExists(configFilePath) {
		common.ColoredText("93", "The name you entered already exists.")
		os.Exit(1)
	}

	upstreamIPs := common.PromptInput("Please enter the list of upstream IP addresses (space separated):")
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
		common.ColoredText("31", "Error writing config file: "+err.Error())
		os.Exit(1)
	}
	common.ColoredText("32", "Creating configuration file for load balancer and reverse proxy:")

	// Remove default configuration files if they exist.
	defaultFiles := []string{"/etc/nginx/sites-enabled/default", "/etc/nginx/conf.d/default.conf"}
	for _, df := range defaultFiles {
		if common.FileExists(df) {
			common.ColoredText("32", "Removing default configuration at "+df)
			_ = os.Remove(df)
		}
	}

	// Test the nginx configuration.
	common.ColoredText("32", "Testing nginx configuration...")
	if err := common.RunCommand("nginx", "-t"); err != nil {
		common.ColoredText("31", "Error in nginx configuration. Please check the config files.")
		os.Exit(1)
	}

	// Reload and enable nginx.
	common.ColoredText("32", "Reloading nginx...")
	_ = common.RunCommand("systemctl", "reload", "nginx")
	common.ColoredText("32", "Enabling nginx service to automatically start after reboot...")
	_ = common.RunCommand("systemctl", "enable", "nginx")

	common.ColoredText("36", "Reverse proxy and Load balancer installation and configuration completed successfully.")

	////////////////////////////////////////
	// Firewall (ufw) Setup
	////////////////////////////////////////

	common.ColoredText("32", "Allowing SSH on port 22 and web traffic on ports 80, 443...")
	_ = common.RunCommand("ufw", "allow", "9011/tcp")
	_ = common.RunCommand("ufw", "allow", "22/tcp")
	_ = common.RunCommand("ufw", "allow", "80/tcp")
	_ = common.RunCommand("ufw", "allow", "443/tcp")

	// Enable ufw (this may prompt for confirmation).
	_ = common.RunCommand("ufw", "--force", "enable")

	common.ColoredText("36", "All is done.")
}
