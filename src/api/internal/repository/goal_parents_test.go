package repository

import "testing"

func TestParseGoalParents(t *testing.T) {
	cases := []struct {
		name string
		raw  []byte
		want []string
	}{
		{
			name: "tom lista",
			raw:  []byte(`[]`),
			want: []string{},
		},
		{
			name: "saknad data",
			raw:  nil,
			want: []string{},
		},
		{
			name: "ogiltig json",
			raw:  []byte(`{"id": 1`),
			want: []string{},
		},
		{
			name: "en förälder",
			raw:  []byte(`[{"id": 12, "title": "Hälsa"}]`),
			want: []string{"Hälsa"},
		},
		{
			name: "flera föräldrar",
			raw:  []byte(`[{"id": 12, "title": "Hälsa"}, {"id": 13, "title": "Ekonomi"}]`),
			want: []string{"Hälsa", "Ekonomi"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseGoalParents(tc.raw)
			if len(got) != len(tc.want) {
				t.Fatalf("parseGoalParents gav %d föräldrar, vill ha %d", len(got), len(tc.want))
			}
			for i, title := range tc.want {
				if got[i].Title != title {
					t.Fatalf("förälder %d = %q, vill ha %q", i, got[i].Title, title)
				}
			}
		})
	}
}
