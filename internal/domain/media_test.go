package domain_test

import (
	"byto/internal/domain"
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAppendLog_SingleEntry(t *testing.T) {
	m := &domain.Media{ID: "1"}
	m.AppendLog("line one")
	if len(m.Progress.Logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(m.Progress.Logs))
	}
	if m.Progress.Logs[0] != "line one" {
		t.Errorf("expected 'line one', got %q", m.Progress.Logs[0])
	}
}

func TestAppendLog_MultipleEntries(t *testing.T) {
	m := &domain.Media{ID: "1"}
	m.AppendLog("a")
	m.AppendLog("b")
	m.AppendLog("c")
	if len(m.Progress.Logs) != 3 {
		t.Fatalf("expected 3 logs, got %d", len(m.Progress.Logs))
	}
}

func TestAppendLog_EmptyString(t *testing.T) {
	m := &domain.Media{ID: "1"}
	m.AppendLog("")
	if len(m.Progress.Logs) != 1 {
		t.Fatalf("expected 1 log (empty), got %d", len(m.Progress.Logs))
	}
	if m.Progress.Logs[0] != "" {
		t.Errorf("expected empty string, got %q", m.Progress.Logs[0])
	}
}

func TestAppendLog_CallsOnProgress(t *testing.T) {
	var called int32
	m := &domain.Media{
		ID: "1",
		OnProgress: func(id string, p domain.DownloadProgress) {
			atomic.AddInt32(&called, 1)
		},
	}
	m.AppendLog("test")
	time.Sleep(50 * time.Millisecond) // callback runs in goroutine
	if atomic.LoadInt32(&called) == 0 {
		t.Error("OnProgress callback was not called")
	}
}

func TestAppendLog_NoCallbackNoPanic(t *testing.T) {
	m := &domain.Media{ID: "1"}
	// Should not panic when OnProgress is nil
	m.AppendLog("no panic")
}

func TestAppendLog_CallbackReceivesCorrectID(t *testing.T) {
	var receivedID string
	var mu sync.Mutex
	m := &domain.Media{
		ID: "my-media-42",
		OnProgress: func(id string, p domain.DownloadProgress) {
			mu.Lock()
			receivedID = id
			mu.Unlock()
		},
	}
	m.AppendLog("check id")
	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if receivedID != "my-media-42" {
		t.Errorf("expected id my-media-42, got %q", receivedID)
	}
}

// ===========================================================================
// SetTitle
// ===========================================================================

func TestSetTitle_Updates(t *testing.T) {
	m := &domain.Media{ID: "1", Title: "Old"}
	m.SetTitle("New Title")
	if m.Title != "New Title" {
		t.Errorf("expected 'New Title', got %q", m.Title)
	}
}

func TestSetTitle_EmptyString(t *testing.T) {
	m := &domain.Media{ID: "1", Title: "Has Title"}
	m.SetTitle("")
	if m.Title != "" {
		t.Errorf("expected empty title, got %q", m.Title)
	}
}

func TestSetTitle_CallsOnTitleChange(t *testing.T) {
	var called int32
	var receivedTitle string
	var mu sync.Mutex
	m := &domain.Media{
		ID: "1",
		OnTitleChange: func(id string, title string) {
			atomic.AddInt32(&called, 1)
			mu.Lock()
			receivedTitle = title
			mu.Unlock()
		},
	}
	m.SetTitle("Updated")
	time.Sleep(50 * time.Millisecond)
	if atomic.LoadInt32(&called) == 0 {
		t.Error("OnTitleChange callback was not called")
	}
	mu.Lock()
	defer mu.Unlock()
	if receivedTitle != "Updated" {
		t.Errorf("expected title 'Updated', got %q", receivedTitle)
	}
}

func TestSetTitle_NoCallbackNoPanic(t *testing.T) {
	m := &domain.Media{ID: "1"}
	m.SetTitle("safe")
}

