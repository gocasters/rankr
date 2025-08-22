package services

func stringsTrim(s string) string {
	return trimSpace(s)
}

func stringsTrimPtr(p *string) *string {
	if p == nil {
		return nil
	}
	t := trimSpace(*p)
	if t == "" {
		return nil
	}
	return &t
}

func trimSpace(s string) string {
	i := 0
	j := len(s)
	for i < j && (s[i] == ' ' || s[i] == '\n' || s[i] == '\t' || s[i] == '\r') {
		i++
	}
	for i < j && (s[j-1] == ' ' || s[j-1] == '\n' || s[j-1] == '\t' || s[j-1] == '\r') {
		j--
	}
	if i == 0 && j == len(s) {
		return s
	}
	return s[i:j]
import "strings"
func stringsTrim(s string) string {
	return strings.TrimSpace(s)
}

func stringsTrimPtr(p *string) *string {
	if p == nil {
		return nil
	}
	t := strings.TrimSpace(*p)
	if t == "" {
		return nil
	}
	return &t
}
