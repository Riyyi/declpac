package merge

import "errors"

var ErrEmptyList = errors.New("package list is empty")

func Merge(packages map[string]bool) ([]string, error) {
	result := make([]string, 0, len(packages))
	for name := range packages {
		result = append(result, name)
	}
	if len(result) == 0 {
		return nil, ErrEmptyList
	}
	return result, nil
}
