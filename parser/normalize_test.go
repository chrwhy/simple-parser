package parser

import "testing"

func TestToSimplified(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"hello", "hello"},
		{"中国", "中国"},
		{"中華人民共和國", "中华人民共和国"},
		{"我愛中國", "我爱中国"},
		{"繁簡混合 Jay Chou", "繁简混合 Jay Chou"},
		{"觀", "观"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToSimplified(tt.input)
			if got != tt.want {
				t.Errorf("ToSimplified(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
