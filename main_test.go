package main

import (
	"testing"
)

func TestNewLink(t *testing.T) {
	type Input struct {
		href       string
		currentURL string
	}

	testCases := []struct {
		desc  string
		input Input
		want  string
	}{
		{desc: "", input: Input{href: "https://example1.com", currentURL: ""}, want: "https://example1.com"},
		{desc: "", input: Input{href: "https://example2.com", currentURL: "http://asdf2.com"}, want: "https://example2.com"},
		{desc: "", input: Input{href: "https://example3.com", currentURL: "asdf2.com"}, want: ""},
		{desc: "", input: Input{href: "/path", currentURL: "http://asdf.com"}, want: "http://asdf.com/path"},
		{desc: "", input: Input{href: "/path", currentURL: "http://asdf.com"}, want: "http://asdf.com/path"},
		{desc: "", input: Input{href: "/path", currentURL: "http://asdf.com"}, want: "http://asdf.com/path"},
		{desc: "", input: Input{href: "//en.m.wikipedia.org/w/index.php", currentURL: "https://wikipedia.com"}, want: "https://en.m.wikipedia.org/w/index.php"},
		{desc: "", input: Input{href: "//en.m.wikipedia.org/w/index.php", currentURL: "wikipedia.com"}, want: ""},
		{desc: "", input: Input{href: "https://404notboring.com/articles/true-beauty-mhk", currentURL: "http://404notboring.com/"}, want: "https://404notboring.com/articles/true-beauty-mhk"},
		{desc: "", input: Input{href: "/articles", currentURL: "https://404notboring.com/articles/boids"}, want: "https://404notboring.com/articles"},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ans, _ := ResolveHrefToURL(tC.input.href, tC.input.currentURL)

			if ans != tC.want {
				t.Errorf("FAIL %s: got %v, want %v", tC.desc, ans, tC.want)
			}
		})
	}
}
