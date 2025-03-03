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

const (
	configsBasePath = "/etc/nginx/conf.d/"
	CertBasePath    = "/etc/ssl/files/"
	//configsBasePath = ""
)

var (
	//information       = lipgloss.NewStyle().Margin(0, 0, 0, 0).Padding(0, 0, 0, 0).Foreground(lipgloss.Color("#1E90FF"))
	itemStyle1        = lipgloss.NewStyle()
	information       = lipgloss.NewStyle().Margin(0, 0, 0, 0).Padding(0, 0, 0, 0).Foreground(lipgloss.Color("170"))
	files             = lipgloss.NewStyle().Margin(0, 0, 0, 0).Padding(0, 0, 0, 0).Foreground(lipgloss.Color("170"))
	itemStyle         = lipgloss.NewStyle().MarginLeft(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
)

type state int

const (
	MainList state = iota
	InstallRequirements
	NginxManagement
	FirewallManagement
	CertificateManagement
	ReinstallEverything
	UninstallAndDeleteEverything
)

const (
	InstallNginx state = iota + 7
	DeleteNginx
	ConfigName
	ManageConfigs
)

const (
	Setup state = iota + 11
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

type model struct {
	state state

	mainMenu        ListModel
	nginxMenu       ListModel
	firewallMenu    ListModel
	certificateMenu ListModel
	//-------------------------
	newConfig NewConfig

	cTypes  ListModel
	certs   CertListModel
	domains ListModel
	setups  ListModel
	//-------------------------

	textInput  textinput.Model
	filePicker filepicker.Model
	//-------------------------
	logs []common.LogMsg
}

func initialModel() *model {

	ti := textinput.New()

	fp := filepicker.New()
	fp.AllowedTypes = []string{}

	return &model{
		state: MainList,
		mainMenu: ListModel{
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
		nginxMenu: ListModel{
			Options: []string{
				"Install Nginx",
				"Delete Nginx",
				"Add Configs",
				"Manage Configs",
			},
			ListIndex: 0,
		},
		firewallMenu: ListModel{
			Options: []string{
				"Firewall Status",
				"Install Firewall",
				"Delete Firewall",
				"Open port(s)",
			},
			ListIndex: 0,
		},
		certificateMenu: ListModel{
			Options: []string{
				"Add Certificate",
				"Delete All Certificates",
				"Manage a Certificate",
			},
			ListIndex: 0,
		},
		cTypes: ListModel{
			Options: []string{
				"SSL",
				"No SSL",
			},
			ListIndex: 0,
		},
		setups: ListModel{
			Options: []string{
				"Websocket",
				"Default",
			},
			ListIndex: 0,
		},
		textInput:  ti,
		filePicker: fp,
	}
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmdS []tea.Cmd
		cmd  tea.Cmd
	)

	m.textInput, cmd = m.textInput.Update(msg)
	cmdS = append(cmdS, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch m.state {
		case MainList:
			menu := m.mainMenu
			switch key {
			case "q":
				return m, tea.Quit
			case "up", "w":
				if menu.ListIndex > 0 {
					m.mainMenu.ListIndex--
				}
			case "down", "s":
				if menu.ListIndex < len(menu.Options)-1 {
					m.mainMenu.ListIndex++
				}
			case "enter":
				switch menu.Options[menu.ListIndex] {
				case "Install Requirements":
					m.state = InstallRequirements
				case "Nginx Management":
					m.state = NginxManagement
				case "Firewall Management":
					m.state = FirewallManagement
				case "Certificate Management":
					m.state = CertificateManagement
				case "Reinstall everything":
					m.state = ReinstallEverything
				case "Uninstall and delete everything":
					m.state = UninstallAndDeleteEverything
				}
			}
		case InstallRequirements:
			menu := m.mainMenu
			switch key {
			case "q":
				return m, tea.Quit
			case "b":
				m.state = MainList
			case "up", "w":
				if menu.ListIndex > 0 {
					m.mainMenu.ListIndex--
				}
			case "down", "s":
				if menu.ListIndex < len(menu.Options)-1 {
					m.mainMenu.ListIndex++
				}
			case "enter":

			}

		case NginxManagement:
			menu := m.nginxMenu
			switch key {
			case "q":
				return m, tea.Quit
			case "b":
				m.state = MainList
			case "up", "w":
				if menu.ListIndex > 0 {
					m.nginxMenu.ListIndex--
				}
			case "down", "s":
				if menu.ListIndex < len(menu.Options)-1 {
					m.nginxMenu.ListIndex++
				}
			case "enter":
				switch menu.Options[menu.ListIndex] {
				case "Install Nginx":
					m.state = InstallNginx
				case "Delete Nginx":
					m.textInput.SetValue("")
					m.textInput.Focus()
					m.state = DeleteNginx
				case "Add Configs":
					m.textInput.SetValue("")
					m.textInput.Focus()
					m.state = ConfigName
				case "Manage Configs":
					m.state = ManageConfigs
				}
			}

		case FirewallManagement:
			menu := m.firewallMenu
			switch key {
			case "q":
				return m, tea.Quit
			case "b":
				m.state = MainList
			case "up", "w":
				if menu.ListIndex > 0 {
					m.firewallMenu.ListIndex--
				}
			case "down", "s":
				if menu.ListIndex < len(menu.Options)-1 {
					m.firewallMenu.ListIndex++
				}
			case "enter":

			}

		case CertificateManagement:
			menu := m.nginxMenu
			switch key {
			case "q":
				return m, tea.Quit
			case "b":
				m.state = MainList
			case "up", "w":
				if menu.ListIndex > 0 {
					m.nginxMenu.ListIndex--
				}
			case "down", "s":
				if menu.ListIndex < len(menu.Options)-1 {
					m.nginxMenu.ListIndex++
				}
			case "enter":

			}

		case ReinstallEverything:
			switch key {
			case "q":
				return m, tea.Quit
			case "b":
				m.state = MainList
			case "enter":

			}
		case UninstallAndDeleteEverything:
			switch key {
			case "q":
				return m, tea.Quit
			case "b":
				m.state = MainList
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
				value := m.textInput.Value()
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
				value := m.textInput.Value()
				if value != "" {
					m.newConfig.Name = value
					configFilePath := filepath.Join(configsBasePath, value+".conf")
					if common.FileExists(configFilePath) {
						m.newConfig.DuplicateName = true
					} else {
						m.newConfig.DuplicateName = false
						m.SetState(Setup, nil)
					}

				}
			}
		case Setup:
			menu := m.setups
			switch key {
			case "q":
				return m, tea.Quit
			case "b":
				m.SetState(NginxManagement, nil)
			case "up", "w":
				if menu.ListIndex > 0 {
					m.setups.ListIndex--
				}
			case "down", "s":
				if menu.ListIndex < len(menu.Options)-1 {
					m.setups.ListIndex++
				}
			case "enter":
				m.newConfig.Setup = menu.Options[menu.ListIndex]
				m.textInput.SetValue("")
				m.textInput.Focus()
				m.SetState(Upstreams, nil)
			}
		case Upstreams:
			switch key {
			case "ctrl+c":
				return m, tea.Quit
			case "ctrl+b":
				m.SetState(NginxManagement, nil)
			case "enter":
				value := m.textInput.Value()
				if value != "" {
					m.newConfig.Upstreams = strings.Fields(value)
					m.SetState(CType, nil)
				}
			}
		case CType:
			menu := m.cTypes
			switch key {
			case "q":
				return m, tea.Quit
			case "b":
				m.SetState(NginxManagement, nil)
			case "up", "w":
				if menu.ListIndex > 0 {
					m.cTypes.ListIndex--
				}
			case "down", "s":
				if menu.ListIndex < len(menu.Options)-1 {
					m.cTypes.ListIndex++
				}
			case "enter":
				m.newConfig.CType = menu.Options[menu.ListIndex]
				if m.newConfig.CType == "SSL" {
					certs, logMsg := common.Certificates(CertBasePath)
					m.certs = CertListModel{Options: certs}
					m.SetState(SelectCert, &logMsg)
				} else {
					m.textInput.SetValue("")
					m.textInput.Focus()
					m.SetState(ServerIp, nil)
				}
			}
		case SelectCert:
			menu := m.certs
			switch key {
			case "q":
				return m, tea.Quit
			case "b":
				m.SetState(NginxManagement, nil)
			case "up", "w":
				if menu.ListIndex > 0 {
					m.cTypes.ListIndex--
				}
			case "down", "s":
				if menu.ListIndex < len(menu.Options)-1 {
					m.cTypes.ListIndex++
				}
			case "enter":
				m.newConfig.CertName = common.GetKeyByIndex(m.certs.Options, m.certs.ListIndex)
				domains, _ := common.ExtractDomains(CertBasePath + m.newConfig.CertName + ".crt")
				m.domains = ListModel{Options: domains}
				m.SetState(Domains, nil)
			}
		case Domains:
			menu := m.domains
			switch key {
			case "q":
				return m, tea.Quit
			case "b":
				m.SetState(NginxManagement, nil)
			case "up", "w":
				if menu.ListIndex > 0 {
					m.domains.ListIndex--
				}
			case "down", "s":
				if menu.ListIndex < len(menu.Options)-1 {
					m.domains.ListIndex++
				}
			case "enter":
				m.newConfig.Domain = m.domains.Options[m.domains.ListIndex]
				m.textInput.SetValue("")
				m.textInput.Focus()
				m.SetState(HttpPort, nil)
			}
		case ServerIp:
			switch key {
			case "ctrl+c":
				return m, tea.Quit
			case "ctrl+b":
				m.SetState(NginxManagement, nil)
			case "enter":
				value := m.textInput.Value()
				if value != "" {
					m.newConfig.ServerIp = value
					m.textInput.SetValue("")
					m.textInput.Focus()
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
				value := m.textInput.Value()
				if value != "" {
					m.newConfig.HttpPort = value
					m.textInput.SetValue("")
					m.textInput.Focus()
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
				value := m.textInput.Value()
				if value != "" {
					m.newConfig.HttpsPort = value
					ch := make(chan common.LogMsg)
					go configure.Configure(
						configsBasePath,
						CertBasePath,
						m.newConfig.Name,
						m.newConfig.Setup,
						m.newConfig.Upstreams,
						m.newConfig.CType,
						m.newConfig.CertName,
						m.newConfig.Domain,
						m.newConfig.ServerIp,
						m.newConfig.HttpPort,
						m.newConfig.HttpsPort,
						ch,
					)
					for newLog := range ch {
						m.logs = append(m.logs, newLog)
					}
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

	}

	return m, tea.Batch(cmdS...)
}

func (m *model) SetState(s state, log *common.LogMsg) {
	m.state = s
	if log != nil {
		m.logs = append(m.logs, *log)
	} else {
		m.logs = nil
	}
}

func (m *model) View() string {

	var sb strings.Builder

	switch m.state {
	case MainList:
		sb.WriteString(buildListItems(m.mainMenu))
	case InstallRequirements:
		sb.WriteString(itemStyle.Render("Install Requirements"))
	case NginxManagement:
		sb.WriteString(buildListItems(m.nginxMenu))

	case FirewallManagement:
		sb.WriteString(buildListItems(m.firewallMenu))

	case CertificateManagement:
		sb.WriteString(buildListItems(m.certificateMenu))

	case ReinstallEverything:
		sb.WriteString(itemStyle.Render("Reinstall Everything"))
	case UninstallAndDeleteEverything:
		sb.WriteString(itemStyle.Render("Uninstall And Delete Everything"))
	//-----------------------------------------------------------------------------------
	case InstallNginx:
		sb.WriteString(itemStyle.Render("Install Nginx"))
	case DeleteNginx:
		sb.WriteString("Do you really want to uninstall nginx? (yes/y to confirm, no/n to cancel):\n" + m.textInput.View() + "\n")
	case ConfigName:
		text := "Please enter a unique name for config file. Previous configs are shown below:\n"
		existingConfigs, _ := filepath.Glob(filepath.Join(configsBasePath, "*.conf"))
		for _, cfg := range existingConfigs {
			text += cfg + "\n"
		}
		text += m.textInput.View() + "\n"
		if m.newConfig.DuplicateName {
			text += "The name you entered already exists."
		}
		sb.WriteString(text)
	case Setup:
		sb.WriteString(buildListItems(m.setups))
	case Upstreams:
		sb.WriteString("Please enter the list of upstream IP addresses (space separated):\n" + m.textInput.View() + "\n")
	case CType:
		sb.WriteString(buildListItems(m.cTypes))
	case SelectCert:
		sb.WriteString(buildCertListItems(m.certs))
	case Domains:
		sb.WriteString(buildListItems(m.domains))
	case ServerIp:
		sb.WriteString("Please enter the ip of this server:\n" + m.textInput.View() + "\n")
	case HttpPort:
		sb.WriteString("Please enter http port (80 is default):\n" + m.textInput.View() + "\n")
	case HttpsPort:
		sb.WriteString("Please enter https port (443 is default):\n" + m.textInput.View() + "\n")
	case ManageConfigs:
		sb.WriteString(itemStyle.Render("Manage Configs"))

	}

	for _, logMsg := range m.logs {
		if logMsg.Color != "" {
			sb.WriteString(itemStyle.Foreground(lipgloss.Color(logMsg.Color)).Render(logMsg.Msg + "\n"))
		} else {
			sb.WriteString(itemStyle.Foreground(lipgloss.Color(common.White)).Render(logMsg.Msg + "\n"))
		}
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
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal("Error: " + err.Error())
	}
}
