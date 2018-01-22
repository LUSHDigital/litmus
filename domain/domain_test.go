package domain

import "testing"

func TestHttpStatusFmt(t *testing.T) {
	tests := []struct {
		name string
		code int
		want string
	}{
		{
			name: "valid code 200",
			code: 200,
			want: "200 (OK)",
		},
		{
			name: "valid code 500",
			code: 500,
			want: "500 (Internal Server Error)",
		},
		{
			name: "invalid code 999",
			code: 999,
			want: "999 (INVALID RESPONSE CODE)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HttpStatusFmt(tt.code); got != tt.want {
				t.Errorf("HttpStatusFmt() = %v, want %v", got, tt.want)
			}
		})
	}
}
