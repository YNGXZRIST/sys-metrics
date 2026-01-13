package stringsparser

import "testing"

func TestCapitalize(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "all lower case",
			s:    "hello",
			want: "Hello",
		},
		{
			name: "all upper case",
			s:    "HELLO",
			want: "Hello",
		},
		{
			name: "diff",
			s:    "HeLlo",
			want: "Hello",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Capitalize(tt.s); got != tt.want {
				t.Errorf("Capitalize() = %v, want %v", got, tt.want)
			}
		})
	}
}
