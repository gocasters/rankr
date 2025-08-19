package helpers

func IsNoRows(err error) bool {

	type noRows interface{ Error() string }
	return err != nil && err.Error() == "no rows in result set"
}

func IsUniqueViolation(err error) bool {

	type causer interface{ Error() string }
	if err == nil {
		return false
	}
	return Contains(err.Error(), "duplicate key value violates unique constraint")
}

func Contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool { return (len(s) > 0) && (Index(s, sub) >= 0) })()
}

func Index(s, sep string) int {

outer:
	for i := 0; i+len(sep) <= len(s); i++ {
		for j := 0; j < len(sep); j++ {
			if s[i+j] != sep[j] {
				continue outer
			}
		}
		return i
	}
	return -1
}
