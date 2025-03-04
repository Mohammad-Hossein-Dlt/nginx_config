package model

//import (
//	"github.com/charmbracelet/bubbles/filepicker"
//	"github.com/charmbracelet/bubbles/textinput"
//	"nginx_configure/common"
//)
//
//type ListModel struct {
//	Options   []string
//	ListIndex int
//}
//
//type CertListModel struct {
//	Options   map[string]string
//	ListIndex int
//}
//
//type NewConfig struct {
//	DuplicateName bool
//	Name          string
//	Setup         string
//	Upstreams     []string
//	CType         string
//	CertName      string
//	Domain        string
//	ServerIp      string
//	HttpPort      string
//	HttpsPort     string
//}
//
//type CLIModel struct {
//	State State
//
//	MainMenu        ListModel
//	NginxMenu       ListModel
//	FirewallMenu    ListModel
//	CertificateMenu ListModel
//	//-------------------------
//	NewConfig NewConfig
//
//	CTypes  ListModel
//	Certs   CertListModel
//	Domains ListModel
//	Setups  ListModel
//	//-------------------------
//
//	TextInput  textinput.Model
//	FilePicker filepicker.Model
//	//-------------------------
//	Logs []common.LogMsg
//}
