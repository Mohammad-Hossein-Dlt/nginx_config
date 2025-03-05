package nginx

import (
	tea "github.com/charmbracelet/bubbletea"
	"nginx_configure/common"
	"os"
	"os/exec"
)

func Delete() tea.Cmd {

	if _, err := exec.LookPath("nginx"); err == nil {

		return tea.Sequence(
			common.LogMessage("Nginx is installed. Purging existing installation and configuration files...", common.Gold),
			common.LogMessage("Stopping nginx service...", common.Gold),
			common.RunCommandWithLogs("systemctl stop nginx"),
			common.LogMessage("Purging nginx...", common.Gold),
			common.RunCommandWithLogs("apt-get purge -y nginx"),
			common.LogMessage("Auto removing packages...", common.Gold),
			common.RunCommandWithLogs("apt-get autoremove -y"),
			common.LogMessage("Removing nginx directory...", common.Gold),
			func() tea.Msg {
				err := os.RemoveAll("/etc/nginx")
				if err != nil {
					return common.LogItem{Msg: "An Error occurred : " + err.Error(), Color: common.Red}
				}
				return common.LogItem{Msg: "All is done.", Color: common.Green}
			},
		)

	}

	return tea.Sequence(
		common.LogMessage("Nginx is not installed.", common.Blue),
	)
}
