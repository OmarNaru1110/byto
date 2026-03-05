package updater

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

// ─── NewUpdater ────────────────────────────────────────────────────────────────

func TestNewUpdater(t *testing.T) {
	u := NewUpdater()
	if u == nil {
		t.Fatal("expected non-nil Updater")
	}
	if u.httpClient == nil {
		t.Fatal("expected non-nil httpClient")
	}
	if u.httpClient.Timeout == 0 {
		t.Error("expected non-zero client timeout")
	}
}

// ─── GetAppVersion ─────────────────────────────────────────────────────────────

func TestGetAppVersion(t *testing.T) {
	u := NewUpdater()
	v := u.GetAppVersion()
	if v != AppVersion {
		t.Errorf("GetAppVersion() = %q, want %q", v, AppVersion)
	}
	// Sanity: version looks like semver
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		t.Errorf("AppVersion %q does not look like semver", v)
	}
}

// ─── compareVersions ───────────────────────────────────────────────────────────

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want int
	}{
		{"equal", "1.0.0", "1.0.0", 0},
		{"v1 greater major", "2.0.0", "1.0.0", 1},
		{"v1 less major", "1.0.0", "2.0.0", -1},
		{"v1 greater minor", "1.2.0", "1.1.0", 1},
		{"v1 less minor", "1.1.0", "1.2.0", -1},
		{"v1 greater patch", "1.0.2", "1.0.1", 1},
		{"v1 less patch", "1.0.1", "1.0.2", -1},
		{"with v prefix both", "v1.2.3", "v1.2.3", 0},
		{"with v prefix mixed", "v2.0.0", "1.0.0", 1},
		{"different lengths short vs long", "1.0", "1.0.0", 0},
		{"different lengths v1 shorter greater", "1.1", "1.0.0", 1},
		{"different lengths v1 shorter less", "1.0", "1.0.1", -1},
		{"large numbers", "10.20.30", "10.20.29", 1},
		{"zeroes", "0.0.0", "0.0.0", 0},
		{"single digit", "2", "1", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVersions(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

// ─── CheckAppUpdate ────────────────────────────────────────────────────────────

func TestCheckAppUpdate_HasUpdate(t *testing.T) {
	versionInfo := VersionInfo{
		Version:     "99.0.0",
		ReleaseDate: "2026-01-01",
		Changelog:   "Big update",
		MinVersion:  "1.0.0",
	}
	versionInfo.Downloads.Windows = "https://example.com/windows.exe"
	versionInfo.Downloads.Darwin = "https://example.com/darwin.dmg"
	versionInfo.Downloads.Linux = "https://example.com/linux.tar.gz"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(versionInfo)
	}))
	defer server.Close()

	u := NewUpdater()
	u.httpClient = server.Client()
	u.httpClient.Transport = &urlRewriteTransport{
		defaultBase: server.URL,
	}

	result := u.CheckAppUpdate()
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Message)
	}
	if !result.HasUpdate {
		t.Error("expected HasUpdate=true")
	}
	if result.LatestVersion != "99.0.0" {
		t.Errorf("expected LatestVersion=99.0.0, got %s", result.LatestVersion)
	}
	if result.CurrentVersion != AppVersion {
		t.Errorf("expected CurrentVersion=%s, got %s", AppVersion, result.CurrentVersion)
	}
	if result.Changelog != "Big update" {
		t.Errorf("unexpected changelog: %s", result.Changelog)
	}
	if result.DownloadURL == "" {
		t.Error("expected non-empty download URL")
	}
}

func TestCheckAppUpdate_NoUpdate(t *testing.T) {
	versionInfo := VersionInfo{
		Version: AppVersion, // same as current
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(versionInfo)
	}))
	defer server.Close()

	u := NewUpdater()
	u.httpClient = server.Client()
	u.httpClient.Transport = &urlRewriteTransport{defaultBase: server.URL}

	result := u.CheckAppUpdate()
	if !result.Success {
		t.Fatalf("expected success: %s", result.Message)
	}
	if result.HasUpdate {
		t.Error("expected HasUpdate=false for same version")
	}
}