func TestUpdateProgress_SetsValues(t *testing.T) {
	m := &domain.Media{ID: "1"}
	m.UpdateProgress(500, 1000, 50)
	if m.Progress.DownloadedBytes != 500 {
		t.Errorf("expected downloaded 500, got %d", m.Progress.DownloadedBytes)
	}
	if m.TotalBytes != 1000 {
		t.Errorf("expected total 1000, got %d", m.TotalBytes)
	}
	if m.Progress.Percentage != 50 {
		t.Errorf("expected percentage 50, got %d", m.Progress.Percentage)
	}
}

func TestUpdateProgress_ZeroValues(t *testing.T) {
	m := &domain.Media{ID: "1"}
	m.UpdateProgress(0, 0, 0)
	if m.Progress.DownloadedBytes != 0 || m.TotalBytes != 0 || m.Progress.Percentage != 0 {
		t.Error("expected all zeros")
	}
}

func TestUpdateProgress_FullProgress(t *testing.T) {
	m := &domain.Media{ID: "1"}
	m.UpdateProgress(1000, 1000, 100)
	if m.Progress.Percentage != 100 {
		t.Errorf("expected 100%%, got %d", m.Progress.Percentage)
	}
}

func TestUpdateProgress_CallsOnProgress(t *testing.T) {
	var called int32
	m := &domain.Media{
		ID: "1",
		OnProgress: func(id string, p domain.DownloadProgress) {
			atomic.AddInt32(&called, 1)
		},
	}
	m.UpdateProgress(100, 200, 50)
	time.Sleep(50 * time.Millisecond)
	if atomic.LoadInt32(&called) == 0 {
		t.Error("OnProgress callback was not called")
	}
}

func TestUpdateProgress_NoCallbackNoPanic(t *testing.T) {
	m := &domain.Media{ID: "1"}
	m.UpdateProgress(100, 200, 50)
}

func TestUpdateProgress_LargeValues(t *testing.T) {
	m := &domain.Media{ID: "1"}
	m.UpdateProgress(999_999_999_999, 1_000_000_000_000, 99)
	if m.Progress.DownloadedBytes != 999_999_999_999 {
		t.Errorf("large downloaded value not stored correctly: %d", m.Progress.DownloadedBytes)
	}
	if m.TotalBytes != 1_000_000_000_000 {
		t.Errorf("large total value not stored correctly: %d", m.TotalBytes)
	}
}

func TestUpdateProgress_OverwritesPrevious(t *testing.T) {
	m := &domain.Media{ID: "1"}
	m.UpdateProgress(100, 200, 50)
	m.UpdateProgress(200, 200, 100)
	if m.Progress.DownloadedBytes != 200 {
		t.Errorf("expected 200 after overwrite, got %d", m.Progress.DownloadedBytes)
	}
	if m.Progress.Percentage != 100 {
		t.Errorf("expected 100%% after overwrite, got %d", m.Progress.Percentage)
	}
}

func TestSetStatus_AllStatuses(t *testing.T) {
	statuses := []domain.DownloadStatus{
		domain.Pending,
		domain.InProgress,
		domain.Completed,
		domain.Failed,
		domain.Paused,
	}
	for _, s := range statuses {
		m := &domain.Media{ID: "1"}
		m.SetStatus(s)
		if m.Status != s {
			t.Errorf("expected status %d, got %d", s, m.Status)
		}
	}
}

func TestSetStatus_CallsOnStatusChange(t *testing.T) {
	var receivedStatus domain.DownloadStatus
	var mu sync.Mutex
	var called int32
	m := &domain.Media{
		ID: "1",
		OnStatusChange: func(id string, s domain.DownloadStatus) {
			atomic.AddInt32(&called, 1)
			mu.Lock()
			receivedStatus = s
			mu.Unlock()
		},
	}
	m.SetStatus(domain.Completed)
	time.Sleep(50 * time.Millisecond)
	if atomic.LoadInt32(&called) == 0 {
		t.Error("OnStatusChange not called")
	}
	mu.Lock()
	defer mu.Unlock()
	if receivedStatus != domain.Completed {
		t.Errorf("expected Completed, got %d", receivedStatus)
	}
}

func TestSetStatus_NoCallbackNoPanic(t *testing.T) {
	m := &domain.Media{ID: "1"}
	m.SetStatus(domain.Failed)
}

