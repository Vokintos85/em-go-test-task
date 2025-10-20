package api

import (
	"testing"
	"time"
)

func TestParseMonth(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		year    int
		month   time.Month
		wantErr bool
	}{
		{"dash format", "2024-03", 2024, time.March, false},
		{"month first", "03-2024", 2024, time.March, false},
		{"invalid", "2024/03", 0, 0, true},
		{"empty", "", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMonth(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Year() != tt.year || got.Month() != tt.month || got.Day() != 1 {
				t.Fatalf("unexpected time: %v", got)
			}
			if got.Location() != time.UTC {
				t.Fatalf("expected UTC location, got %v", got.Location())
			}
		})
	}
}
