package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/dan-3200/hyprprofile/internal/application/dto"
	"github.com/dan-3200/hyprprofile/internal/domain/profile"
)

type fakeProfileLister struct {
	calls int
	items []dto.ProfileListItem
}

func (f *fakeProfileLister) ListProfiles() ([]dto.ProfileListItem, error) {
	f.calls++
	return f.items, nil
}

func TestDashboardLoadsAndRefreshesFromService(t *testing.T) {
	item := profile.NewManifest("work", "Work", "", profile.DefaultEntrypoint, time.Now())
	source := &fakeProfileLister{items: []dto.ProfileListItem{{Manifest: item, Active: true}}}
	dashboard := model{profilesSource: source}
	loaded, ok := dashboard.Init()().(loadedMessage)
	if !ok {
		t.Fatal("initial command did not load profiles")
	}
	updated, _ := dashboard.Update(loaded)
	dashboard, ok = updated.(model)
	if !ok {
		t.Fatal("update returned an unexpected model")
	}
	if source.calls != 1 || !strings.Contains(dashboard.View().Content, "Work (work)") {
		t.Fatalf("dashboard did not load profile: calls=%d", source.calls)
	}
	source.items = nil
	loaded, ok = dashboard.load().(loadedMessage)
	if !ok {
		t.Fatal("refresh did not load profiles")
	}
	updated, _ = dashboard.Update(loaded)
	dashboard, ok = updated.(model)
	if !ok {
		t.Fatal("refresh returned an unexpected model")
	}
	if source.calls != 2 || !strings.Contains(dashboard.View().Content, "Nenhum perfil criado") {
		t.Fatalf("dashboard did not refresh: calls=%d", source.calls)
	}
}
