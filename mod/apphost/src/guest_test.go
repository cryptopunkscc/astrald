package apphost

import (
	"net/url"
	"strings"
	"testing"
)

func TestPrepareQueryString_BinaryMode(t *testing.T) {
	g := &Guest{mode: ModeBinary}
	cases := []string{"", "foo", "foo?bar=1", "foo?out=text&in=json"}
	for _, in := range cases {
		if got := g.prepareQueryString(in); got != in {
			t.Errorf("binary mode must pass through unchanged: %q -> %q", in, got)
		}
	}
}

func TestPrepareQueryString_JSONMode(t *testing.T) {
	g := &Guest{mode: ModeJSON}

	cases := []struct {
		name      string
		in        string
		wantPath  string
		wantOut   string
		wantIn    string
		wantOther map[string]string
	}{
		{
			name:     "no params",
			in:       "user.info",
			wantPath: "user.info",
			wantOut:  "json",
			wantIn:   "json",
		},
		{
			name:     "empty string",
			in:       "",
			wantPath: "",
			wantOut:  "json",
			wantIn:   "json",
		},
		{
			name:     "preserves other params",
			in:       "user.info?name=alice",
			wantPath: "user.info",
			wantOut:  "json",
			wantIn:   "json",
			wantOther: map[string]string{
				"name": "alice",
			},
		},
		{
			name:     "out already set is preserved",
			in:       "user.info?out=text",
			wantPath: "user.info",
			wantOut:  "text",
			wantIn:   "json",
		},
		{
			name:     "in already set is preserved",
			in:       "user.info?in=text",
			wantPath: "user.info",
			wantOut:  "json",
			wantIn:   "text",
		},
		{
			name:     "both already set",
			in:       "user.info?out=text&in=binary",
			wantPath: "user.info",
			wantOut:  "text",
			wantIn:   "binary",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := g.prepareQueryString(c.in)

			path, rawParams, _ := strings.Cut(got, "?")
			if path != c.wantPath {
				t.Errorf("path = %q, want %q", path, c.wantPath)
			}

			params, err := url.ParseQuery(rawParams)
			if err != nil {
				t.Fatalf("parse rebuilt params: %v", err)
			}

			if params.Get("out") != c.wantOut {
				t.Errorf("out = %q, want %q (full: %q)", params.Get("out"), c.wantOut, got)
			}
			if params.Get("in") != c.wantIn {
				t.Errorf("in = %q, want %q (full: %q)", params.Get("in"), c.wantIn, got)
			}
			for k, v := range c.wantOther {
				if params.Get(k) != v {
					t.Errorf("%s = %q, want %q (full: %q)", k, params.Get(k), v, got)
				}
			}
		})
	}
}
