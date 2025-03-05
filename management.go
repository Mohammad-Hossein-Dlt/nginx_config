package main

//
//import (
//	"bufio"
//	"fmt"
//	"nginx_configure/common"
//	"os"
//	"os/exec"
//	"path/filepath"
//	"strings"
//)
//
//// --------------------
//// Main Menu
//// --------------------
//
//func Management() {
//
//	// Run initial routines.
//	if err := common.FixDpkgLock(); err != nil {
//		common.ColoredText("31", "Error fixing dpkg lock: "+err.Error())
//	}
//	common.ClearCache()
//	ensureCertBasePath()
//
//	mainOptions := []string{
//		"Install Requirements",
//		"Nginx Management",
//		"Firewall Management",
//		"Certificate Management",
//		"Reinstall everything",
//		"Uninstall and delete everything",
//	}
//
//	common.ColoredText("32", "Management menu")
//	mainOpt := common.SelectMenu(mainOptions)
//
//	switch mainOpt {
//	case "Install Requirements":
//		installRequirements()
//
//	case "Nginx Management":
//		nginxOptions := []string{"Install Nginx", "Delete Nginx", "Add Config", "Manage Configs"}
//		nginxOpt := common.SelectMenu(nginxOptions)
//		switch nginxOpt {
//		case "Install Nginx":
//			installNginx()
//		case "Delete Nginx":
//			if confirmPrompt("Do you really want to uninstall nginx?") {
//				uninstallNginx()
//			} else {
//				common.ColoredText("93", "Uninstall nginx canceled.")
//			}
//		case "Add Config":
//			// Check for nginx and ufw before executing the remote script.
//			if _, err := exec.LookPath("nginx"); err != nil {
//				common.ColoredText("31", "Nginx not installed.")
//			} else if _, err := exec.LookPath("ufw"); err != nil {
//				common.ColoredText("31", "Firewall not installed.")
//			} else {
//				//configure.Configure()
//			}
//		case "Manage Configs":
//			configs, err := getConfigs()
//			if err != nil {
//				common.ColoredText("31", "Error fetching configs: "+err.Error())
//				os.Exit(1)
//			}
//			selectedConfig := common.SelectMenu(configs)
//			manageOptions := []string{"Edit Config", "Delete Config"}
//			manageOpt := common.SelectMenu(manageOptions)
//			switch manageOpt {
//			case "Edit Config":
//				editConfig(selectedConfig)
//			case "Delete Config":
//				if confirmPrompt(fmt.Sprintf("Do you really want to delete config %s?", selectedConfig)) {
//					deleteConfig(selectedConfig)
//				} else {
//					common.ColoredText("93", fmt.Sprintf("Delete config %s canceled.", selectedConfig))
//				}
//			}
//		}
//
//	case "Firewall Management":
//		fwOptions := []string{"Firewall Status", "Install Firewall", "Delete Firewall", "Open port(s)"}
//		fwOpt := common.SelectMenu(fwOptions)
//		switch fwOpt {
//		case "Firewall Status":
//			firewallStatus()
//		case "Install Firewall":
//			installFirewall()
//		case "Delete Firewall":
//			if confirmPrompt("Do you really want to uninstall firewall (ufw)?") {
//				uninstallFirewall()
//			} else {
//				common.ColoredText("93", "Uninstall firewall (ufw) canceled.")
//			}
//		case "Open port(s)":
//			openingPorts()
//		}
//
//	case "Certificate Management":
//		certOptions := []string{"Add Certificate", "Delete All Certificates", "Manage a Certificate"}
//		certOpt := common.SelectMenu(certOptions)
//		switch certOpt {
//		case "Add Certificate":
//			common.ColoredText("94", "Enter certificate file name. The name must be unique:")
//			reader := bufio.NewReader(os.Stdin)
//			name, _ := reader.ReadString('\n')
//			name = strings.TrimSpace(name)
//			certPath, err := getCert(name)
//			if err != nil {
//				common.ColoredText("31", "Error adding certificate: "+err.Error())
//				break
//			}
//			keyPath, err := getKey(name)
//			if err != nil {
//				common.ColoredText("31", "Error adding key: "+err.Error())
//				break
//			}
//			if certPath != "" && keyPath != "" {
//				common.ColoredText("32", fmt.Sprintf("Certificate %s created.", name))
//			}
//		case "Delete All Certificates":
//			if confirmPrompt("Do you really want to delete all certificates?") {
//				deleteAllCertificates()
//			} else {
//				common.ColoredText("93", "Delete all certificates canceled.")
//			}
//		case "Manage a Certificate":
//			certs, err := getCertificates()
//			if err != nil {
//				common.ColoredText("31", "Error fetching certificates: "+err.Error())
//				os.Exit(1)
//			}
//			// Build a slice of certificate descriptions.
//			var certDescs []string
//			for _, desc := range certs {
//				certDescs = append(certDescs, desc)
//			}
//			selectedDesc := common.SelectMenu(certDescs)
//			// Find the certificate file path corresponding to the selected description.
//			var selectedPath string
//			for path, desc := range certs {
//				if desc == selectedDesc {
//					selectedPath = path
//					break
//				}
//			}
//			manageCertOptions := []string{"Certificate Info", "Delete Certificate"}
//			certManageOpt := common.SelectMenu(manageCertOptions)
//			switch certManageOpt {
//			case "Certificate Info":
//				certificateInfo(selectedPath)
//			case "Delete Certificate":
//				if confirmPrompt(fmt.Sprintf("Do you really want to delete certificate %s?", selectedPath)) {
//					deleteCertificate(selectedPath)
//				} else {
//					common.ColoredText("93", fmt.Sprintf("Delete certificate %s canceled.", selectedPath))
//				}
//			}
//		}
//
//	case "Reinstall everything":
//		if confirmPrompt("Do you really want to reinstall everything?") {
//			uninstallNginx()
//			installNginx()
//			uninstallFirewall()
//			installFirewall()
//			deleteAllCertificates()
//			installRequirements()
//		} else {
//			common.ColoredText("93", "Reinstall everything canceled.")
//		}
//
//	case "Uninstall and delete everything":
//		if confirmPrompt("Do you really want to uninstall and delete everything?") {
//			uninstallEverything()
//		} else {
//			common.ColoredText("93", "Uninstall and delete everything canceled.")
//		}
//	}
//
//	// If nginx is installed, test configuration and reload.
//	if _, err := exec.LookPath("nginx"); err == nil {
//		common.ColoredText("32", "Testing nginx configuration...")
//		if err := common.RunCommand("nginx", "-t"); err != nil {
//			common.ColoredText("31", "Error in nginx configuration. Please check the config files.")
//			os.Exit(1)
//		}
//		common.ColoredText("32", "Reloading nginx...")
//		_ = common.RunCommand("systemctl", "reload", "nginx")
//	}
//
//	common.ClearCache()
//}
