package handlers

import "testing"

// TestValidateColor kontrollerar vilka färgkoder som accepteras och hur de
// normaliseras.
func TestValidateColor(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    string
		wantOk bool
	}{
		{name: "sex siffror", input: "#6366f1", want: "#6366f1", wantOk: true},
		{name: "tre siffror expanderas", input: "#abc", want: "#aabbcc", wantOk: true},
		{name: "versaler sänks", input: "#AABBCC", want: "#aabbcc", wantOk: true},
		{name: "mellanslag trimmas", input: "  #6366f1  ", want: "#6366f1", wantOk: true},
		{name: "tomt betyder ofärgat", input: "", want: "", wantOk: true},
		{name: "saknad hash avvisas", input: "6366f1", wantOk: false},
		{name: "fel längd avvisas", input: "#12345", wantOk: false},
		{name: "ogiltiga tecken avvisas", input: "#6366fG", wantOk: false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			input := testCase.input
			color, ok := validateColor(&input)
			if ok != testCase.wantOk {
				t.Fatalf("ok = %v, want %v", ok, testCase.wantOk)
			}
			if !ok {
				return
			}
			if color == nil {
				t.Fatal("color = nil, want a value")
			}
			if *color != testCase.want {
				t.Fatalf("color = %q, want %q", *color, testCase.want)
			}
		})
	}
}
