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

	return tea.Sequence(
		common.LogMessage("Creating config file...", common.Gold),
		func() tea.Msg {
			err := os.WriteFile(configFilePath, []byte(configContent), 0644)
			if err != nil {
				return common.CreateSingleLog("Error writing config file: "+err.Error(), common.Red)
			}
			return common.CreateSingleLog("Config file created successfully.", common.Gold)
		},
		func() tea.Msg {
			defaultFiles := []string{"/etc/nginx/sites-enabled/default", "/etc/nginx/conf.d/default.conf"}
			var logs []common.LogItem
			haveError := false
			for _, df := range defaultFiles {
				if common.FileExists(df) {
					logs = append(logs, common.LogItem{Msg: "Removing default configuration at " + df, Color: common.White})
					err := os.Remove(df)
					if err != nil {
						logs = append(logs, common.LogItem{Msg: "Error Removing default configuration at " + df, Color: common.Red})
						haveError = true
					}
				}
			}
			if !haveError {
				logs = append(logs, common.LogItem{Msg: "Default files removed.", Color: common.Gold})
			}
			return common.LogData{Messages: logs}
		},
		common.LogMessage("Testing nginx configuration...", common.Gold),
		common.RunCommand("nginx -t"),
		common.LogMessage("Reloading nginx...", common.Gold),
		common.RunCommand("systemctl reload nginx"),
		common.LogMessage("Enabling nginx service to automatically start after reboot...", common.Gold),
		common.RunCommand("systemctl enable nginx"),
		common.LogMessage("Reverse proxy and Load balancer installation and configuration completed successfully.", common.Gold),
		common.LogMessage("Allowing SSH on port 22 and web traffic on ports 80, 443...", common.Gold),
		common.RunCommand("ufw allow 9011/tcp"),
		common.RunCommand("ufw allow 22/tcp"),
		common.RunCommand("ufw allow 80/tcp"),
		common.RunCommand("ufw allow 443/tcp"),
		common.RunCommand("ufw --force enable"),
		common.LogMessage("All is done.", common.Green),
	)

}
