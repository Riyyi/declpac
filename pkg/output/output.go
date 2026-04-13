package output

import (
	"fmt"
	"strings"
)

type Result struct {
	Installed int
	Removed   int
	ToInstall []string
	ToRemove  []string
}

func Format(r *Result) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Installed %d packages, removed %d packages", r.Installed, r.Removed))
	if len(r.ToInstall) > 0 {
		b.WriteString("\nWould install: ")
		b.WriteString(strings.Join(r.ToInstall, ", "))
	}
	if len(r.ToRemove) > 0 {
		b.WriteString("\nWould remove: ")
		b.WriteString(strings.Join(r.ToRemove, ", "))
	}
	return b.String()
}
