package output

import "fmt"

type Result struct {
	Installed int
	Removed   int
}

func Format(r *Result) string {
	return fmt.Sprintf("Installed %d packages, removed %d packages", r.Installed, r.Removed)
}
