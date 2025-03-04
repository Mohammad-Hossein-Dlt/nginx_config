package tui

import (
	"fmt"
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"log"
	"nginx_configure/common"
	"nginx_configure/configure"
	"path/filepath"
	"strconv"
	"strings"
)

type State int

const (
	MainList State = iota
	InstallRequirements
	NginxManagement
	FirewallManagement
	CertificateManagement
	ReinstallEverything
	UninstallAndDeleteEverything
)

const (
	InstallNginx State = iota + 7
	DeleteNginx
	ConfigName
	ManageConfigs
)

const (
	Setup State = iota + 11
	Upstreams
	CType
	SelectCert
	Domains
	ServerIp
	HttpPort
	HttpsPort
)

type ListModel struct {
	Options   []string
	ListIndex int
}

type CertListModel struct {
	Options   map[string]string
	ListIndex int
}

type NewConfig struct {
	DuplicateName bool
	Name          string
	Setup         string
	Upstreams     []string
	CType         string
	CertName      string
	Domain        string
	ServerIp      string
	HttpPort      string
	HttpsPort     string
}

type CLIModel struct {
	State State

	MainMenu        ListModel
	NginxMenu       ListModel
	FirewallMenu    ListModel
	CertificateMenu ListModel
	//-------------------------
	NewConfig NewConfig

	CTypes  ListModel
	Certs   CertListModel
	Domains ListModel
	Setups  ListModel
	//-------------------------

	TextInput  textinput.Model
	FilePicker filepicker.Model
	//-------------------------
	Logs []common.LogMsg
}

const (
	configsBasePath = "/etc/nginx/conf.d/"
	CertBasePath    = "/etc/ssl/files/"
)

