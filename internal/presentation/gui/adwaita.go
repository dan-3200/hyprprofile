//go:build gui

package gui

/*
#cgo pkg-config: libadwaita-1
#include <adwaita.h>

static void hyprprofile_adwaita_init(void) {
	adw_init();
}
*/
import "C"

func initializeAdwaita() {
	C.hyprprofile_adwaita_init()
}
