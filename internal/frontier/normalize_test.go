package frontier

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "lowercase scheme and host", in: "HTTPS://Example.COM/Path", want: "https://example.com/Path"},
		{name: "strip fragment", in: "https://example.com/docs#intro", want: "https://example.com/docs"},
		{name: "drop default https port", in: "https://example.com:443/x", want: "https://example.com/x"},
		{name: "drop default http port", in: "http://example.com:80/x", want: "http://example.com/x"},
		{name: "keep custom port", in: "http://localhost:8080/app", want: "http://localhost:8080/app"},
		{name: "empty path becomes root", in: "https://example.com", want: "https://example.com/"},
		{name: "clean path segments", in: "https://example.com/a//b/../c", want: "https://example.com/a/c"},
		{name: "strip trailing slash", in: "https://example.com/about/", want: "https://example.com/about"},
		{name: "keep root slash", in: "https://example.com/", want: "https://example.com/"},
		{name: "drop tracking params", in: "https://example.com/?utm_source=tw&id=7", want: "https://example.com/?id=7"},
		{name: "drop click ids", in: "https://example.com/?gclid=abc&page=2", want: "https://example.com/?page=2"},
		{name: "sort query params", in: "https://example.com/?b=2&a=1", want: "https://example.com/?a=1&b=2"},
		{name: "reject ftp", in: "ftp://example.com/file", wantErr: true},
		{name: "reject missing host", in: "https:///path", wantErr: true},
		{name: "reject garbage", in: "not a url at all", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Normalize(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Normalize(%q) = %q, want error", tt.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Normalize(%q) returned error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestResolveRef(t *testing.T) {
	const base = "https://example.com/docs/intro"

	tests := []struct {
		name    string
		ref     string
		want    string
		wantErr bool
	}{
		{name: "relative link", ref: "page.html", want: "https://example.com/docs/page.html"},
		{name: "parent directory", ref: "../about", want: "https://example.com/about"},
		{name: "absolute path", ref: "/pricing", want: "https://example.com/pricing"},
		{name: "cross host", ref: "https://other.com/x", want: "https://other.com/x"},
		{name: "scheme relative", ref: "//cdn.example.com/a", want: "https://cdn.example.com/a"},
		{name: "fragment only", ref: "#section", want: "https://example.com/docs/intro"},
		{name: "query only", ref: "?q=1", want: "https://example.com/docs/intro?q=1"},
		{name: "skip mailto", ref: "mailto:hi@example.com", wantErr: true},
		{name: "skip javascript", ref: "javascript:void(0)", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveRef(base, tt.ref)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ResolveRef(%q) = %q, want error", tt.ref, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveRef(%q) returned error: %v", tt.ref, err)
			}
			if got != tt.want {
				t.Errorf("ResolveRef(%q) = %q, want %q", tt.ref, got, tt.want)
			}
		})
	}
}
