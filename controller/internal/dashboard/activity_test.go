package dashboard

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func waitForActivity(t *testing.T, check func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !check() {
		if time.Now().After(deadline) {
			t.Fatal("activity writer did not reach expected state")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestActivitySurvivesRestartWithStableIDsAndSessions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "activity.sqlite")
	s := testServer(t)
	if err := s.OpenActivity(path, 100); err != nil {
		t.Fatal(err)
	}
	s.event("Home saved before restart")
	original := s.events[0]
	s.Close() // flushes even if the 100ms batch timer has not fired
	restarted := testServer(t)
	if err := restarted.OpenActivity(path, 100); err != nil {
		t.Fatal(err)
	}
	if restarted.session == s.session {
		t.Fatal("restart reused session ID")
	}
	if restarted.events[0].ID <= original.ID {
		t.Fatal("event IDs reset on restart")
	}
	found := false
	for _, e := range restarted.events {
		if e.ID == original.ID {
			found = true
			if e.Text != original.Text || e.Session != original.Session || !e.Time.Equal(original.Time) {
				t.Fatal("restored event changed")
			}
		}
	}
	if !found {
		t.Fatal("previous session missing from recent activity")
	}
}

func TestActivityRetentionAndCursorPagination(t *testing.T) {
	s := testServer(t)
	if err := s.OpenActivity("-", 60); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 80; i++ {
		s.event(fmt.Sprintf("event %02d", i))
	}
	waitForActivity(t, func() bool { return s.activityStatus().Queued == 0 })
	if s.activityStatus().Retained != 60 || len(s.events) != 12 {
		t.Fatal(s.activityStatus(), len(s.events))
	}
	first, err := s.activity.page(context.Background(), 0, 25)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Events) != 25 || first.NextBefore == 0 {
		t.Fatal(first)
	}
	s.event("new activity while browsing")
	waitForActivity(t, func() bool { return s.activityStatus().Queued == 0 })
	second, err := s.activity.page(context.Background(), first.NextBefore, 100)
	if err != nil {
		t.Fatal(err)
	}
	if second.NextBefore != 0 || len(second.Events) != 34 {
		t.Fatal("wrong retained remainder", len(second.Events))
	}
	seen := map[int64]bool{}
	for _, e := range first.Events {
		seen[e.ID] = true
	}
	for _, e := range second.Events {
		if seen[e.ID] || e.ID >= first.NextBefore {
			t.Fatal("pagination duplicated or reordered an event")
		}
		seen[e.ID] = true
	}
}

func TestActivityRetriesFailedWritesAndSurfacesError(t *testing.T) {
	s := testServer(t)
	if err := s.OpenActivity("-", 100); err != nil {
		t.Fatal(err)
	}
	waitForActivity(t, func() bool { return s.activityStatus().Queued == 0 })
	_, err := s.activity.db.Exec(`CREATE TRIGGER fail_activity BEFORE INSERT ON activity BEGIN SELECT RAISE(ABORT,'test storage failure'); END`)
	if err != nil {
		t.Fatal(err)
	}
	s.event("must survive transient storage failure")
	waitForActivity(t, func() bool { return s.activityStatus().Error != "" })
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/state", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "test storage failure") {
		t.Fatal(w.Body.String())
	}
	// Stops continue to be accepted while persistence is retrying off-loop.
	if _, err = s.execute("/api/stop", command{}); err != nil {
		t.Fatal(err)
	}
	if _, err = s.activity.db.Exec(`DROP TRIGGER fail_activity`); err != nil {
		t.Fatal(err)
	}
	waitForActivity(t, func() bool { return s.activityStatus().Queued == 0 })
	if s.activityStatus().Error != "" {
		t.Fatal("recovered storage still reports an error")
	}
	p, err := s.activity.page(context.Background(), 0, 100)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, e := range p.Events {
		if e.Text == "must survive transient storage failure" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("retried event stored %d times", count)
	}
}

func TestActivityQueueIsBoundedAndReportsOverflow(t *testing.T) {
	a := &activityStore{queue: make(chan Event, 2)}
	a.enqueue(Event{ID: 1})
	a.enqueue(Event{ID: 2})
	a.enqueue(Event{ID: 3})
	status := a.statusSnapshot()
	if status.Queued != 2 || status.Dropped != 1 {
		t.Fatal(status)
	}
}

func TestActivitySeparatesModesAndRejectsInvalidDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.sqlite")
	a, _, _, err := openActivity(path, true, 100)
	if err != nil {
		t.Fatal(err)
	}
	if err = a.close(); err != nil {
		t.Fatal(err)
	}
	if b, _, _, err := openActivity(path, false, 100); err == nil {
		b.close()
		t.Fatal("mixed hardware and simulation history")
	}
	bad := filepath.Join(t.TempDir(), "broken.sqlite")
	data := []byte("this is not an activity database")
	if err = os.WriteFile(bad, data, 0600); err != nil {
		t.Fatal(err)
	}
	if b, _, _, err := openActivity(bad, true, 100); err == nil {
		b.close()
		t.Fatal("invalid database accepted")
	}
	got, _ := os.ReadFile(bad)
	if string(got) != string(data) {
		t.Fatal("invalid file overwritten")
	}
}

func TestActivityDatabaseHasOneWriterOwner(t *testing.T) {
	path := filepath.Join(t.TempDir(), "activity.sqlite")
	a, _, _, err := openActivity(path, true, 100)
	if err != nil {
		t.Fatal(err)
	}
	defer a.close()
	if b, _, _, err := openActivity(path, true, 100); err == nil {
		b.close()
		t.Fatal("two services could allocate colliding event IDs")
	}
}

func TestActivityEndpointValidatesPagingAndReturnsHistory(t *testing.T) {
	s := testServer(t)
	if err := s.OpenActivity("-", 100); err != nil {
		t.Fatal(err)
	}
	s.event("history outside the recent preview")
	for i := 0; i < 20; i++ {
		s.event(fmt.Sprintf("later %d", i))
	}
	waitForActivity(t, func() bool { return s.activityStatus().Queued == 0 })
	for _, query := range []string{"limit=0", "limit=101", "limit=bad", "before=-1", "before=bad"} {
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/activity?"+query, nil))
		if w.Code != 400 {
			t.Fatalf("accepted %s", query)
		}
	}
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/activity?limit=100", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "history outside the recent preview") {
		t.Fatal(w.Body.String())
	}
}