func TestSetStatus_TransitionSequence(t *testing.T) {
	m := &domain.Media{ID: "1", Status: domain.Pending}
	m.SetStatus(domain.InProgress)
	if m.Status != domain.InProgress {
		t.Errorf("expected InProgress, got %d", m.Status)
	}
	m.SetStatus(domain.Completed)
	if m.Status != domain.Completed {
		t.Errorf("expected Completed, got %d", m.Status)
	}
}

func TestCancel_WithCancelFunc(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m := &domain.Media{
		ID:         "1",
		Ctx:        ctx,
		CancelFunc: cancel,
	}

	m.Cancel()
	select {
	case <-ctx.Done():
		// expected
	default:
		t.Error("context was not cancelled")
	}
}

func TestCancel_NilCancelFunc_NoPanic(t *testing.T) {
	m := &domain.Media{ID: "1", CancelFunc: nil}
	// Should not panic
	m.Cancel()
}

func TestCancel_DoubleCancel_NoPanic(t *testing.T) {
	_, cancel := context.WithCancel(context.Background())
	m := &domain.Media{ID: "1", CancelFunc: cancel}
	m.Cancel()
	m.Cancel() // second call should not panic
}

func TestMedia_ConcurrentAppendLog(t *testing.T) {
	m := &domain.Media{ID: "1"}
	var wg sync.WaitGroup
	n := 100
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			m.AppendLog("log entry")
		}(i)
	}
	wg.Wait()
	if len(m.Progress.Logs) != n {
		t.Errorf("expected %d logs after concurrent writes, got %d", n, len(m.Progress.Logs))
	}
}

func TestMedia_ConcurrentSetTitle(t *testing.T) {
	m := &domain.Media{ID: "1"}
	var wg sync.WaitGroup
	n := 100
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			m.SetTitle("title")
		}(i)
	}
	wg.Wait()
	if m.Title != "title" {
		t.Errorf("expected 'title', got %q", m.Title)
	}
}

func TestMedia_ConcurrentUpdateProgress(t *testing.T) {
	m := &domain.Media{ID: "1"}
	var wg sync.WaitGroup
	n := 100
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			m.UpdateProgress(int64(i), 100, i)
		}(i)
	}
	wg.Wait()
	// Just verify no race/panic occurred
}

func TestMedia_ConcurrentSetStatus(t *testing.T) {
	m := &domain.Media{ID: "1"}
	var wg sync.WaitGroup
	n := 100
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			m.SetStatus(domain.DownloadStatus(i % 5))
		}(i)
	}
	wg.Wait()
	// Just verify no race/panic occurred
}

func TestMedia_ConcurrentMixedOperations(t *testing.T) {
	m := &domain.Media{
		ID:             "1",
		OnProgress:     func(id string, p domain.DownloadProgress) {},
		OnStatusChange: func(id string, s domain.DownloadStatus) {},
		OnTitleChange:  func(id string, title string) {},
	}
	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			m.AppendLog("log")
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			m.SetTitle("title")
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			m.UpdateProgress(int64(i), 100, i)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			m.SetStatus(domain.InProgress)
		}
	}()
	wg.Wait()
}

func TestMedia_ZeroValue_SafeOperations(t *testing.T) {
	var m domain.Media
	// All operations on zero-value Media should not panic
	m.AppendLog("test")
	m.SetTitle("title")
	m.UpdateProgress(0, 0, 0)
	m.SetStatus(domain.Pending)
	m.Cancel()
}

func TestMedia_ProgressLogsInitiallyEmpty(t *testing.T) {
	m := &domain.Media{ID: "1"}
	if m.Progress.Logs != nil {
		t.Error("expected nil logs initially")
	}
	if m.Progress.Percentage != 0 {
		t.Errorf("expected 0 percentage, got %d", m.Progress.Percentage)
	}
	if m.Progress.DownloadedBytes != 0 {
		t.Errorf("expected 0 downloaded bytes, got %d", m.Progress.DownloadedBytes)
	}
}

