package merge

func Merge(packages map[string]bool) []string {
	result := make([]string, 0, len(packages))
	for name := range packages {
		result = append(result, name)
	}
	return result
}