func TestCheckAppUpdate_OlderVersion(t *testing.T) {
	versionInfo := VersionInfo{
		Version: "0.0.1", // older than current
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(versionInfo)
	}))
	defer server.Close()

	u := NewUpdater()
	u.httpClient = server.Client()
	u.httpClient.Transport = &urlRewriteTransport{defaultBase: server.URL}

	result := u.CheckAppUpdate()
	if !result.Success {
		t.Fatalf("expected success: %s", result.Message)
	}
	if result.HasUpdate {
		t.Error("expected HasUpdate=false for older version")
	}
}

func TestCheckAppUpdate_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	u := NewUpdater()
	u.httpClient = server.Client()
	u.httpClient.Transport = &urlRewriteTransport{defaultBase: server.URL}

	result := u.CheckAppUpdate()
	if result.Success {
		t.Error("expected failure for server error")
	}
	if result.CurrentVersion != AppVersion {
		t.Errorf("expected CurrentVersion even on failure, got %s", result.CurrentVersion)
	}
}

func TestCheckAppUpdate_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{invalid"))
	}))
	defer server.Close()

	u := NewUpdater()
	u.httpClient = server.Client()
	u.httpClient.Transport = &urlRewriteTransport{defaultBase: server.URL}

	result := u.CheckAppUpdate()
	if result.Success {
		t.Error("expected failure for invalid JSON")
	}
	if !strings.Contains(result.Message, "Failed to parse version info") {
		t.Errorf("unexpected message: %s", result.Message)
	}
}

func TestCheckAppUpdate_ConnectionRefused(t *testing.T) {
	u := NewUpdater()
	// Use a transport that redirects to a non-existent server
	u.httpClient.Transport = &urlRewriteTransport{defaultBase: "http://127.0.0.1:1"}

	result := u.CheckAppUpdate()
	if result.Success {
		t.Error("expected failure when server is unreachable")
	}
	if result.CurrentVersion != AppVersion {
		t.Errorf("expected CurrentVersion=%s even on error", AppVersion)
	}
}

// ─── DownloadAppUpdate ─────────────────────────────────────────────────────────

func TestDownloadAppUpdate_Success(t *testing.T) {
	content := []byte("installer-binary-content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.Write(content)
	}))
	defer server.Close()

	u := NewUpdater()

	var progressCalls int64
	destPath, err := u.DownloadAppUpdate(server.URL+"/byto-update.exe", func(downloaded, total int64) {
		atomic.AddInt64(&progressCalls, 1)
		if total != int64(len(content)) {
			t.Errorf("expected total=%d, got %d", len(content), total)
		}
	})
	if err != nil {
		t.Fatalf("DownloadAppUpdate error: %v", err)
	}
	defer os.Remove(destPath)

	if destPath == "" {
		t.Fatal("expected non-empty dest path")
	}

	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Error("downloaded content mismatch")
	}

	if atomic.LoadInt64(&progressCalls) == 0 {
		t.Error("expected progress callback to be called")
	}
}

func TestDownloadAppUpdate_EmptyURL(t *testing.T) {
	u := NewUpdater()
	_, err := u.DownloadAppUpdate("", nil)
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
	if !strings.Contains(err.Error(), "no download URL") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDownloadAppUpdate_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	u := NewUpdater()
	_, err := u.DownloadAppUpdate(server.URL+"/update.exe", nil)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !strings.Contains(err.Error(), "HTTP 404") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDownloadAppUpdate_NilProgressCallback(t *testing.T) {
	content := []byte("binary")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(content)
	}))
	defer server.Close()

	u := NewUpdater()
	destPath, err := u.DownloadAppUpdate(server.URL+"/byto-update.exe", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.Remove(destPath)

	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("file was not created")
	}
}

func TestDownloadAppUpdate_FilenameFromURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("x"))
	}))
	defer server.Close()

	u := NewUpdater()
	destPath, err := u.DownloadAppUpdate(server.URL+"/my-custom-installer.exe", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.Remove(destPath)

	base := filepath.Base(destPath)
	if base != "my-custom-installer.exe" {
		t.Errorf("expected filename my-custom-installer.exe, got %s", base)
	}
}

