package builder_test

import (
	"byto/internal/builder"
	"byto/internal/deps"
	"byto/internal/domain"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

func TestNewYTDLPBuilder_NotNil(t *testing.T) {
	b := builder.NewYTDLPBuilder()
	if b == nil {
		t.Fatal("NewYTDLPBuilder returned nil")
	}
}

func TestNewYTDLPBuilder_EmptyBuild(t *testing.T) {
	args := builder.NewYTDLPBuilder().Build()
	if len(args) != 0 {
		t.Errorf("expected empty args from fresh builder, got %v", args)
	}
}

func TestNewYTDLPBuilder_WithoutManager_YtDlpPathEmpty(t *testing.T) {
	p := builder.NewYTDLPBuilder().GetYtDlpPath()
	if p != "" {
		t.Fatalf("expected empty yt-dlp path without manager, got %q", p)
	}
}

func TestNewYTDLPBuilderWithDeps_UsesManagerPath(t *testing.T) {
	depRoot := t.TempDir()
	store, err := deps.NewStateStore(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	m := deps.NewManager(store)
	dep := deps.NewYTDLPDependency(depRoot, time.Hour)
	m.Add(dep)

	b := builder.NewYTDLPBuilderWithDeps(m)
	if got := b.GetYtDlpPath(); got != dep.Path() {
		t.Errorf("GetYtDlpPath: got %q, want %q", got, dep.Path())
	}
}

// ---------------------------------------------------------------------------
// ProgressTemplate
// ---------------------------------------------------------------------------

func TestProgressTemplate_AddsArgs(t *testing.T) {
	args := builder.NewYTDLPBuilder().ProgressTemplate("TPL").Build()
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d: %v", len(args), args)
	}
	if args[0] != "--progress-template" {
		t.Errorf("expected --progress-template, got %s", args[0])
	}
	if args[1] != "TPL" {
		t.Errorf("expected TPL, got %s", args[1])
	}
}

func TestProgressTemplate_EmptyString(t *testing.T) {
	args := builder.NewYTDLPBuilder().ProgressTemplate("").Build()
	if len(args) != 2 || args[1] != "" {
		t.Errorf("expected empty template value, got %v", args)
	}
}

func TestProgressTemplate_ComplexTemplate(t *testing.T) {
	tpl := "[byto] %(info.title)s [downloaded] %(progress.downloaded_bytes)s"
	args := builder.NewYTDLPBuilder().ProgressTemplate(tpl).Build()
	if args[1] != tpl {
		t.Errorf("template mismatch: got %s", args[1])
	}
}

// ---------------------------------------------------------------------------
// Newline
// ---------------------------------------------------------------------------

func TestNewline_AddsFlag(t *testing.T) {
	args := builder.NewYTDLPBuilder().Newline().Build()
	if len(args) != 1 || args[0] != "--newline" {
		t.Errorf("expected [--newline], got %v", args)
	}
}

// ---------------------------------------------------------------------------
// Quality
// ---------------------------------------------------------------------------

func TestQuality_AllLevels(t *testing.T) {
	tests := []struct {
		quality  domain.VideoQuality
		contains string
	}{
		{domain.Quality360p, "360"},
		{domain.Quality480p, "480"},
		{domain.Quality720p, "720"},
		{domain.Quality1080p, "1080"},
		{domain.Quality1440p, "1440"},
		{domain.Quality2160p, "2160"},
	}

	for _, tt := range tests {
		t.Run(tt.contains+"p", func(t *testing.T) {
			args := builder.NewYTDLPBuilder().Video(tt.quality).Build()
			if len(args) != 2 {
				t.Fatalf("expected 2 args, got %d: %v", len(args), args)
			}
			if args[0] != "-f" {
				t.Errorf("expected -f flag, got %s", args[0])
			}
			if !strings.Contains(args[1], tt.contains) {
				t.Errorf("expected format string to contain %s, got %s", tt.contains, args[1])
			}
		})
	}
}

func TestVideo_DefaultFallback(t *testing.T) {
	// An invalid quality value should produce a "best" fallback
	args := builder.NewYTDLPBuilder().Video(domain.VideoQuality(99)).Build()
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d: %v", len(args), args)
	}
	if !strings.Contains(args[1], "best") {
		t.Errorf("expected best fallback, got %s", args[1])
	}
}

