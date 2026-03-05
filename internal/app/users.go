package app

import "strings"

type UserFilterOptions struct {
	Domains            []string
	IncludeGuests      bool
	IncludeDeactivated bool
}

func FilterUsers(users []map[string]any, opts UserFilterOptions) []map[string]any {
	domainSet := map[string]struct{}{}
	for _, d := range opts.Domains {
		n := normalizeDomain(d)
		if n != "" {
			domainSet[n] = struct{}{}
		}
	}

	filtered := make([]map[string]any, 0, len(users))
	for _, user := range users {
		if !opts.IncludeDeactivated {
			if active, ok := user["is_active"].(bool); ok && !active {
				continue
			}
		}
		if !opts.IncludeGuests {
			if guest, ok := user["is_guest"].(bool); ok && guest {
				continue
			}
		}
		if len(domainSet) > 0 {
			email, _ := user["email"].(string)
			domain := normalizeDomain(emailDomain(email))
			if _, ok := domainSet[domain]; !ok {
				continue
			}
		}
		filtered = append(filtered, user)
	}
	return filtered
}

func normalizeDomain(v string) string {
	return strings.ToLower(strings.TrimSpace(strings.TrimPrefix(v, "@")))
}

func emailDomain(email string) string {
	parts := strings.Split(strings.TrimSpace(email), "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}
