package configure

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"nginx_configure/common"
	"os"
	"path/filepath"
	"strings"
)

func Configure(
	configsBasePath string,
	certBasePath string,
	configName string,
	setup string,
	upstreamIPs []string,
	cType string,
	certName string,
	domain string,
	serverIp string,
	httpPort string,
	httpsPort string,

) tea.Cmd {
	//var cmds []tea.Cmd

	certPath := certBasePath + certName + ".crt"
	keyPath := certBasePath + certName + ".key"
	configFilePath := filepath.Join(configsBasePath, configName+".conf")

	var upstreamConf strings.Builder
	upstreamConf.WriteString(fmt.Sprintf("upstream %s {", configName))
	if setup == "Websocket" {
		upstreamConf.WriteString("\n    ip_hash;")
	}
	for _, ip := range upstreamIPs {
		upstreamConf.WriteString(fmt.Sprintf("\n    server %s;", ip))
	}
	upstreamConf.WriteString("\n}")

	var configContent string
	// Build configuration based on the chosen options.
	if cType == "SSL" && setup == "Default" {
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
		configContent = fmt.Sprintf(config, upstreamConf.String(), httpPort, domain, httpsPort, domain, certPath, keyPath, configName)
	} else if cType == "SSL" && setup == "Websocket" {
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
		configContent = fmt.Sprintf(config, upstreamConf.String(), httpPort, domain, httpsPort, domain, certPath, keyPath, configName)
	} else if cType == "No SSL" && setup == "Default" {
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
		configContent = fmt.Sprintf(config, upstreamConf.String(), httpPort, serverIp, configName)
	} else if cType == "No SSL" && setup == "Websocket" {
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
		configContent = fmt.Sprintf(config, upstreamConf.String(), httpPort, serverIp, configName)
	}

	_ = configFilePath
	_ = configContent

	//Write the configuration file.
	//err := os.WriteFile(configFilePath, []byte(configContent), 0644)
	//if err != nil {
	//	common.ColoredText("31", "Error writing config file: "+err.Error())
	//	os.Exit(1)
	//}
	//common.ColoredText("32", "Creating configuration file for load balancer and reverse proxy:")

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
	cmd1 := "nginx -t"

	// Reload and enable nginx.
	common.ColoredText("32", "Reloading nginx...")
	cmd2 := "systemctl reload nginx"
	common.ColoredText("32", "Enabling nginx service to automatically start after reboot...")
	cmd3 := "systemctl enable nginx"

	common.ColoredText("36", "Reverse proxy and Load balancer installation and configuration completed successfully.")

	////////////////////////////////////////
	// Firewall (ufw) Setup
	////////////////////////////////////////

	common.ColoredText("32", "Allowing SSH on port 22 and web traffic on ports 80, 443...")
	cmd4 := "ufw allow 9011/tcp"
	cmd5 := "ufw allow 22/tcp"
	cmd6 := "ufw allow 80/tcp"
	cmd7 := "ufw allow 443/tcp"

	// Enable ufw (this may prompt for confirmation).
	cmd8 := "ufw --force enable"

	common.ColoredText("36", "All is done.")

	return tea.Sequence(
		common.RunCommand(cmd1),
		common.RunCommand(cmd2),
		common.RunCommand(cmd3),
		common.RunCommand(cmd4),
		common.RunCommand(cmd5),
		common.RunCommand(cmd6),
		common.RunCommand(cmd7),
		common.RunCommand(cmd8),
	)

}
