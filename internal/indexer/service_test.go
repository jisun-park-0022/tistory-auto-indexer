package indexer

import "testing"

func TestDetectNewURLs(t *testing.T) {
	tests := []struct {
		name    string
		current []string
		known   []string
		want    []string
	}{
		{
			name:    "all new",
			current: []string{"a", "b"},
			known:   []string{},
			want:    []string{"a", "b"},
		},
		{
			name:    "no new",
			current: []string{"a", "b"},
			known:   []string{"a", "b"},
			want:    nil,
		},
		{
			name:    "partial new",
			current: []string{"a", "b", "c"},
			known:   []string{"a", "b"},
			want:    []string{"c"},
		},
		{
			name:    "empty current",
			current: []string{},
			known:   []string{"a"},
			want:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := detectNewURLs(tc.current, tc.known)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for i, u := range got {
				if u != tc.want[i] {
					t.Errorf("[%d] got %q, want %q", i, u, tc.want[i])
				}
			}
		})
	}
}