var (
	//information       = lipgloss.NewStyle().Margin(0, 0, 0, 0).Padding(0, 0, 0, 0).Foreground(lipgloss.Color("#1E90FF"))
	itemStyle1        = lipgloss.NewStyle()
	information       = lipgloss.NewStyle().Margin(0, 0, 0, 0).Padding(0, 0, 0, 0).Foreground(lipgloss.Color("170"))
	files             = lipgloss.NewStyle().Margin(0, 0, 0, 0).Padding(0, 0, 0, 0).Foreground(lipgloss.Color("170"))
	itemStyle         = lipgloss.NewStyle().MarginLeft(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	logStyle          = lipgloss.NewStyle().Margin(0, 0, 0, 2)
	simpleStyle       = lipgloss.NewStyle().Margin(0, 0, 0, 2)
)

func initial() *CLIModel {

	ti := textinput.New()

	fp := filepicker.New()
	fp.AllowedTypes = []string{}

	m := CLIModel{
		State: MainList,
		MainMenu: ListModel{
			Options: []string{
				"Install Requirements",
				"Nginx Management",
				"Firewall Management",
				"Certificate Management",
				"Reinstall everything",
				"Uninstall and delete everything",
			},
			ListIndex: 0,
		},
		NginxMenu: ListModel{
			Options: []string{
				"Install Nginx",
				"Delete Nginx",
				"Add Configs",
				"Manage Configs",
			},
			ListIndex: 0,
		},
		FirewallMenu: ListModel{
			Options: []string{
				"Firewall Status",
				"Install Firewall",
				"Delete Firewall",
				"Open port(s)",
			},
			ListIndex: 0,
		},
		CertificateMenu: ListModel{
			Options: []string{
				"Add Certificate",
				"Delete All Certificates",
				"Manage a Certificate",
			},
			ListIndex: 0,
		},
		CTypes: ListModel{
			Options: []string{
				"SSL",
				"No SSL",
			},
			ListIndex: 0,
		},
		Setups: ListModel{
			Options: []string{
				"Websocket",
				"Default",
			},
			ListIndex: 0,
		},
		TextInput:  ti,
		FilePicker: fp,
	}

	return &m
}

func (m *CLIModel) Init() tea.Cmd {
	return nil
}

func (m *CLIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmdS []tea.Cmd
		cmd  tea.Cmd
	)

	m.TextInput, cmd = m.TextInput.Update(msg)
	cmdS = append(cmdS, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch m.State {
		case MainList:
			menu := m.MainMenu
			switch key {
			case "q":
				return m, tea.Quit
			case "up", "w":
				if menu.ListIndex > 0 {
					m.MainMenu.ListIndex--
				}
			case "down", "s":
				if menu.ListIndex < len(menu.Options)-1 {
					m.MainMenu.ListIndex++
				}
			case "enter":
				switch menu.Options[menu.ListIndex] {
				case "Install Requirements":
					m.State = InstallRequirements
				case "Nginx Management":
					m.State = NginxManagement
				case "Firewall Management":
					m.State = FirewallManagement
				case "Certificate Management":
					m.State = CertificateManagement
				case "Reinstall everything":
					m.State = ReinstallEverything
				case "Uninstall and delete everything":
					m.State = UninstallAndDeleteEverything
				}
			}
		case InstallRequirements:
			menu := m.MainMenu
			switch key {
			case "q":
				return m, tea.Quit
			case "b":
				m.State = MainList
			case "up", "w":
				if menu.ListIndex > 0 {
					m.MainMenu.ListIndex--
				}
			case "down", "s":
				if menu.ListIndex < len(menu.Options)-1 {
					m.MainMenu.ListIndex++
				}
			case "enter":

			}

		case NginxManagement:
			menu := m.NginxMenu
			switch key {
			case "q":
				return m, tea.Quit
			case "b":
				m.State = MainList
			case "up", "w":
				if menu.ListIndex > 0 {
					m.NginxMenu.ListIndex--
				}
			case "down", "s":
				if menu.ListIndex < len(menu.Options)-1 {
					m.NginxMenu.ListIndex++
				}
			case "enter":
				switch menu.Options[menu.ListIndex] {
				case "Install Nginx":
					m.State = InstallNginx
				case "Delete Nginx":
					m.TextInput.SetValue("")
					m.TextInput.Focus()
					m.State = DeleteNginx
				case "Add Configs":
					m.TextInput.SetValue("")
					m.TextInput.Focus()
					m.State = ConfigName
				case "Manage Configs":
					m.State = ManageConfigs
				}
			}

		case FirewallManagement:
			menu := m.FirewallMenu
			switch key {
			case "q":
				return m, tea.Quit
			case "b":
				m.State = MainList
			case "up", "w":
				if menu.ListIndex > 0 {
					m.FirewallMenu.ListIndex--
				}
			case "down", "s":
				if menu.ListIndex < len(menu.Options)-1 {
					m.FirewallMenu.ListIndex++
				}
			case "enter":

			}

		case CertificateManagement:
			menu := m.NginxMenu
			switch key {
			case "q":
				return m, tea.Quit
			case "b":
				m.State = MainList
			case "up", "w":
				if menu.ListIndex > 0 {
					m.NginxMenu.ListIndex--
				}
			case "down", "s":
				if menu.ListIndex < len(menu.Options)-1 {
					m.NginxMenu.ListIndex++
				}
			case "enter":

			}

		case ReinstallEverything:
			switch key {
			case "q":
				return m, tea.Quit
			case "b":
				m.State = MainList
			case "enter":

			}
		case UninstallAndDeleteEverything:
			switch key {
			case "q":
				return m, tea.Quit
			case "b":
				m.State = MainList
			case "enter":

			}

		//-----------------------------------------------------
		case InstallNginx:
			switch key {
			case "q":
				return m, tea.Quit
			case "b":
				m.SetState(NginxManagement, nil)
			case "enter":
			}
		case DeleteNginx:
			switch key {
			case "ctrl+b":
				m.SetState(NginxManagement, nil)
			case "enter":
				value := m.TextInput.Value()
				if value == "yes" {
					fmt.Println("Deleting Nginx.")
				} else {
					fmt.Println("Keep Nginx Installed.")
				}
				return m, tea.Quit
			}
		case ConfigName:
			switch key {
			case "ctrl+c":
				return m, tea.Quit
			case "ctrl+b":
				m.SetState(NginxManagement, nil)
			case "enter":
				value := m.TextInput.Value()
				if value != "" {
					m.NewConfig.Name = value
					configFilePath := filepath.Join(configsBasePath, value+".conf")
					if common.FileExists(configFilePath) {
						m.NewConfig.DuplicateName = true
					} else {
						m.NewConfig.DuplicateName = false
						m.SetState(Setup, nil)
					}

				}
			}
		case Setup:
			menu := m.Setups
			switch key {
			case "q":
				return m, tea.Quit
			case "b":
				m.SetState(NginxManagement, nil)
			case "up", "w":
				if menu.ListIndex > 0 {
					m.Setups.ListIndex--
				}
			case "down", "s":
				if menu.ListIndex < len(menu.Options)-1 {
					m.Setups.ListIndex++
				}
			case "enter":
				m.NewConfig.Setup = menu.Options[menu.ListIndex]
				m.TextInput.SetValue("")
				m.TextInput.Focus()
				m.SetState(Upstreams, nil)
			}
		case Upstreams:
			switch key {
			case "ctrl+c":
				return m, tea.Quit
			case "ctrl+b":
				m.SetState(NginxManagement, nil)
			case "enter":
				value := m.TextInput.Value()
				if value != "" {
					m.NewConfig.Upstreams = strings.Fields(value)
					m.SetState(CType, nil)
				}
			}
		case CType:
			menu := m.CTypes
			switch key {
			case "q":
				return m, tea.Quit
			case "b":
				m.SetState(NginxManagement, nil)
			case "up", "w":
				if menu.ListIndex > 0 {
					m.CTypes.ListIndex--
				}
			case "down", "s":
				if menu.ListIndex < len(menu.Options)-1 {
					m.CTypes.ListIndex++
				}
			case "enter":
				m.NewConfig.CType = menu.Options[menu.ListIndex]
				if m.NewConfig.CType == "SSL" {
					certs, logMsg := common.Certificates(CertBasePath)
					m.Certs = CertListModel{Options: certs}
					m.SetState(SelectCert, &logMsg)
				} else {
					m.TextInput.SetValue("")
					m.TextInput.Focus()
					m.SetState(ServerIp, nil)
				}
			}
		case SelectCert:
			menu := m.Certs
			switch key {
			case "q":
				return m, tea.Quit
			case "b":
				m.SetState(NginxManagement, nil)
			case "up", "w":
				if menu.ListIndex > 0 {
					m.CTypes.ListIndex--
				}
			case "down", "s":
				if menu.ListIndex < len(menu.Options)-1 {
					m.CTypes.ListIndex++
				}
			case "enter":
				m.NewConfig.CertName = common.GetKeyByIndex(m.Certs.Options, m.Certs.ListIndex)
				domains, _ := common.ExtractDomains(CertBasePath + m.NewConfig.CertName + ".crt")
				m.Domains = ListModel{Options: domains}
				m.SetState(Domains, nil)
			}
		case Domains:
			menu := m.Domains
			switch key {
			case "q":
				return m, tea.Quit
			case "b":
				m.SetState(NginxManagement, nil)
			case "up", "w":
				if menu.ListIndex > 0 {
					m.Domains.ListIndex--
				}
			case "down", "s":
				if menu.ListIndex < len(menu.Options)-1 {
					m.Domains.ListIndex++
				}
			case "enter":
				m.NewConfig.Domain = m.Domains.Options[m.Domains.ListIndex]
				m.TextInput.SetValue("")
				m.TextInput.Focus()
				m.SetState(HttpPort, nil)
			}
		case ServerIp:
			switch key {
			case "ctrl+c":
				return m, tea.Quit
			case "ctrl+b":
				m.SetState(NginxManagement, nil)
			case "enter":
				value := m.TextInput.Value()
				if value != "" {
					m.NewConfig.ServerIp = value
					m.TextInput.SetValue("")
					m.TextInput.Focus()
					m.SetState(HttpPort, nil)
				}
			}
		case HttpPort:
			switch key {
			case "ctrl+c":
				return m, tea.Quit
			case "ctrl+b":
				m.SetState(NginxManagement, nil)
			case "enter":
				value := m.TextInput.Value()
				if value != "" {
					m.NewConfig.HttpPort = value
					m.TextInput.SetValue("")
					m.TextInput.Focus()
					m.SetState(HttpsPort, nil)
				}
			}
		case HttpsPort:
			switch key {
			case "ctrl+c":
				return m, tea.Quit
			case "ctrl+b":
				m.SetState(NginxManagement, nil)
			case "enter":
				value := m.TextInput.Value()
				if value != "" {
					m.NewConfig.HttpsPort = value
					m.Logs = nil
					return m, configure.Configure(
						configsBasePath,
						CertBasePath,
						m.NewConfig.Name,
						m.NewConfig.Setup,
						m.NewConfig.Upstreams,
						m.NewConfig.CType,
						m.NewConfig.CertName,
						m.NewConfig.Domain,
						m.NewConfig.ServerIp,
						m.NewConfig.HttpPort,
						m.NewConfig.HttpsPort,
					)
				}
			}
		case ManageConfigs:
			switch key {
			case "q":
				return m, tea.Quit
			case "b":
				m.SetState(NginxManagement, nil)
			case "enter":

			}
		}
	case common.LogMsg:
		m.Logs = append(m.Logs, msg)
		return m, nil // Append new log message
	case common.Done:
		return m, nil
	}

	return m, tea.Batch(cmdS...)
}

