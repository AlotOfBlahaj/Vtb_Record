package utils

import "testing"

func TestRemoveIllegalChar(t *testing.T) {
	type args struct {
		Title string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"1", args{Title: "👿1"}, "1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveIllegalChar(tt.args.Title); got != tt.want {
				t.Errorf("RemoveIllegalChar() = %v, want %v", got, tt.want)
			}
		})
	}
}
