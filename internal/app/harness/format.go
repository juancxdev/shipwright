package harness

func formatYesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func SortedUniqueForDisplay(values []string) []string {
	return sortedUnique(values)
}
