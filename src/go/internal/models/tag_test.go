package models

import (
	"errors"
	"testing"
)

func TestNormalizeTagName(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "vanlig tagg", raw: "ute", want: "ute"},
		{name: "gemener", raw: "Ute", want: "ute"},
		{name: "blanksteg runt", raw: "  ute  ", want: "ute"},
		{name: "hash-prefix", raw: "#ute", want: "ute"},
		{name: "flera blanksteg", raw: "ute   hemma", want: "ute hemma"},
		{name: "tomt", raw: "   ", wantErr: true},
		{name: "endast hash", raw: "#", wantErr: true},
		{name: "för långt", raw: "abcdefghijklmnopqrstuvwxyz0123456789", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeTagName(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("NormalizeTagName(%q) = %q, vill ha fel", tc.raw, got)
				}
				if !errors.Is(err, ErrInvalidTagName) {
					t.Fatalf("err = %v, vill ha ErrInvalidTagName", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeTagName(%q) gav fel: %v", tc.raw, err)
			}
			if got != tc.want {
				t.Errorf("NormalizeTagName(%q) = %q, vill ha %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestNormalizeTagNames(t *testing.T) {
	got, err := NormalizeTagNames([]string{"Ute", "ute", " hem ", "ute"})
	if err != nil {
		t.Fatalf("NormalizeTagNames gav fel: %v", err)
	}
	want := []string{"ute", "hem"}
	if len(got) != len(want) {
		t.Fatalf("NormalizeTagNames = %v, vill ha %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("NormalizeTagNames[%d] = %q, vill ha %q", i, got[i], want[i])
		}
	}
}

func TestNormalizeTagNamesEmpty(t *testing.T) {
	got, err := NormalizeTagNames(nil)
	if err != nil {
		t.Fatalf("NormalizeTagNames gav fel: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("NormalizeTagNames = %v, vill ha tom lista", got)
	}
}

func TestNormalizeTagNamesRejectsEmptyName(t *testing.T) {
	if _, err := NormalizeTagNames([]string{"ute", "  "}); !errors.Is(err, ErrInvalidTagName) {
		t.Errorf("err = %v, vill ha ErrInvalidTagName", err)
	}
}
