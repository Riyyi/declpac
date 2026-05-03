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
	b.WriteString(fmt.Sprintf("installed %d packages, removed %d packages", r.Installed, r.Removed))
	if len(r.ToInstall) > 0 {
		b.WriteString("\nwould install: ")
		b.WriteString(strings.Join(r.ToInstall, ", "))
	}
	if len(r.ToRemove) > 0 {
		b.WriteString("\nwould remove: ")
		b.WriteString(strings.Join(r.ToRemove, ", "))
	}
	return b.String()
}
