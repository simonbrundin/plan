package repository

import (
	"reflect"
	"testing"
)

func TestMergeOrder(t *testing.T) {
	tests := []struct {
		name      string
		existing  []int64
		requested []int64
		want      []int64
	}{
		{
			name:      "requested order is kept",
			existing:  []int64{1, 2, 3},
			requested: []int64{3, 1, 2},
			want:      []int64{3, 1, 2},
		},
		{
			name:      "children missing from the request are appended",
			existing:  []int64{1, 2, 3},
			requested: []int64{2, 1},
			want:      []int64{2, 1, 3},
		},
		{
			name:      "unknown children are ignored",
			existing:  []int64{1, 2},
			requested: []int64{2, 99, 1},
			want:      []int64{2, 1},
		},
		{
			name:      "duplicates in the request are removed",
			existing:  []int64{1, 2},
			requested: []int64{2, 2, 1},
			want:      []int64{2, 1},
		},
		{
			name:      "empty request keeps the existing order",
			existing:  []int64{1, 2},
			requested: nil,
			want:      []int64{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeOrder(tt.existing, tt.requested)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeOrder(%v, %v) = %v, want %v", tt.existing, tt.requested, got, tt.want)
			}
		})
	}
}
