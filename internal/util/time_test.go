
package util

import (
	"testing"
)

func TestToRFC3339(t *testing.T) {
	tests := []struct {
		name    string
		unixTs  string
		want    string
		wantErr bool
	}{
		{
			name:    "Valid timestamp with fractional part (spec: truncate)",
			unixTs:  "1678886400.123456",
			want:    "2023-03-15T13:20:00Z", // Specification is to truncate nano seconds
			wantErr: false,
		},
		{
			name:    "Valid timestamp without fractional part",
			unixTs:  "1678886400",
			want:    "2023-03-15T13:20:00Z",
			wantErr: false,
		},
		{
			name:    "Zero timestamp",
			unixTs:  "0",
			want:    "1970-01-01T00:00:00Z",
			wantErr: false,
		},
		{
			name:    "Empty string timestamp",
			unixTs:  "",
			want:    "",
			wantErr: false,
		},
		{
			name:    "Invalid timestamp string",
			unixTs:  "not-a-timestamp",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Invalid format with extra text",
			unixTs:  "1678886400.123 abc",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToRFC3339(tt.unixTs)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToRFC3339() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToRFC3339() = %v, want %v", got, tt.want)
			}
		})
	}
}