func TestVideo_FormatContainsFallback(t *testing.T) {
	// All quality levels should contain /best as a final fallback
	args := builder.NewYTDLPBuilder().Video(domain.Quality720p).Build()
	if !strings.HasSuffix(args[1], "/best") {
		t.Errorf("expected format to end with /best fallback, got %s", args[1])
	}
}

// ---------------------------------------------------------------------------
// Audio
// ---------------------------------------------------------------------------

func TestAudio_AddsFormatFlag(t *testing.T) {
	args := builder.NewYTDLPBuilder().Audio().Build()
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d: %v", len(args), args)
	}
	if args[0] != "-f" {
		t.Errorf("expected -f flag, got %s", args[0])
	}
	if args[1] != "bestaudio/best" {
		t.Errorf("expected bestaudio/best, got %s", args[1])
	}
}

func TestAudio_ReturnsSameBuilder(t *testing.T) {
	b := builder.NewYTDLPBuilder()
	b2 := b.Audio()
	if b != b2 {
		t.Error("Audio did not return same builder")
	}
}

func TestAudio_ChainsWithOtherMethods(t *testing.T) {
	args := builder.NewYTDLPBuilder().
		Audio().
		DownloadPath("/tmp").
		URL("https://example.com/audio").
		Build()

	// -f bestaudio/best -o /tmp/%(title).100s.%(ext)s https://example.com/audio
	if len(args) != 5 {
		t.Fatalf("expected 5 args, got %d: %v", len(args), args)
	}
	if args[0] != "-f" || args[1] != "bestaudio/best" {
		t.Errorf("expected -f bestaudio/best at start, got %s %s", args[0], args[1])
	}
}

func TestAudio_ContainsBestFallback(t *testing.T) {
	args := builder.NewYTDLPBuilder().Audio().Build()
	if !strings.HasSuffix(args[1], "/best") {
		t.Errorf("expected format to end with /best fallback, got %s", args[1])
	}
}

// ---------------------------------------------------------------------------
// DownloadPath
// ---------------------------------------------------------------------------

func TestDownloadPath_AddsOutputTemplate(t *testing.T) {
	args := builder.NewYTDLPBuilder().DownloadPath("/tmp/dl").Build()
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d: %v", len(args), args)
	}
	if args[0] != "-o" {
		t.Errorf("expected -o flag, got %s", args[0])
	}
	if !strings.HasPrefix(args[1], "/tmp/dl/") {
		t.Errorf("expected path prefix /tmp/dl/, got %s", args[1])
	}
	if !strings.Contains(args[1], "%(title)") {
		t.Errorf("expected template with %%(title), got %s", args[1])
	}
	if !strings.Contains(args[1], "%(ext)s") {
		t.Errorf("expected template with %%(ext)s, got %s", args[1])
	}
}

func TestDownloadPath_EmptyPath(t *testing.T) {
	args := builder.NewYTDLPBuilder().DownloadPath("").Build()
	if args[1] != "/%(title).100s.%(ext)s" {
		t.Errorf("unexpected output template for empty path: %s", args[1])
	}
}

func TestDownloadPath_PathWithSpaces(t *testing.T) {
	args := builder.NewYTDLPBuilder().DownloadPath("/my path/dir").Build()
	if !strings.Contains(args[1], "/my path/dir/") {
		t.Errorf("path with spaces not preserved: %s", args[1])
	}
}

// ---------------------------------------------------------------------------
// SafeFilenames
// ---------------------------------------------------------------------------

func TestSafeFilenames_PlatformSpecific(t *testing.T) {
	args := builder.NewYTDLPBuilder().SafeFilenames().Build()
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d: %v", len(args), args)
	}
	if runtime.GOOS == "windows" {
		if args[0] != "--windows-filenames" {
			t.Errorf("expected --windows-filenames on Windows, got %s", args[0])
		}
	} else {
		if args[0] != "--restrict-filenames" {
			t.Errorf("expected --restrict-filenames on non-Windows, got %s", args[0])
		}
	}
}

// ---------------------------------------------------------------------------
// URL
// ---------------------------------------------------------------------------

func TestURL_AddsURL(t *testing.T) {
	url := "https://youtube.com/watch?v=abc123"
	args := builder.NewYTDLPBuilder().URL(url).Build()
	if len(args) != 1 || args[0] != url {
		t.Errorf("expected [%s], got %v", url, args)
	}
}