func TestVideoQuality_Values(t *testing.T) {
	if domain.Quality360p != 0 {
		t.Errorf("Quality360p should be 0, got %d", domain.Quality360p)
	}
	if domain.Quality2160p != 5 {
		t.Errorf("Quality2160p should be 5, got %d", domain.Quality2160p)
	}
}

func TestDownloadStatus_Values(t *testing.T) {
	if domain.Pending != 0 {
		t.Errorf("Pending should be 0, got %d", domain.Pending)
	}
	if domain.Paused != 4 {
		t.Errorf("Paused should be 4, got %d", domain.Paused)
	}
}

func TestDownloadStatus_Sequence(t *testing.T) {
	expected := []domain.DownloadStatus{
		domain.Pending,
		domain.InProgress,
		domain.Completed,
		domain.Failed,
		domain.Paused,
	}
	for i, s := range expected {
		if int(s) != i {
			t.Errorf("status %d has value %d, expected %d", i, s, i)
		}
	}
}

// ===========================================================================
// PlaylistSelection.Validate
// ===========================================================================

// ===========================================================================
// TimeRange.Validate
// ===========================================================================

func TestTimeRange_Validate_Valid_ReturnsNil(t *testing.T) {
	tr := domain.TimeRange{Start: "00:01:02", End: "01:23:45"}
	if err := tr.Validate(); err != nil {
		t.Errorf("expected no error for valid time range, got: %v", err)
	}
}

func TestTimeRange_Validate_MissingStart_ReturnsError(t *testing.T) {
	tr := domain.TimeRange{End: "01:23:45"}
	if err := tr.Validate(); err == nil {
		t.Error("expected error when start is empty, got nil")
	}
}

func TestTimeRange_Validate_MissingEnd_ReturnsError(t *testing.T) {
	tr := domain.TimeRange{Start: "00:01:02"}
	if err := tr.Validate(); err == nil {
		t.Error("expected error when end is empty, got nil")
	}
}

func TestTimeRange_Validate_InvalidStartFormat_ReturnsError(t *testing.T) {
	tr := domain.TimeRange{Start: "1:2:3", End: "01:23:45"}
	err := tr.Validate()
	if err == nil {
		t.Fatal("expected error for invalid start format, got nil")
	}
	const want = "invalid time range start"
	if !containsSubstring(err.Error(), want) {
		t.Errorf("expected error message to contain %q, got: %q", want, err.Error())
	}
}

func TestTimeRange_Validate_InvalidEndFormat_ReturnsError(t *testing.T) {
	tr := domain.TimeRange{Start: "00:01:02", End: "1:2:3"}
	err := tr.Validate()
	if err == nil {
		t.Fatal("expected error for invalid end format, got nil")
	}
	const want = "invalid time range end"
	if !containsSubstring(err.Error(), want) {
		t.Errorf("expected error message to contain %q, got: %q", want, err.Error())
	}
}

// --- SelectionAll ----------------------------------------------------------

func TestPlaylistSelection_Validate_All_ReturnsNil(t *testing.T) {
	ps := domain.PlaylistSelection{Type: domain.SelectionAll}
	if err := ps.Validate(); err != nil {
		t.Errorf("expected nil error for SelectionAll, got: %v", err)
	}
}

// --- Unknown / empty type --------------------------------------------------

func TestPlaylistSelection_Validate_EmptyType_ReturnsNil(t *testing.T) {
	ps := domain.PlaylistSelection{} // zero value: Type == ""
	if err := ps.Validate(); err != nil {
		t.Errorf("expected nil error for empty type, got: %v", err)
	}
}

func TestPlaylistSelection_Validate_UnknownType_ReturnsNil(t *testing.T) {
	ps := domain.PlaylistSelection{Type: domain.PlaylistSelectionType("unknown")}
	if err := ps.Validate(); err != nil {
		t.Errorf("expected nil error for unknown type, got: %v", err)
	}
}

// --- SelectionRange --------------------------------------------------------

func TestPlaylistSelection_Validate_Range_Valid(t *testing.T) {
	ps := domain.PlaylistSelection{
		Type:       domain.SelectionRange,
		StartIndex: 1,
		EndIndex:   5,
	}
	if err := ps.Validate(); err != nil {
		t.Errorf("expected no error for valid range [1,5], got: %v", err)
	}
}

