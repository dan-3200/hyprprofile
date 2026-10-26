package profile

// Severity distinguishes blocking errors from warnings.
type Severity string

const (
	// SeverityError blocks profile activation.
	SeverityError Severity = "error"
	// SeverityWarning describes a non-blocking issue.
	SeverityWarning Severity = "warning"
)

// Issue records one profile validation finding.
type Issue struct {
	Severity Severity
	Message  string
}

// Result contains findings from independent validation checks.
type Result struct {
	Issues []Issue
}

// Valid reports whether there are no blocking errors.
func (r Result) Valid() bool {
	for _, issue := range r.Issues {
		if issue.Severity == SeverityError {
			return false
		}
	}
	return true
}
