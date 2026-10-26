package tui

import (
	"testing"
	"time"

	"github.com/dan-3200/hyprprofile/internal/application/profiles"
	"github.com/dan-3200/hyprprofile/internal/domain/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeProfileLister struct {
	calls int
	items []profiles.ProfileListItem
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