func (m *CLIModel) SetState(s State, log *common.LogMsg) {
	m.State = s
	if log != nil {
		m.Logs = append(m.Logs, *log)
	} else {
		m.Logs = nil
	}
}

func (m *CLIModel) View() string {

	var sb strings.Builder

	for _, logMsg := range m.Logs {
		if logMsg.Color != "" {
			sb.WriteString(logStyle.Foreground(lipgloss.Color(logMsg.Color)).Render(logMsg.Msg) + "\n")
		} else {
			sb.WriteString(logStyle.Foreground(lipgloss.Color(common.White)).Render(logMsg.Msg) + "\n")
			//sb.WriteString(logMsg.Msg + "\n")
		}
	}

	switch m.State {
	case MainList:
		sb.WriteString(buildListItems(m.MainMenu))
	case InstallRequirements:
		sb.WriteString(itemStyle.Render("Install Requirements"))
	case NginxManagement:
		sb.WriteString(buildListItems(m.NginxMenu))

	case FirewallManagement:
		sb.WriteString(buildListItems(m.FirewallMenu))

	case CertificateManagement:
		sb.WriteString(buildListItems(m.CertificateMenu))

	case ReinstallEverything:
		sb.WriteString(itemStyle.Render("Reinstall Everything"))
	case UninstallAndDeleteEverything:
		sb.WriteString(itemStyle.Render("Uninstall And Delete Everything"))
	//-----------------------------------------------------------------------------------
	case InstallNginx:
		sb.WriteString(itemStyle.Render("Install Nginx"))
	case DeleteNginx:
		sb.WriteString("Do you really want to uninstall nginx? (yes/y to confirm, no/n to cancel):\n" + m.TextInput.View() + "\n")
	case ConfigName:
		text := "Please enter a unique name for config file. Previous configs are shown below:\n"
		existingConfigs, _ := filepath.Glob(filepath.Join(configsBasePath, "*.conf"))
		for _, cfg := range existingConfigs {
			text += cfg + "\n"
		}
		text += m.TextInput.View() + "\n"
		if m.NewConfig.DuplicateName {
			text += "The name you entered already exists."
		}
		sb.WriteString(simpleStyle.Render(text))
	case Setup:
		sb.WriteString(buildListItems(m.Setups))
	case Upstreams:
		sb.WriteString(simpleStyle.Render("Please enter the list of upstream IP addresses (space separated):\n"+m.TextInput.View()) + "\n")
	case CType:
		sb.WriteString(buildListItems(m.CTypes))
	case SelectCert:
		sb.WriteString(buildCertListItems(m.Certs))
	case Domains:
		sb.WriteString(buildListItems(m.Domains))
	case ServerIp:
		sb.WriteString(simpleStyle.Render("Please enter the ip of this server:\n"+m.TextInput.View()) + "\n")
	case HttpPort:
		sb.WriteString(simpleStyle.Render("Please enter http port (80 is default):\n"+m.TextInput.View()) + "\n")
	case HttpsPort:
		sb.WriteString(simpleStyle.Render("Please enter https port (443 is default):\n"+m.TextInput.View()) + "\n")
	case ManageConfigs:
		sb.WriteString(itemStyle.Render("Manage Configs"))

	}

	return strings.Trim(sb.String(), "!¡")
}

