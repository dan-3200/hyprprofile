package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dan-3200/hyprprofile/internal/app"
)

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	activeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	mutedStyle   = lipgloss.NewStyle().Faint(true)
	quitKey      = key.NewBinding(key.WithKeys("q", "ctrl+c"))
	refreshKey   = key.NewBinding(key.WithKeys("r"))
	selectedMark = "> "
)

type model struct {
	profilesSource ProfileLister
	profiles       []app.ProfileListItem
	cursor         int
	err            error
}

type loadedMessage struct {
	profiles []app.ProfileListItem
	err      error
}

// ProfileLister provides the read model displayed in the dashboard.
type ProfileLister interface {
	ListProfiles() ([]app.ProfileListItem, error)
}

// Run opens the initial dashboard. CLI subcommands remain available for scripts.
func Run(profiles ProfileLister) error {
	program := tea.NewProgram(model{profilesSource: profiles})
	_, err := program.Run()
	if err != nil {
		return fmt.Errorf("run terminal interface: %w", err)
	}
	return nil
}

func (m model) Init() tea.Cmd {
	return m.load
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case loadedMessage:
		m.profiles = message.profiles
		m.err = message.err
		if m.cursor >= len(m.profiles) {
			m.cursor = max(0, len(m.profiles)-1)
		}
		return m, nil
	case tea.KeyPressMsg:
		switch {
		case key.Matches(message, quitKey):
			return m, tea.Quit
		case key.Matches(message, refreshKey):
			return m, m.load
		case message.String() == "up" || message.String() == "k":
			m.cursor = max(0, m.cursor-1)
		case message.String() == "down" || message.String() == "j":
			m.cursor = min(len(m.profiles)-1, m.cursor+1)
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	lines := []string{titleStyle.Render("HyprProfile"), ""}
	if m.err != nil {
		lines = append(lines, errorStyle.Render("Erro: "+m.err.Error()))
		return tea.NewView(strings.Join(lines, "\n"))
	}
	if len(m.profiles) == 0 {
		lines = append(lines, "Nenhum perfil criado.", mutedStyle.Render("Use: hyprprofile create <id> --name <nome>"))
	} else {
		for index, item := range m.profiles {
			prefix := "  "
			if index == m.cursor {
				prefix = selectedMark
			}
			label := fmt.Sprintf("%s (%s)", item.Manifest.Name, item.Manifest.DirectoryID())
			if item.Active {
				label = activeStyle.Render(label + "  ativo")
			}
			lines = append(lines, prefix+label)
		}
	}
	lines = append(lines, "", mutedStyle.Render("↑/↓ ou j/k: navegar • r: atualizar • q: sair"))
	return tea.NewView(strings.Join(lines, "\n"))
}

func (m model) load() tea.Msg {
	profiles, err := m.profilesSource.ListProfiles()
	return loadedMessage{profiles: profiles, err: err}
}
