package developer

import (
	"errors"
	"testing"
)

func TestNormalizeReadOnlyQueryAllowsReadOnlyStatements(t *testing.T) {
	t.Parallel()

	tests := []string{
		"select * from restaurants",
		"SELECT count(*) FROM menu_items;",
		"-- inspect menus\nselect name from menu_items",
		"/* schema */ show statement_timeout",
		"with items as (select 1 as n) select n from items",
		"explain select * from restaurants",
	}

	for _, input := range tests {
		input := input
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			got, err := NormalizeReadOnlyQuery(input)
			if err != nil {
				t.Fatalf("NormalizeReadOnlyQuery(%q) error = %v", input, err)
			}
			if got == "" || got[len(got)-1] == ';' {
				t.Fatalf("normalized query = %q, want non-empty query without trailing semicolon", got)
			}
		})
	}
}

func TestNormalizeReadOnlyQueryRejectsMutationsAndMultipleStatements(t *testing.T) {
	t.Parallel()

	tests := []string{
		"",
		"update restaurants set status = 'lead'",
		"delete from menu_items",
		"drop table restaurants",
		"select 1; select 2",
	}

	for _, input := range tests {
		input := input
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			if _, err := NormalizeReadOnlyQuery(input); err == nil {
				t.Fatalf("NormalizeReadOnlyQuery(%q) error = nil, want error", input)
			}
		})
	}
}

func TestNormalizeReadOnlyQueryRejectsLongQuery(t *testing.T) {
	t.Parallel()

	query := make([]byte, maxQueryLength+1)
	for i := range query {
		query[i] = 'x'
	}
	if _, err := NormalizeReadOnlyQuery(string(query)); !errors.Is(err, ErrQueryTooLong) {
		t.Fatalf("NormalizeReadOnlyQuery(long) error = %v, want ErrQueryTooLong", err)
	}
}
