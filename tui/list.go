//package tui
//
//import (
//	"fmt"
//	"github.com/charmbracelet/bubbles/list"
//	tea "github.com/charmbracelet/bubbletea"
//	"github.com/charmbracelet/lipgloss"
//	"io"
//
//	"os"
//	"strings"
//)
//
//const listHeight = 14
//
//var (
//	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
//	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
//	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
//	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
//	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
//	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
//)
//
//type item string
//
//func (i item) FilterValue() string { return string(i) }
//
//type itemDelegate struct{}
//
//func (d itemDelegate) Height() int                             { return 1 }
//func (d itemDelegate) Spacing() int                            { return 0 }
//func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
//func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
//	i, ok := listItem.(item)
//	if !ok {
//		return
//	}
//
//	str := fmt.Sprintf("%d. %s", index+1, i)
//
//	fn := itemStyle.Render
//	if index == m.Index() {
//		fn = func(s ...string) string {
//			return selectedItemStyle.Render("> " + strings.Join(s, " "))
//		}
//	}
//
//	_, err := fmt.Fprint(w, fn(str))
//	if err != nil {
//		fmt.Println("An Error occurred : " + err.Error())
//		os.Exit(1)
//	}
//}
//
//type model struct {
//	list     list.Model
//	choice   string
//	quitting bool
//}
//
//func (m model) Init() tea.Cmd {
//	return nil
//}
//
//func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	switch msg := msg.(type) {
//	case tea.WindowSizeMsg:
//		m.list.SetWidth(msg.Width)
//		return m, nil
//
//	case tea.KeyMsg:
//		switch keypress := msg.String(); keypress {
//		case "q", "ctrl+c":
//			m.quitting = true
//			return m, tea.Quit
//
//		case "enter":
//			i, ok := m.list.SelectedItem().(item)
//			if ok {
//				m.choice = string(i)
//			}
//			return m, tea.Quit
//		}
//	}
//
//	var cmd tea.Cmd
//	m.list, cmd = m.list.Update(msg)
//	return m, cmd
//}
//
//func (m model) View() string {
//	if m.choice != "" {
//		return m.choice
//	}
//	if m.quitting {
//		return quitTextStyle.Render("CLI Closed.")
//	}
//	return "\n" + m.list.View()
//}
//
//func ListUi(options []string) {
//	items := make([]list.Item, len(options))
//	for i := range options {
//		items[i] = item(options[i])
//	}
//
//	const defaultWidth = 20
//
//	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
//	l.Title = "Avida cli"
//	l.SetShowStatusBar(false)
//	l.SetFilteringEnabled(false)
//	l.Styles.Title = titleStyle
//	l.Styles.PaginationStyle = paginationStyle
//	l.Styles.HelpStyle = helpStyle
//
//	m := model{list: l}
//	fmt.Print("/////")
//	fmt.Print(m.choice)
//	fmt.Print("/////")
//
//	if _, err := tea.NewProgram(m).Run(); err != nil {
//		fmt.Println("Error running program:", err)
//		os.Exit(1)
//	}
//}

package tui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

// NestedItem represents an item that can have children of the same type.

type mfunc func()

type NestedItem struct {
	Title       string
	Description string
	Children    []NestedItem
	Action      mfunc
}

func (n NestedItem) FilterValue() string { return n.Title }

// itemDelegate is used to render each NestedItem in the list.
type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	n, ok := listItem.(NestedItem)
	if !ok {
		return
	}
	str := fmt.Sprintf("%d. %s", index+1, n.Title)
	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}
	_, err := fmt.Fprint(w, fn(str))
	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}
}

// the model holds the navigation stack for the dynamic list.
type model struct {
	navStack []list.Model // each level is a list.Model; top is the current level.
	choice   string       // full breadcrumb path when an item without children is selected.
	quitting bool
}

// dynamicList creates a list.Model from a slice of NestedItem.
func dynamicList(items []NestedItem, title string, width int) list.Model {
	l := list.New(makeItems(items), itemDelegate{}, width, listHeight)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	return l
}

// makeItems converts []NestedItem into []list.Item.
func makeItems(items []NestedItem) []list.Item {
	result := make([]list.Item, len(items))
	for i, itm := range items {
		result[i] = itm
	}
	return result
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If the navigation stack is empty, there's nothing to update.
	if len(m.navStack) == 0 {
		return m, nil
	}
	current := m.navStack[len(m.navStack)-1]

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		current.SetWidth(msg.Width)
		m.navStack[len(m.navStack)-1] = current
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			// In the current list, get the selected item.
			sel, ok := current.SelectedItem().(NestedItem)
			if !ok {
				return m, nil
			}

			if sel.Action != nil {
				sel.Action()
			}
			// If the selected item has children, push a new list onto the stack.
			if len(sel.Children) > 0 {
				newList := dynamicList(sel.Children, sel.Title, current.Width())
				m.navStack = append(m.navStack, newList)
				return m, nil
			}

			// If no children exist, finalize the selection by building the full breadcrumb path.
			m.choice = buildPath(m.navStack)
			return m, tea.Quit

		case "b":
			// Pop the navigation stack (go back one level) if possible.
			if len(m.navStack) > 1 {
				m.navStack = m.navStack[:len(m.navStack)-1]
				return m, nil
			}
		}
	}

	// Update the current list.
	var cmd tea.Cmd
	current, cmd = current.Update(msg)
	m.navStack[len(m.navStack)-1] = current
	return m, cmd
}

// buildPath constructs a breadcrumb path from the navigation stack.
func buildPath(stack []list.Model) string {
	var parts []string
	for _, l := range stack {
		if itm, ok := l.SelectedItem().(NestedItem); ok {
			parts = append(parts, itm.Title)
		}
	}
	return strings.Join(parts, " > ")
}

func (m model) View() string {
	if m.choice != "" {
		return "Selected: " + m.choice
	}
	if m.quitting {
		return quitTextStyle.Render("CLI Closed.")
	}
	breadcrumb := buildPath(m.navStack)
	s := "Breadcrumb: " + breadcrumb + "\n\n"
	s += m.navStack[len(m.navStack)-1].View()
	s += "\n\nPress 'b' to go back, 'enter' to select, 'q' to quit."
	//return s
	return "\n" + m.navStack[len(m.navStack)-1].View()
}

// ListUi launches the dynamic nested list UI with the given root items.
func ListUi(rootItems []NestedItem) {
	const defaultWidth = 20
	rootList := dynamicList(rootItems, "Avida CLI", defaultWidth)
	m := model{
		navStack: []list.Model{rootList},
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
