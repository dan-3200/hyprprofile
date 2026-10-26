package launch

// Plan is the validated process invocation for a Hyprland session.
type Plan struct {
	Executable string
	Arguments  []string
}
