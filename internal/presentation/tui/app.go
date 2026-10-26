package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dan-3200/hyprprofile/internal/application/profiles"
	"github.com/dan-3200/hyprprofile/internal/domain/profile"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	activeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	mutedStyle    = lipgloss.NewStyle().Faint(true)
	labelStyle    = lipgloss.NewStyle().Bold(true)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	quitKey       = key.NewBinding(key.WithKeys("q", "ctrl+c"))
	selectedMark  = "> "
)

type model struct {
	profilesSource ProfileController
	profiles       []profiles.ProfileListItem
	cursor         int
	err            error
	showHelp       bool
	notice         string
}

type loadedMessage struct {
	profiles []profiles.ProfileListItem
	err      error
}

type activatedMessage struct {
	id  string
	err error
}

// ProfileController provides the profile actions used by the dashboard.
type ProfileController interface {
	ListProfiles() ([]profiles.ProfileListItem, error)
	ActivateProfile(string) (profile.Manifest, error)
}

// Run opens the initial dashboard. CLI subcommands remain available for scripts.
func Run(profiles ProfileController) error {
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
	case activatedMessage:
		if message.err != nil {
			m.notice = "Não foi possível ativar o perfil: " + message.err.Error()
			return m, nil
		}
		m.notice = fmt.Sprintf("Perfil %q selecionado para o próximo login.", message.id)
		return m, m.load
	case tea.KeyPressMsg:
		return m.handleKey(message)
	}
	return m, nil
}

func (m model) handleKey(message tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(message, quitKey):
		return m, tea.Quit
	case message.String() == "h" || message.String() == "?":
		m.showHelp = !m.showHelp
	case message.String() == "enter":
		if len(m.profiles) == 0 || m.profiles[m.cursor].Active {
			return m, nil
		}
		m.notice = "Validando e selecionando perfil…"
		return m, m.activateSelected
	case message.String() == "up" || message.String() == "k":
		m.cursor = max(0, m.cursor-1)
	case message.String() == "down" || message.String() == "j":
		m.cursor = min(len(m.profiles)-1, m.cursor+1)
	case message.String() == "home" || message.String() == "g":
		m.cursor = 0
	case message.String() == "end" || message.String() == "G":
		m.cursor = max(0, len(m.profiles)-1)
	}
	return m, nil
}

func (m model) View() tea.View {
	lines := []string{titleStyle.Render("HyprProfile"), mutedStyle.Render("Perfis disponíveis para o próximo login"), ""}
	if m.err != nil {
		lines = append(lines, errorStyle.Render("Erro ao carregar perfis: "+m.err.Error()), "", mutedStyle.Render("q: sair"))
		return tea.NewView(strings.Join(lines, "\n"))
	}
	if m.notice != "" {
		lines = append(lines, activeStyle.Render(m.notice), "")
	}
	if len(m.profiles) == 0 {
		lines = append(lines, "Nenhum perfil criado.", mutedStyle.Render("Use: hyprprofile create <id> --name <nome>"))
	} else {
		lines = append(lines, labelStyle.Render(fmt.Sprintf("Perfis (%d)", len(m.profiles))))
		for index, item := range m.profiles {
			prefix := "  "
			if index == m.cursor {
				prefix = selectedMark
			}
			label := fmt.Sprintf("%s (%s)", item.Manifest.Name, item.Manifest.DirectoryID())
			if item.Active {
				label = activeStyle.Render(label + "  ativo")
			}
			if index == m.cursor {
				label = selectedStyle.Render(label)
			}
			lines = append(lines, prefix+label)
		}
		lines = append(lines, "", m.selectedProfileView())
	}
	if m.showHelp {
		lines = append(lines, "", labelStyle.Render("Atalhos"), "↑/↓ ou j/k: selecionar • g/Home: primeiro • G/End: último", "Enter: selecionar para o próximo login • h/?: ocultar ajuda • q/Ctrl+C: sair")
	} else {
		lines = append(lines, "", mutedStyle.Render("↑/↓ ou j/k: navegar • Enter: selecionar • h: ajuda • q: sair"))
	}
	return tea.NewView(strings.Join(lines, "\n"))
}

func (m model) activateSelected() tea.Msg {
	item := m.profiles[m.cursor]
	_, err := m.profilesSource.ActivateProfile(item.Manifest.DirectoryID())
	return activatedMessage{id: item.Manifest.DirectoryID(), err: err}
}

func (m model) selectedProfileView() string {
	if len(m.profiles) == 0 {
		return ""
	}
	item := m.profiles[m.cursor]
	manifest := item.Manifest
	description := manifest.Description
	if description == "" {
		description = "Sem descrição"
	}
	template := manifest.Template
	if template == "" {
		template = "não informado"
	}
	state := "inativo"
	if item.Active {
		state = "ativo no próximo login"
	}
	return strings.Join([]string{
		labelStyle.Render("Perfil selecionado"),
		fmt.Sprintf("%s %s", mutedStyle.Render("Nome:"), manifest.Name),
		fmt.Sprintf("%s %s", mutedStyle.Render("ID:"), manifest.DirectoryID()),
		fmt.Sprintf("%s %s", mutedStyle.Render("Template:"), template),
		fmt.Sprintf("%s %s", mutedStyle.Render("Entrada:"), manifest.Entrypoint),
		fmt.Sprintf("%s %s", mutedStyle.Render("Estado:"), state),
		fmt.Sprintf("%s %s", mutedStyle.Render("Descrição:"), description),
	}, "\n")
}

func (m model) load() tea.Msg {
	profiles, err := m.profilesSource.ListProfiles()
	return loadedMessage{profiles: profiles, err: err}
}
