package helper

func In_Array(val string, resources []string) bool {
	for _, r := range resources {
		if r == val {
			return true
		}
	}
	return false
}