func TestPlaylistSelection_Validate_Range_EqualStartEnd_Valid(t *testing.T) {
	// StartIndex == EndIndex is a single-item range, should be valid
	ps := domain.PlaylistSelection{
		Type:       domain.SelectionRange,
		StartIndex: 3,
		EndIndex:   3,
	}
	if err := ps.Validate(); err != nil {
		t.Errorf("expected no error for equal start/end [3,3], got: %v", err)
	}
}

func TestPlaylistSelection_Validate_Range_StartIndexZero_ReturnsError(t *testing.T) {
	ps := domain.PlaylistSelection{
		Type:       domain.SelectionRange,
		StartIndex: 0,
		EndIndex:   5,
	}
	if err := ps.Validate(); err == nil {
		t.Error("expected error when StartIndex is 0, got nil")
	}
}

func TestPlaylistSelection_Validate_Range_NegativeStartIndex_ReturnsError(t *testing.T) {
	ps := domain.PlaylistSelection{
		Type:       domain.SelectionRange,
		StartIndex: -1,
		EndIndex:   5,
	}
	if err := ps.Validate(); err == nil {
		t.Error("expected error when StartIndex is negative, got nil")
	}
}

func TestPlaylistSelection_Validate_Range_EndIndexLessThanStart_ReturnsError(t *testing.T) {
	ps := domain.PlaylistSelection{
		Type:       domain.SelectionRange,
		StartIndex: 5,
		EndIndex:   3,
	}
	if err := ps.Validate(); err == nil {
		t.Error("expected error when EndIndex < StartIndex, got nil")
	}
}

func TestPlaylistSelection_Validate_Range_BothZero_ReturnsError(t *testing.T) {
	ps := domain.PlaylistSelection{
		Type:       domain.SelectionRange,
		StartIndex: 0,
		EndIndex:   0,
	}
	if err := ps.Validate(); err == nil {
		t.Error("expected error for range [0,0], got nil")
	}
}

func TestPlaylistSelection_Validate_Range_ErrorMessageContainsIndexes(t *testing.T) {
	ps := domain.PlaylistSelection{
		Type:       domain.SelectionRange,
		StartIndex: 0,
		EndIndex:   0,
	}
	err := ps.Validate()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	msg := err.Error()
	for _, keyword := range []string{"invalid playlist range", "0", "0"} {
		if !containsSubstring(msg, keyword) {
			t.Errorf("expected error message to contain %q, got: %q", keyword, msg)
		}
	}
}

// --- SelectionItems --------------------------------------------------------

func TestPlaylistSelection_Validate_Items_NonEmpty_Valid(t *testing.T) {
	ps := domain.PlaylistSelection{
		Type:  domain.SelectionItems,
		Items: "1,3,5",
	}
	if err := ps.Validate(); err != nil {
		t.Errorf("expected no error for non-empty items, got: %v", err)
	}
}

func TestPlaylistSelection_Validate_Items_SingleItem_Valid(t *testing.T) {
	ps := domain.PlaylistSelection{
		Type:  domain.SelectionItems,
		Items: "7",
	}
	if err := ps.Validate(); err != nil {
		t.Errorf("expected no error for single item, got: %v", err)
	}
}

func TestPlaylistSelection_Validate_Items_Empty_ReturnsError(t *testing.T) {
	ps := domain.PlaylistSelection{
		Type:  domain.SelectionItems,
		Items: "",
	}
	if err := ps.Validate(); err == nil {
		t.Error("expected error for empty items list, got nil")
	}
}

func TestPlaylistSelection_Validate_Items_ErrorMessageDescriptive(t *testing.T) {
	ps := domain.PlaylistSelection{
		Type:  domain.SelectionItems,
		Items: "",
	}
	err := ps.Validate()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	const want = "items selection type requires a non-empty items list"
	if !containsSubstring(err.Error(), want) {
		t.Errorf("expected error message to contain %q, got: %q", want, err.Error())
	}
}

// --- helpers ---------------------------------------------------------------

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		})())
}
