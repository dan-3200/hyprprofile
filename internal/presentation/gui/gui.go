//go:build gui

package gui

import (
	"fmt"
	"os"

	"github.com/dan-3200/hyprprofile/internal/application/profiles"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// Run opens the native GTK 4 and libadwaita profile dashboard.
func Run(controller ProfileController) error {
	initializeAdwaita()
	app := gtk.NewApplication("io.github.dan3200.hyprprofile", gio.ApplicationFlagsNone)
	var activeDashboard *dashboard
	app.ConnectActivate(func() {
		activeDashboard = newDashboard(app, controller)
		activeDashboard.present()
	})
	if code := app.Run(os.Args); code != 0 {
		return fmt.Errorf("run graphical interface: exited with status %d", code)
	}
	return nil
}

type dashboard struct {
	controller ProfileController
	window     *gtk.ApplicationWindow
	list       *gtk.ListBox
	details    *gtk.Label
	status     *gtk.Label
	activate   *gtk.Button
	items      []profiles.ProfileListItem
	rows       []*gtk.ListBoxRow
	selected   int
}

func newDashboard(app *gtk.Application, controller ProfileController) *dashboard {
	dashboard := &dashboard{controller: controller}
	dashboard.window = gtk.NewApplicationWindow(app)
	dashboard.window.SetTitle("HyprProfile")
	dashboard.window.SetDefaultSize(840, 560)

	header := gtk.NewHeaderBar()
	header.SetTitleWidget(gtk.NewLabel("HyprProfile"))
	dashboard.window.SetTitlebar(header)

	content := gtk.NewBox(gtk.OrientationHorizontal, 0)
	dashboard.list = gtk.NewListBox()
	dashboard.list.SetSizeRequest(280, -1)
	dashboard.list.ConnectSelectedRowsChanged(func() {
		dashboard.selectRow(dashboard.list.SelectedRow())
	})
	scroll := gtk.NewScrolledWindow()
	scroll.SetChild(dashboard.list)
	content.Append(scroll)

	detailBox := gtk.NewBox(gtk.OrientationVertical, 16)
	detailBox.SetMarginTop(24)
	detailBox.SetMarginBottom(24)
	detailBox.SetMarginStart(24)
	detailBox.SetMarginEnd(24)
	detailBox.SetHExpand(true)
	dashboard.details = gtk.NewLabel("")
	dashboard.details.SetXAlign(0)
	dashboard.details.SetWrap(true)
	detailBox.Append(dashboard.details)
	dashboard.activate = gtk.NewButtonWithLabel("Selecionar para o próximo login")
	dashboard.activate.SetHAlign(gtk.AlignStart)
	dashboard.activate.ConnectClicked(dashboard.activateSelected)
	detailBox.Append(dashboard.activate)
	dashboard.status = gtk.NewLabel("")
	dashboard.status.SetXAlign(0)
	dashboard.status.SetWrap(true)
	detailBox.Append(dashboard.status)
	content.Append(detailBox)
	dashboard.window.SetChild(content)
	dashboard.reload()
	return dashboard
}

func (d *dashboard) present() {
	d.window.Present()
}

func (d *dashboard) reload() {
	items, err := d.controller.ListProfiles()
	if err != nil {
		d.items = nil
		d.status.SetText("Não foi possível carregar perfis: " + err.Error())
		d.render()
		return
	}
	d.items = items
	d.selected = 0
	d.status.SetText("")
	d.render()
}

func (d *dashboard) render() {
	for _, row := range d.rows {
		d.list.Remove(row)
	}
	d.rows = nil
	for index, item := range d.items {
		row := gtk.NewListBoxRow()
		label := item.Manifest.Name + "\n" + item.Manifest.DirectoryID()
		if item.Active {
			label += " · ativo"
		}
		text := gtk.NewLabel(label)
		text.SetXAlign(0)
		text.SetMarginTop(12)
		text.SetMarginBottom(12)
		text.SetMarginStart(12)
		text.SetMarginEnd(12)
		row.SetChild(text)
		d.list.Append(row)
		d.rows = append(d.rows, row)
		if index == d.selected {
			d.list.SelectRow(row)
		}
	}
	d.renderDetails()
}

func (d *dashboard) selectRow(row *gtk.ListBoxRow) {
	if row == nil {
		return
	}
	d.selected = row.Index()
	d.renderDetails()
}

func (d *dashboard) renderDetails() {
	if len(d.items) == 0 {
		d.details.SetText("Nenhum perfil criado. Use a CLI para criar ou importar um perfil.")
		d.activate.SetSensitive(false)
		return
	}
	item := d.items[d.selected]
	manifest := item.Manifest
	description := manifest.Description
	if description == "" {
		description = "Sem descrição"
	}
	template := manifest.Template
	if template == "" {
		template = "Não informado"
	}
	state := "Inativo"
	if item.Active {
		state = "Ativo no próximo login"
	}
	d.details.SetText(fmt.Sprintf("%s\n\nEstado: %s\nID: %s\nTemplate: %s\nEntrada: %s\n\n%s", manifest.Name, state, manifest.DirectoryID(), template, manifest.Entrypoint, description))
	d.activate.SetSensitive(!item.Active)
}

func (d *dashboard) activateSelected() {
	if len(d.items) == 0 {
		return
	}
	id := d.items[d.selected].Manifest.DirectoryID()
	if _, err := d.controller.ActivateProfile(id); err != nil {
		d.status.SetText("Não foi possível selecionar o perfil: " + err.Error())
		return
	}
	d.reload()
	d.status.SetText(fmt.Sprintf("Perfil %q selecionado para o próximo login.", id))
}
