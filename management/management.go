package management

import (
	"bufio"
	"fmt"
	"nginx_configure/common"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// --------------------
// Nginx Management
// --------------------

const ConfigsBasePath = "/etc/nginx/conf.d"

// installNginx updates packages and installs nginx.
func installNginx() {
	common.ColoredText("32", "Updating package list...")
	_ = common.RunCommand("apt-get", "update", "-y")
	common.ColoredText("32", "Installing nginx...")
	_ = common.RunCommand("apt-get", "install", "nginx", "-y")
}

// uninstallNginx stops nginx, purges its packages, and removes its configuration.
func uninstallNginx() {
	if _, err := exec.LookPath("nginx"); err == nil {
		common.ColoredText("32", "Nginx is installed. Purging existing installation and configuration files...")
		common.ColoredText("32", "Stopping nginx service...")
		_ = common.RunCommand("systemctl", "stop", "nginx")
		common.ColoredText("32", "Purging nginx...")
		_ = common.RunCommand("apt-get", "purge", "-y", "nginx")
		common.ColoredText("32", "Auto removing packages...")
		_ = common.RunCommand("apt-get", "autoremove", "-y")
		common.ColoredText("32", "Removing nginx directory...")
		err := os.RemoveAll("/etc/nginx")
		if err != nil {
			common.ColoredText("31", "An Error occurred : "+err.Error())
			os.Exit(1)
		}
	}
}

// getConfigs returns the list of config file names in ConfigsBasePath.
func getConfigs() ([]string, error) {
	var configs []string
	err := filepath.Walk(ConfigsBasePath, func(path string, info os.FileInfo, err error) error {
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
		common.ColoredText("93", "No configs found.")
		os.Exit(1)
	}
	return configs, nil
}

// deleteConfig deletes a config file from ConfigsBasePath.
func deleteConfig(fileName string) {
	if fileName != "" {
		path := filepath.Join(ConfigsBasePath, fileName)
		_ = os.Remove(path)
		common.ColoredText("32", fmt.Sprintf("Config %s deleted.", fileName))
	} else {
		common.ColoredText("31", "Cannot delete directory conf.d in path /etc/nginx/conf.d")
	}
}

// editConfig opens a config file in nano.
func editConfig(fileName string) {
	path := filepath.Join(ConfigsBasePath, fileName)
	_ = common.RunCommand("nano", path)
	common.ColoredText("32", fmt.Sprintf("Config %s edited.", fileName))
}

// --------------------
// Firewall Management
// --------------------

// firewallStatus displays the ufw status.
func firewallStatus() {
	_ = common.RunCommand("ufw", "status")
}

// installFirewall installs ufw and opens some default ports.
func installFirewall() {
	common.ColoredText("32", "Updating package list...")
	_ = common.RunCommand("apt-get", "update", "-y")
	common.ColoredText("32", "Installing firewall...")
	_ = common.RunCommand("apt-get", "install", "-y", "ufw")
	if _, err := exec.LookPath("ufw"); err == nil {
		_ = common.RunCommand("ufw", "allow", "9011/tcp")
		_ = common.RunCommand("ufw", "allow", "22/tcp")
	}
}

// uninstallFirewall disables ufw, purges it, and removes its directory.
func uninstallFirewall() {
	if _, err := exec.LookPath("ufw"); err == nil {
		common.ColoredText("32", "Firewall is installed. Purging existing installation and configuration files...")
		common.ColoredText("32", "Stopping ufw service...")
		_ = common.RunCommand("ufw", "disable")
		common.ColoredText("32", "Purging firewall ufw...")
		_ = common.RunCommand("apt-get", "purge", "-y", "ufw")
		common.ColoredText("32", "Auto removing packages...")
		_ = common.RunCommand("apt-get", "autoremove", "-y", "ufw")
		common.ColoredText("32", "Removing ufw directory...")
		err := os.RemoveAll("/etc/ufw")
		if err != nil {
			common.ColoredText("31", "An Error occurred : "+err.Error())
			os.Exit(1)
		}
	}
}

// openingPorts prompts for ports and opens them in ufw.
func openingPorts() {
	reader := bufio.NewReader(os.Stdin)
	common.ColoredText("32", "Enter ports to open (separated by space):")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	ports := strings.Split(input, " ")
	for _, port := range ports {
		common.ColoredText("32", fmt.Sprintf("Opening port: %s", port))
		_ = common.RunCommand("ufw", "allow", fmt.Sprintf("%s/tcp", port))
	}
	common.ColoredText("32", "Check firewall status")
	_ = common.RunCommand("ufw", "status")
}

// --------------------
// Certificate Management
// --------------------

const CertBasePath = "/etc/ssl/files"

// ensureCertBasePath creates CertBasePath if it does not exist.
func ensureCertBasePath() {
	err := os.MkdirAll(CertBasePath, 0755)
	if err != nil {
		common.ColoredText("31", "An Error occurred : "+err.Error())
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
		common.ColoredText("93", "No certificates found.")
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
	common.ColoredText("36", info)
}

// getCert prompts the user to enter certificate content using nano and writes it to file.
func getCert(name string) (string, error) {
	tmpFile, err := os.CreateTemp("", "cert_*.crt")
	if err != nil {
		common.ColoredText("31", "An Error occurred : "+err.Error())
		os.Exit(1)
	}
	tmpFilePath := tmpFile.Name()
	err2 := tmpFile.Close()
	if err2 != nil {
		common.ColoredText("31", "An Error occurred : "+err2.Error())
		os.Exit(1)
	}
	common.ColoredText("36", "Please enter your certificate content in nano. Save and exit when done.")
	if err := common.RunCommand("nano", tmpFilePath); err != nil {
		return "", err
	}
	content, err := os.ReadFile(tmpFilePath)
	if err != nil {
		return "", err
	}
	err3 := os.Remove(tmpFilePath)
	if err3 != nil {
		common.ColoredText("31", "An Error occurred : "+err3.Error())
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
		common.ColoredText("31", "An Error occurred : "+err2.Error())
		os.Exit(1)
	}
	common.ColoredText("36", "Please enter your private key content in nano. Save and exit when done.")
	if err := common.RunCommand("nano", tmpFilePath); err != nil {
		return "", err
	}
	content, err := os.ReadFile(tmpFilePath)
	if err != nil {
		return "", err
	}
	err3 := os.Remove(tmpFilePath)
	if err3 != nil {
		common.ColoredText("31", "An Error occurred : "+err3.Error())
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
	common.ColoredText("32", fmt.Sprintf("Removing ssl certificate '%s'", baseName))
	err := os.Remove(filepath.Join(CertBasePath, baseName+".crt"))
	if err != nil {
		common.ColoredText("31", "An Error occurred : "+err.Error())
		os.Exit(1)
	}
	err2 := os.Remove(filepath.Join(CertBasePath, baseName+".key"))
	if err2 != nil {
		common.ColoredText("31", "An Error occurred : "+err2.Error())
		os.Exit(1)
	}
}

// deleteAllCertificates removes all certificate and key files from CertBasePath.
func deleteAllCertificates() {
	common.ColoredText("32", "Removing all ssl certificates...")
	if files, err := filepath.Glob(filepath.Join(CertBasePath, "*.crt")); err == nil {
		for _, f := range files {
			err := os.Remove(f)
			if err != nil {
				common.ColoredText("31", "An Error occurred : "+err.Error())
				os.Exit(1)
			}
		}
	}
	if files, err := filepath.Glob(filepath.Join(CertBasePath, "*.key")); err == nil {
		for _, f := range files {
			err := os.Remove(f)
			if err != nil {
				common.ColoredText("31", "An Error occurred : "+err.Error())
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
	common.ColoredText("32", "Updating package list...")
	_ = common.RunCommand("apt-get", "update", "-y")
	common.ColoredText("32", "Installing nginx...")
	_ = common.RunCommand("apt-get", "install", "nginx", "-y")
	common.ColoredText("32", "Installing firewall...")
	_ = common.RunCommand("apt-get", "install", "-y", "ufw")
}

// uninstallEverything calls the uninstallation routines.
func uninstallEverything() {
	uninstallNginx()
	uninstallFirewall()
	deleteAllCertificates()
}

// confirmPrompt asks for a yes/no confirmation.
func confirmPrompt(message string) bool {
	common.ColoredText("94", message+" (yes/y to confirm, no/n to cancel):")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))
	return input == "yes" || input == "y"
}

// --------------------
// Main Menu
// --------------------

func Management() {
	// Check if running as root.
	//if os.Geteuid() != 0 {
	//	common.ColoredText("31", "Please run as root (sudo).")
	//	os.Exit(1)
	//}

	// Run initial routines.
	if err := common.FixDpkgLock(); err != nil {
		common.ColoredText("31", "Error fixing dpkg lock: "+err.Error())
	}
	common.ClearCache()
	ensureCertBasePath()

	mainOptions := []string{
		"Install Requirements",
		"Nginx Management",
		"Firewall Management",
		"Certificate Management",
		"Reinstall everything",
		"Uninstall and delete everything",
	}

	common.ColoredText("32", "Management menu")
	mainOpt := common.SelectMenu(mainOptions)

	switch mainOpt {
	case "Install Requirements":
		installRequirements()

	case "Nginx Management":
		nginxOptions := []string{"Install Nginx", "Delete Nginx", "Add Config", "Manage Configs"}
		nginxOpt := common.SelectMenu(nginxOptions)
		switch nginxOpt {
		case "Install Nginx":
			installNginx()
		case "Delete Nginx":
			if confirmPrompt("Do you really want to uninstall nginx?") {
				uninstallNginx()
			} else {
				common.ColoredText("93", "Uninstall nginx canceled.")
			}
		case "Add Config":
			// Check for nginx and ufw before executing the remote script.
			if _, err := exec.LookPath("nginx"); err != nil {
				common.ColoredText("31", "Nginx not installed.")
			} else if _, err := exec.LookPath("ufw"); err != nil {
				common.ColoredText("31", "Firewall not installed.")
			} else {
				//configure.Configure()
			}
		case "Manage Configs":
			configs, err := getConfigs()
			if err != nil {
				common.ColoredText("31", "Error fetching configs: "+err.Error())
				os.Exit(1)
			}
			selectedConfig := common.SelectMenu(configs)
			manageOptions := []string{"Edit Config", "Delete Config"}
			manageOpt := common.SelectMenu(manageOptions)
			switch manageOpt {
			case "Edit Config":
				editConfig(selectedConfig)
			case "Delete Config":
				if confirmPrompt(fmt.Sprintf("Do you really want to delete config %s?", selectedConfig)) {
					deleteConfig(selectedConfig)
				} else {
					common.ColoredText("93", fmt.Sprintf("Delete config %s canceled.", selectedConfig))
				}
			}
		}

	case "Firewall Management":
		fwOptions := []string{"Firewall Status", "Install Firewall", "Delete Firewall", "Open port(s)"}
		fwOpt := common.SelectMenu(fwOptions)
		switch fwOpt {
		case "Firewall Status":
			firewallStatus()
		case "Install Firewall":
			installFirewall()
		case "Delete Firewall":
			if confirmPrompt("Do you really want to uninstall firewall (ufw)?") {
				uninstallFirewall()
			} else {
				common.ColoredText("93", "Uninstall firewall (ufw) canceled.")
			}
		case "Open port(s)":
			openingPorts()
		}

	case "Certificate Management":
		certOptions := []string{"Add Certificate", "Delete All Certificates", "Manage a Certificate"}
		certOpt := common.SelectMenu(certOptions)
		switch certOpt {
		case "Add Certificate":
			common.ColoredText("94", "Enter certificate file name. The name must be unique:")
			reader := bufio.NewReader(os.Stdin)
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)
			certPath, err := getCert(name)
			if err != nil {
				common.ColoredText("31", "Error adding certificate: "+err.Error())
				break
			}
			keyPath, err := getKey(name)
			if err != nil {
				common.ColoredText("31", "Error adding key: "+err.Error())
				break
			}
			if certPath != "" && keyPath != "" {
				common.ColoredText("32", fmt.Sprintf("Certificate %s created.", name))
			}
		case "Delete All Certificates":
			if confirmPrompt("Do you really want to delete all certificates?") {
				deleteAllCertificates()
			} else {
				common.ColoredText("93", "Delete all certificates canceled.")
			}
		case "Manage a Certificate":
			certs, err := getCertificates()
			if err != nil {
				common.ColoredText("31", "Error fetching certificates: "+err.Error())
				os.Exit(1)
			}
			// Build a slice of certificate descriptions.
			var certDescs []string
			for _, desc := range certs {
				certDescs = append(certDescs, desc)
			}
			selectedDesc := common.SelectMenu(certDescs)
			// Find the certificate file path corresponding to the selected description.
			var selectedPath string
			for path, desc := range certs {
				if desc == selectedDesc {
					selectedPath = path
					break
				}
			}
			manageCertOptions := []string{"Certificate Info", "Delete Certificate"}
			certManageOpt := common.SelectMenu(manageCertOptions)
			switch certManageOpt {
			case "Certificate Info":
				certificateInfo(selectedPath)
			case "Delete Certificate":
				if confirmPrompt(fmt.Sprintf("Do you really want to delete certificate %s?", selectedPath)) {
					deleteCertificate(selectedPath)
				} else {
					common.ColoredText("93", fmt.Sprintf("Delete certificate %s canceled.", selectedPath))
				}
			}
		}

	case "Reinstall everything":
		if confirmPrompt("Do you really want to reinstall everything?") {
			uninstallNginx()
			installNginx()
			uninstallFirewall()
			installFirewall()
			deleteAllCertificates()
			installRequirements()
		} else {
			common.ColoredText("93", "Reinstall everything canceled.")
		}

	case "Uninstall and delete everything":
		if confirmPrompt("Do you really want to uninstall and delete everything?") {
			uninstallEverything()
		} else {
			common.ColoredText("93", "Uninstall and delete everything canceled.")
		}
	}

	// If nginx is installed, test configuration and reload.
	if _, err := exec.LookPath("nginx"); err == nil {
		common.ColoredText("32", "Testing nginx configuration...")
		if err := common.RunCommand("nginx", "-t"); err != nil {
			common.ColoredText("31", "Error in nginx configuration. Please check the config files.")
			os.Exit(1)
		}
		common.ColoredText("32", "Reloading nginx...")
		_ = common.RunCommand("systemctl", "reload", "nginx")
	}

	common.ClearCache()
}