// ─── VersionInfo / UpdateResult struct tests ───────────────────────────────────

func TestVersionInfo_JSONRoundtrip(t *testing.T) {
	vi := VersionInfo{
		Version:     "2.0.0",
		ReleaseDate: "2026-01-01",
		Changelog:   "test changes",
		MinVersion:  "1.0.0",
	}
	vi.Downloads.Windows = "https://example.com/win"
	vi.Downloads.Darwin = "https://example.com/mac"
	vi.Downloads.Linux = "https://example.com/linux"

	data, err := json.Marshal(vi)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded VersionInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.Version != vi.Version {
		t.Errorf("version mismatch: %s vs %s", decoded.Version, vi.Version)
	}
	if decoded.Downloads.Windows != vi.Downloads.Windows {
		t.Errorf("windows URL mismatch")
	}
}

func TestUpdateResult_JSONRoundtrip(t *testing.T) {
	ur := UpdateResult{
		Success:        true,
		Message:        "ok",
		CurrentVersion: "1.0.0",
		LatestVersion:  "2.0.0",
		HasUpdate:      true,
		Changelog:      "changes",
		DownloadURL:    "https://example.com",
	}

	data, err := json.Marshal(ur)
	if err != nil {
		t.Fatal(err)
	}

	var decoded UpdateResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.Success != ur.Success || decoded.HasUpdate != ur.HasUpdate {
		t.Error("roundtrip mismatch")
	}
}

// ─── Edge cases for compareVersions ────────────────────────────────────────────

func TestCompareVersions_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want int
	}{
		{"empty both", "", "", 0},
		{"empty vs version", "", "1.0.0", -1},
		{"version vs empty", "1.0.0", "", 1},
		{"v prefix only", "v", "v", 0},
		{"extra dots", "1.0.0.0", "1.0.0", 0},
		{"extra dots with value", "1.0.0.1", "1.0.0", 1},
		{"non-numeric defaults to zero", "a.b.c", "0.0.0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVersions(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

// ─── Constants ─────────────────────────────────────────────────────────────────

func TestConstants(t *testing.T) {
	if GitHubOwner == "" {
		t.Error("GitHubOwner should not be empty")
	}
	if GitHubRepo == "" {
		t.Error("GitHubRepo should not be empty")
	}
}

// ─── Helpers ───────────────────────────────────────────────────────────────────

// urlRewriteTransport rewrites specific URLs to point at test servers.
type urlRewriteTransport struct {
	base        http.RoundTripper
	rewrites    map[string]string
	defaultBase string // if set, all non-rewritten URLs go here
}

func (t *urlRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	urlStr := req.URL.String()

	// Check explicit rewrites
	for from, to := range t.rewrites {
		if urlStr == from || strings.HasPrefix(urlStr, from) {
			newReq := req.Clone(req.Context())
			newURL := to
			if urlStr != from {
				suffix := urlStr[len(from):]
				newURL = to + suffix
			}
			parsed, err := req.URL.Parse(newURL)
			if err != nil {
				return nil, err
			}
			newReq.URL = parsed
			newReq.Host = parsed.Host
			transport := t.base
			if transport == nil {
				transport = http.DefaultTransport
			}
			return transport.RoundTrip(newReq)
		}
	}

	// Default rewrite: redirect all requests to the test server
	if t.defaultBase != "" {
		newReq := req.Clone(req.Context())
		parsed, err := req.URL.Parse(t.defaultBase + req.URL.Path)
		if err != nil {
			return nil, err
		}
		newReq.URL = parsed
		newReq.Host = parsed.Host
		transport := t.base
		if transport == nil {
			transport = http.DefaultTransport
		}
		return transport.RoundTrip(newReq)
	}

	transport := t.base
	if transport == nil {
		transport = http.DefaultTransport
	}
	return transport.RoundTrip(req)
}