func TestURL_EmptyURL(t *testing.T) {
	args := builder.NewYTDLPBuilder().URL("").Build()
	if len(args) != 1 || args[0] != "" {
		t.Errorf("expected [\"\"], got %v", args)
	}
}

func TestURL_URLWithSpecialChars(t *testing.T) {
	url := "https://example.com/video?a=1&b=2&list=PLabc"
	args := builder.NewYTDLPBuilder().URL(url).Build()
	if args[0] != url {
		t.Errorf("URL not preserved: got %s", args[0])
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestUpdate_AddsFlag(t *testing.T) {
	args := builder.NewYTDLPBuilder().Update().Build()
	if len(args) != 1 || args[0] != "--update" {
		t.Errorf("expected [--update], got %v", args)
	}
}

// ---------------------------------------------------------------------------
// Chaining
// ---------------------------------------------------------------------------

func TestChaining_MultipleMethods(t *testing.T) {
	args := builder.NewYTDLPBuilder().
		ProgressTemplate("TPL").
		Newline().
		Video(domain.Quality720p).
		DownloadPath("/tmp").
		SafeFilenames().
		URL("http://example.com").
		Build()

	// --progress-template TPL --newline -f <fmt> -o <output> --windows-filenames/--restrict-filenames http://example.com
	expected := 9
	if runtime.GOOS != "windows" {
		expected = 9 // same count, different flag name
	}
	if len(args) != expected {
		t.Errorf("expected %d args from full chain, got %d: %v", expected, len(args), args)
	}
}

func TestChaining_ReturnsSameBuilder(t *testing.T) {
	b := builder.NewYTDLPBuilder()
	b2 := b.ProgressTemplate("T")
	if b != b2 {
		t.Error("ProgressTemplate did not return same builder")
	}
	b3 := b.Newline()
	if b != b3 {
		t.Error("Newline did not return same builder")
	}
	b4 := b.Video(domain.Quality720p)
	if b != b4 {
		t.Error("Video did not return same builder")
	}
	b5 := b.DownloadPath("/tmp")
	if b != b5 {
		t.Error("DownloadPath did not return same builder")
	}
	b6 := b.SafeFilenames()
	if b != b6 {
		t.Error("SafeFilenames did not return same builder")
	}
	b7 := b.URL("http://example.com")
	if b != b7 {
		t.Error("URL did not return same builder")
	}
	b8 := b.Update()
	if b != b8 {
		t.Error("Update did not return same builder")
	}
}

// ---------------------------------------------------------------------------
// Build idempotency
// ---------------------------------------------------------------------------

func TestBuild_ReturnsAccumulatedArgs(t *testing.T) {
	b := builder.NewYTDLPBuilder().URL("http://a.com").Newline()
	first := b.Build()
	second := b.Build()
	if len(first) != len(second) {
		t.Errorf("Build not idempotent: %v vs %v", first, second)
	}
	for i := range first {
		if first[i] != second[i] {
			t.Errorf("Build[%d] mismatch: %s vs %s", i, first[i], second[i])
		}
	}
}

func TestBuild_ArgsAfterBuild(t *testing.T) {
	b := builder.NewYTDLPBuilder().URL("http://a.com")
	args1 := b.Build()
	b.Newline()
	args2 := b.Build()
	if len(args2) != len(args1)+1 {
		t.Errorf("expected args to grow after adding Newline: %d vs %d", len(args1), len(args2))
	}
}

// ---------------------------------------------------------------------------
// GetYtDlpPath
// ---------------------------------------------------------------------------

func TestGetYtDlpPath_ConsistentAcrossCalls(t *testing.T) {
	b := builder.NewYTDLPBuilder()
	p1 := b.GetYtDlpPath()
	p2 := b.GetYtDlpPath()
	if p1 != p2 {
		t.Errorf("GetYtDlpPath not consistent: %s vs %s", p1, p2)
	}
}

// ---------------------------------------------------------------------------
// Multiple builders are independent
// ---------------------------------------------------------------------------

func TestMultipleBuilders_Independent(t *testing.T) {
	b1 := builder.NewYTDLPBuilder().URL("http://a.com")
	b2 := builder.NewYTDLPBuilder().URL("http://b.com").Newline()

	args1 := b1.Build()
	args2 := b2.Build()

	if len(args1) == len(args2) {
		t.Error("independent builders should have different arg counts")
	}
	if args1[0] == args2[0] {
		// URLs are at different positions or same position but different values
		// Actually URL is always last, so args1[0] = "http://a.com", args2[0] = "http://b.com"
		t.Error("independent builders should have different URLs")
	}
}

// ---------------------------------------------------------------------------
// Playlist
// ---------------------------------------------------------------------------

// --- SelectionAll -----------------------------------------------------------

func TestPlaylist_SelectionAll_NoArgsAdded(t *testing.T) {
	ps := domain.PlaylistSelection{Type: domain.SelectionAll}
	args := builder.NewYTDLPBuilder().Playlist(ps).Build()
	// SelectionAll passes Validate but hits no switch case → no extra args
	if len(args) != 0 {
		t.Errorf("expected no args for SelectionAll, got %v", args)
	}
}

// --- Unknown / empty type ---------------------------------------------------

func TestPlaylist_EmptyType_NoArgsAdded(t *testing.T) {
	ps := domain.PlaylistSelection{} // zero value
	args := builder.NewYTDLPBuilder().Playlist(ps).Build()
	if len(args) != 0 {
		t.Errorf("expected no args for empty type, got %v", args)
	}
}

func TestPlaylist_UnknownType_NoArgsAdded(t *testing.T) {
	ps := domain.PlaylistSelection{Type: domain.PlaylistSelectionType("custom")}
	args := builder.NewYTDLPBuilder().Playlist(ps).Build()
	if len(args) != 0 {
		t.Errorf("expected no args for unknown type, got %v", args)
	}
}

// --- SelectionRange (valid) -------------------------------------------------

func TestPlaylist_Range_Valid_AddsPlaylistItemsFlag(t *testing.T) {
	ps := domain.PlaylistSelection{
		Type:       domain.SelectionRange,
		StartIndex: 2,
		EndIndex:   7,
	}
	args := builder.NewYTDLPBuilder().Playlist(ps).Build()
	if len(args) != 2 {
		t.Fatalf("expected 2 args for valid range, got %d: %v", len(args), args)
	}
	if args[0] != "--playlist-items" {
		t.Errorf("expected --playlist-items flag, got %q", args[0])
	}
}

func TestPlaylist_Range_Valid_FormatsCorrectly(t *testing.T) {
	ps := domain.PlaylistSelection{
		Type:       domain.SelectionRange,
		StartIndex: 2,
		EndIndex:   7,
	}
	args := builder.NewYTDLPBuilder().Playlist(ps).Build()
	if args[1] != "2-7" {
		t.Errorf("expected range value \"2-7\", got %q", args[1])
	}
}

func TestPlaylist_Range_EqualStartEnd_FormatsCorrectly(t *testing.T) {
	// Single-item range: StartIndex == EndIndex, still valid (>= 1)
	ps := domain.PlaylistSelection{
		Type:       domain.SelectionRange,
		StartIndex: 4,
		EndIndex:   4,
	}
	args := builder.NewYTDLPBuilder().Playlist(ps).Build()
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d: %v", len(args), args)
	}
	if args[1] != "4-4" {
		t.Errorf("expected \"4-4\", got %q", args[1])
	}
}

func TestPlaylist_Range_LargeRange_FormatsCorrectly(t *testing.T) {
	ps := domain.PlaylistSelection{
		Type:       domain.SelectionRange,
		StartIndex: 1,
		EndIndex:   100,
	}
	args := builder.NewYTDLPBuilder().Playlist(ps).Build()
	if args[1] != "1-100" {
		t.Errorf("expected \"1-100\", got %q", args[1])
	}
}

// --- SelectionRange (invalid — Validate guard) ------------------------------

func TestPlaylist_Range_StartIndexZero_NoArgsAdded(t *testing.T) {
	// Validate fails → Playlist must be a no-op
	ps := domain.PlaylistSelection{
		Type:       domain.SelectionRange,
		StartIndex: 0,
		EndIndex:   5,
	}
	args := builder.NewYTDLPBuilder().Playlist(ps).Build()
	if len(args) != 0 {
		t.Errorf("expected no args for invalid range (start=0), got %v", args)
	}
}

func TestPlaylist_Range_NegativeStart_NoArgsAdded(t *testing.T) {
	ps := domain.PlaylistSelection{
		Type:       domain.SelectionRange,
		StartIndex: -3,
		EndIndex:   10,
	}
	args := builder.NewYTDLPBuilder().Playlist(ps).Build()
	if len(args) != 0 {
		t.Errorf("expected no args for invalid range (start=-3), got %v", args)
	}
}

func TestPlaylist_Range_EndLessThanStart_NoArgsAdded(t *testing.T) {
	ps := domain.PlaylistSelection{
		Type:       domain.SelectionRange,
		StartIndex: 5,
		EndIndex:   2,
	}
	args := builder.NewYTDLPBuilder().Playlist(ps).Build()
	if len(args) != 0 {
		t.Errorf("expected no args for invalid range (end<start), got %v", args)
	}
}

// --- SelectionItems (valid) -------------------------------------------------

func TestPlaylist_Items_NonEmpty_AddsPlaylistItemsFlag(t *testing.T) {
	ps := domain.PlaylistSelection{
		Type:  domain.SelectionItems,
		Items: "1,3,5",
	}
	args := builder.NewYTDLPBuilder().Playlist(ps).Build()
	if len(args) != 2 {
		t.Fatalf("expected 2 args for items selection, got %d: %v", len(args), args)
	}
	if args[0] != "--playlist-items" {
		t.Errorf("expected --playlist-items flag, got %q", args[0])
	}
	if args[1] != "1,3,5" {
		t.Errorf("expected items value \"1,3,5\", got %q", args[1])
	}
}

func TestPlaylist_Items_SingleItem_FormatsCorrectly(t *testing.T) {
	ps := domain.PlaylistSelection{
		Type:  domain.SelectionItems,
		Items: "7",
	}
	args := builder.NewYTDLPBuilder().Playlist(ps).Build()
	if args[1] != "7" {
		t.Errorf("expected \"7\", got %q", args[1])
	}
}

// --- SelectionItems (invalid — Validate guard) ------------------------------

func TestPlaylist_Items_Empty_NoArgsAdded(t *testing.T) {
	ps := domain.PlaylistSelection{
		Type:  domain.SelectionItems,
		Items: "",
	}
	args := builder.NewYTDLPBuilder().Playlist(ps).Build()
	if len(args) != 0 {
		t.Errorf("expected no args for empty items, got %v", args)
	}
}

// --- Return-same-builder (fluent API identity) ------------------------------

func TestPlaylist_ReturnsSameBuilder(t *testing.T) {
	b := builder.NewYTDLPBuilder()
	ps := domain.PlaylistSelection{Type: domain.SelectionAll}
	if b.Playlist(ps) != b {
		t.Error("Playlist did not return the same builder")
	}
}

func TestPlaylist_InvalidInput_ReturnsSameBuilder(t *testing.T) {
	// Even when Validate fails the builder pointer must be returned
	b := builder.NewYTDLPBuilder()
	ps := domain.PlaylistSelection{Type: domain.SelectionItems, Items: ""}
	if b.Playlist(ps) != b {
		t.Error("Playlist did not return the same builder on invalid input")
	}
}

// --- Chaining ---------------------------------------------------------------

func TestPlaylist_ChainsWithOtherMethods(t *testing.T) {
	ps := domain.PlaylistSelection{
		Type:       domain.SelectionRange,
		StartIndex: 1,
		EndIndex:   3,
	}
	args := builder.NewYTDLPBuilder().
		Video(domain.Quality720p).
		Playlist(ps).
		URL("https://youtube.com/playlist?list=PLxyz").
		Build()

	// -f <fmt>  --playlist-items 1-3  https://...
	if len(args) != 5 {
		t.Fatalf("expected 5 args in chain, got %d: %v", len(args), args)
	}
	// --playlist-items must appear at positions 2-3
	if args[2] != "--playlist-items" {
		t.Errorf("expected --playlist-items at index 2, got %q", args[2])
	}
	if args[3] != "1-3" {
		t.Errorf("expected \"1-3\" at index 3, got %q", args[3])
	}
	if args[4] != "https://youtube.com/playlist?list=PLxyz" {
		t.Errorf("expected URL at index 4, got %q", args[4])
	}
}
