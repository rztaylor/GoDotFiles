package updater

import (
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestGetLatestVersion(t *testing.T) {
	t.Run("returns latest release when matching asset exists", func(t *testing.T) {
		assetArch := runtime.GOARCH
		if assetArch == "amd64" {
			assetArch = "x86_64"
		}

		oldBaseURL := githubAPIBaseURL
		oldClient := githubHTTPClient
		githubAPIBaseURL = "https://api.github.com"
		githubHTTPClient = &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/repos/rztaylor/GoDotFiles/releases/latest" {
					return &http.Response{
						StatusCode: http.StatusNotFound,
						Body:       io.NopCloser(strings.NewReader("not found")),
						Header:     make(http.Header),
					}, nil
				}
				body := fmt.Sprintf(`{
					"tag_name":"v1.2.3",
					"name":"v1.2.3",
					"body":"notes",
					"published_at":"2026-02-15T00:00:00Z",
					"html_url":"https://example.com/release",
					"assets":[{"name":"gdf_%s_%s.tar.gz","browser_download_url":"https://example.com/gdf.tar.gz"}]
				}`, runtime.GOOS, assetArch)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(body)),
					Header:     make(http.Header),
				}, nil
			}),
		}
		t.Cleanup(func() {
			githubAPIBaseURL = oldBaseURL
			githubHTTPClient = oldClient
		})

		got, err := GetLatestVersion()
		if err != nil {
			t.Fatalf("GetLatestVersion() unexpected error: %v", err)
		}
		if got == nil {
			t.Fatal("GetLatestVersion() returned nil, want release")
		}
		if got.Version.String() != "1.2.3" {
			t.Fatalf("GetLatestVersion() version = %q, want %q", got.Version.String(), "1.2.3")
		}
		if got.AssetURL == "" {
			t.Fatal("GetLatestVersion() asset URL is empty")
		}
	})

	t.Run("returns nil for non-200 response", func(t *testing.T) {
		oldBaseURL := githubAPIBaseURL
		oldClient := githubHTTPClient
		githubAPIBaseURL = "https://api.github.com"
		githubHTTPClient = &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusTooManyRequests,
					Body:       io.NopCloser(strings.NewReader("rate limited")),
					Header:     make(http.Header),
				}, nil
			}),
		}
		t.Cleanup(func() {
			githubAPIBaseURL = oldBaseURL
			githubHTTPClient = oldClient
		})

		got, err := GetLatestVersion()
		if err != nil {
			t.Fatalf("GetLatestVersion() unexpected error: %v", err)
		}
		if got != nil {
			t.Fatalf("GetLatestVersion() = %#v, want nil", got)
		}
	})

	t.Run("returns nil when no compatible asset exists", func(t *testing.T) {
		oldBaseURL := githubAPIBaseURL
		oldClient := githubHTTPClient
		githubAPIBaseURL = "https://api.github.com"
		githubHTTPClient = &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{
						"tag_name":"v1.2.3",
						"name":"v1.2.3",
						"body":"notes",
						"published_at":"2026-02-15T00:00:00Z",
						"html_url":"https://example.com/release",
						"assets":[{"name":"gdf_windows_arm64.zip","browser_download_url":"https://example.com/gdf.zip"}]
					}`)),
					Header: make(http.Header),
				}, nil
			}),
		}
		t.Cleanup(func() {
			githubAPIBaseURL = oldBaseURL
			githubHTTPClient = oldClient
		})

		got, err := GetLatestVersion()
		if err != nil {
			t.Fatalf("GetLatestVersion() unexpected error: %v", err)
		}
		if got != nil {
			t.Fatalf("GetLatestVersion() = %#v, want nil", got)
		}
	})
}