func buildListItems(menu ListModel) string {
	var sb strings.Builder
	for i, opt := range menu.Options {

		fn := itemStyle.Render
		if i == menu.ListIndex {
			fn = func(s ...string) string {
				return selectedItemStyle.Render("> " + strings.Join(s, ""))
			}
		} else {
			fn = func(s ...string) string {
				return itemStyle.Render("  " + strings.Join(s, ""))
			}
		}

		sb.WriteString(fn(strconv.Itoa(i+1)+". "+opt) + "\n")
	}

	return sb.String()
}

func buildCertListItems(menu CertListModel) string {

	keys := common.ExtractKeys(menu.Options)

	var sb strings.Builder
	for i, k := range keys {
		fn := itemStyle.Render
		if i == menu.ListIndex {
			fn = func(s ...string) string {
				return selectedItemStyle.Render("> " + strings.Join(s, ""))
			}
		} else {
			fn = func(s ...string) string {
				return itemStyle.Render("  " + strings.Join(s, ""))
			}
		}

		sb.WriteString(fn(strconv.Itoa(i+1)+". "+menu.Options[k]) + "\n")
	}

	return sb.String()
}

func RunApp() {
	p := tea.NewProgram(initial())
	if _, err := p.Run(); err != nil {
		log.Fatal("Error: " + err.Error())
	}
}
