package tui

import (
	"fmt"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/dan-3200/hyprprofile/internal/application/profiles"
	"github.com/dan-3200/hyprprofile/internal/domain/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeProfileLister struct {
	calls       int
	activatedID string
	items       []profiles.ProfileListItem
}

func (f *fakeProfileLister) ActivateProfile(id string) (profile.Manifest, error) {
	f.activatedID = id
	for index := range f.items {
		f.items[index].Active = f.items[index].Manifest.DirectoryID() == id
		if f.items[index].Active {
			return f.items[index].Manifest, nil
		}
	}
	return profile.Manifest{}, fmt.Errorf("profile %q not found", id)
}

func (f *fakeProfileLister) ListProfiles() ([]profiles.ProfileListItem, error) {
	f.calls++
	return f.items, nil
}

func TestDashboardLoadsAndRefreshesFromService(t *testing.T) {
	item := profile.NewManifest("work", "Work", "", profile.DefaultEntrypoint, time.Now())
	source := &fakeProfileLister{items: []profiles.ProfileListItem{{Manifest: item, Active: true}}}
	dashboard := model{profilesSource: source}
	loaded, ok := dashboard.Init()().(loadedMessage)
	require.True(t, ok, "initial command did not load profiles")
	updated, _ := dashboard.Update(loaded)
	dashboard, ok = updated.(model)
	require.True(t, ok, "update returned an unexpected model")
	assert.Equal(t, 1, source.calls)
	assert.Contains(t, dashboard.View().Content, "Work (work)")
	source.items = nil
	loaded, ok = dashboard.load().(loadedMessage)
	require.True(t, ok, "refresh did not load profiles")
	updated, _ = dashboard.Update(loaded)
	dashboard, ok = updated.(model)
	require.True(t, ok, "refresh returned an unexpected model")
	assert.Equal(t, 2, source.calls)
	assert.Contains(t, dashboard.View().Content, "Nenhum perfil criado")
}

func TestDashboardNavigationDetailsAndHelp(t *testing.T) {
	first := profile.NewManifest("work", "Work", "Configuração de trabalho", profile.DefaultEntrypoint, time.Now())
	first.Template = "minimal"
	second := profile.NewManifest("home", "Home", "", profile.DefaultEntrypoint, time.Now())
	dashboard := model{profiles: []profiles.ProfileListItem{{Manifest: first, Active: true}, {Manifest: second}}}

	updated, _ := dashboard.Update(keyMessage("j", 'j'))
	dashboard, ok := updated.(model)
	require.True(t, ok, "navigation returned an unexpected model")
	assert.Equal(t, 1, dashboard.cursor)
	assert.Contains(t, dashboard.View().Content, "Perfil selecionado")
	assert.Contains(t, dashboard.View().Content, "Home")
	assert.Contains(t, dashboard.View().Content, "Sem descrição")

	updated, _ = dashboard.Update(keyMessage("g", 'g'))
	dashboard, ok = updated.(model)
	require.True(t, ok, "first-item navigation returned an unexpected model")
	assert.Equal(t, 0, dashboard.cursor)
	assert.Contains(t, dashboard.View().Content, "Configuração de trabalho")

	updated, _ = dashboard.Update(keyMessage("h", 'h'))
	dashboard, ok = updated.(model)
	require.True(t, ok, "help toggle returned an unexpected model")
	assert.True(t, dashboard.showHelp)
	assert.Contains(t, dashboard.View().Content, "Atalhos")
}

func TestDashboardActivatesSelectedProfile(t *testing.T) {
	first := profile.NewManifest("work", "Work", "", profile.DefaultEntrypoint, time.Now())
	second := profile.NewManifest("home", "Home", "", profile.DefaultEntrypoint, time.Now())
	source := &fakeProfileLister{items: []profiles.ProfileListItem{{Manifest: first, Active: true}, {Manifest: second}}}
	dashboard := model{profilesSource: source, profiles: source.items, cursor: 1}

	updated, command := dashboard.Update(keyMessage("enter", tea.KeyEnter))
	dashboard, ok := updated.(model)
	require.True(t, ok)
	require.NotNil(t, command)
	message := command()
	updated, command = dashboard.Update(message)
	dashboard, ok = updated.(model)
	require.True(t, ok)
	require.NotNil(t, command)

	loaded, ok := command().(loadedMessage)
	require.True(t, ok)
	updated, _ = dashboard.Update(loaded)
	dashboard, ok = updated.(model)
	require.True(t, ok)
	assert.Equal(t, "home", source.activatedID)
	assert.True(t, dashboard.profiles[1].Active)
	assert.False(t, dashboard.profiles[0].Active)
	assert.Contains(t, dashboard.View().Content, "selecionado para o próximo login")
}

func keyMessage(text string, code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Text: text, Code: code})
}
