package services

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
