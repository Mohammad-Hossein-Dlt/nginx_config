package nginx

import (
	tea "github.com/charmbracelet/bubbletea"
	"nginx_configure/common"
)

func Install() tea.Cmd {
	return tea.Sequence(
		common.LogMessage("Updating package list...", common.Gold),
		common.RunCommandWithLogs("apt-get update -y"),
		common.LogMessage("Installing nginx...", common.Gold),
		common.RunCommandWithLogs("apt-get install nginx -y"),
	)
}
