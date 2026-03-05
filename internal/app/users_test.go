package app

import "testing"

func TestFilterUsers_DomainAndFlags(t *testing.T) {
	users := []map[string]any{
		{"gid": "1", "email": "a@example.com", "is_guest": false, "is_active": true},
		{"gid": "2", "email": "b@example.com", "is_guest": true, "is_active": true},
		{"gid": "3", "email": "c@other.com", "is_guest": false, "is_active": true},
		{"gid": "4", "email": "d@example.com", "is_guest": false, "is_active": false},
	}
	filtered := FilterUsers(users, UserFilterOptions{Domains: []string{"example.com"}})
	if len(filtered) != 1 {
		t.Fatalf("expected 1 user, got %d", len(filtered))
	}
	if filtered[0]["gid"] != "1" {
		t.Fatalf("expected gid=1, got %v", filtered[0]["gid"])
	}

	includeGuests := FilterUsers(users, UserFilterOptions{Domains: []string{"example.com"}, IncludeGuests: true})
	if len(includeGuests) != 2 {
		t.Fatalf("expected 2 users with guests included, got %d", len(includeGuests))
	}

	includeAll := FilterUsers(users, UserFilterOptions{Domains: []string{"example.com"}, IncludeGuests: true, IncludeDeactivated: true})
	if len(includeAll) != 3 {
		t.Fatalf("expected 3 users with guests+deactivated included, got %d", len(includeAll))
	}
}
