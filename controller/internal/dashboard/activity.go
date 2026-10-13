package dashboard

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

const recentActivityLimit = 12

type ActivityStatus struct {
	Persistent bool   `json:"persistent"`
	Retained   int    `json:"retained"`
	Limit      int    `json:"limit"`
	Queued     int    `json:"queued"`
	Dropped    int    `json:"dropped"`
	Error      string `json:"error,omitempty"`
}

type ActivityPage struct {
	Events     []Event        `json:"events"`
	NextBefore int64          `json:"nextBefore"`
	Storage    ActivityStatus `json:"storage"`
}

// activityStore owns the database writer. The control loop only enqueues into
// bounded memory; it never waits for an INSERT, fsync, checkpoint, or reader.
type activityStore struct {
	db       *sql.DB
	queue    chan Event
	stopping chan struct{}
	done     chan struct{}
	mu       sync.Mutex
	status   ActivityStatus
	closed   bool
	closeErr error
}

func openActivity(path string, simulation bool, limit int) (*activityStore, []Event, int64, error) {
	if limit < 1 || limit > 1000000 {
		return nil, nil, 0, fmt.Errorf("activity limit must be 1..1000000")
	}
	if path == "" {
		return nil, nil, 0, fmt.Errorf("activity path required; use '-' for memory only")
	}
	dsn := ":memory:"
	if path != "-" {
		abs, err := filepath.Abs(path)
		if err != nil {
			return nil, nil, 0, err
		}
		if err = os.MkdirAll(filepath.Dir(abs), 0700); err != nil {
			return nil, nil, 0, err
		}
		f, err := os.OpenFile(abs, os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			return nil, nil, 0, err
		}
		if err = f.Close(); err != nil {
			return nil, nil, 0, err
		}
		dsn = (&url.URL{Scheme: "file", Path: abs}).String()
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, nil, 0, err
	}
	db.SetMaxOpenConns(1)
	fail := func(err error) (*activityStore, []Event, int64, error) { db.Close(); return nil, nil, 0, err }
	// One control service owns each file. Holding the SQLite lock also prevents
	// a second service from allocating conflicting IDs from the same history.
	if _, err = db.Exec(`PRAGMA busy_timeout=1000; PRAGMA locking_mode=EXCLUSIVE; PRAGMA journal_mode=WAL; PRAGMA synchronous=FULL; PRAGMA journal_size_limit=8388608;`); err != nil {
		return fail(err)
	}
	var version int
	if err = db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return fail(err)
	}
	if version > 1 {
		return fail(fmt.Errorf("unsupported activity database version %d", version))
	}
	if _, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS activity_meta (key TEXT PRIMARY KEY, value TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS activity (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			time TEXT NOT NULL,
			text TEXT NOT NULL,
			session TEXT NOT NULL
		);
		PRAGMA user_version=1;
	`); err != nil {
		return fail(err)
	}
	mode := "hardware"
	if simulation {
		mode = "simulation"
	}
	if _, err = db.Exec(`INSERT OR IGNORE INTO activity_meta(key,value) VALUES ('mode',?)`, mode); err != nil {
		return fail(err)
	}
	var storedMode string
	if err = db.QueryRow(`SELECT value FROM activity_meta WHERE key='mode'`).Scan(&storedMode); err != nil {
		return fail(err)
	}
	if storedMode != mode {
		return fail(fmt.Errorf("activity database is for %s, not %s", storedMode, mode))
	}
	if _, err = db.Exec(retainActivitySQL, limit); err != nil {
		return fail(err)
	}
	a := &activityStore{db: db, queue: make(chan Event, 1024), stopping: make(chan struct{}), done: make(chan struct{}), status: ActivityStatus{Persistent: path != "-", Limit: limit}}
	var lastID int64
	if err = db.QueryRow(`SELECT COALESCE(MAX(id),0), COUNT(*) FROM activity`).Scan(&lastID, &a.status.Retained); err != nil {
		return fail(err)
	}
	page, err := a.page(context.Background(), 0, recentActivityLimit)
	if err != nil {
		return fail(err)
	}
	go a.run()
	return a, page.Events, lastID, nil
}

const retainActivitySQL = `DELETE FROM activity WHERE id <= COALESCE((SELECT id FROM activity ORDER BY id DESC LIMIT 1 OFFSET ?),0)`

func (a *activityStore) statusSnapshot() ActivityStatus {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.status
}

func (a *activityStore) enqueue(e Event) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return
	}
	select {
	case a.queue <- e:
		a.status.Queued++
	default:
		a.status.Dropped++
		log.Printf("activity queue full; event not stored: %s", e.Text)
	}
}

func (a *activityStore) writeBatch(batch []Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO activity(id,time,text,session) VALUES (?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, e := range batch {
		if _, err = stmt.ExecContext(ctx, e.ID, e.Time.UTC().Format(time.RFC3339Nano), e.Text, e.Session); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, retainActivitySQL, a.statusSnapshot().Limit); err != nil {
		return err
	}
	var count int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM activity`).Scan(&count); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	a.mu.Lock()
	a.status.Retained = count
	a.status.Queued -= len(batch)
	a.status.Error = ""
	a.mu.Unlock()
	return nil
}

func (a *activityStore) run() {
	defer close(a.done)
	for first := range a.queue {
		batch := []Event{first}
		// Batch short bursts of stage/stop events into one durable transaction.
		timer := time.NewTimer(100 * time.Millisecond)
	collect:
		for len(batch) < 128 {
			select {
			case e, ok := <-a.queue:
				if !ok {
					break collect
				}
				batch = append(batch, e)
			case <-timer.C:
				break collect
			}
		}
		timer.Stop()
		for {
			err := a.writeBatch(batch)
			if err == nil {
				break
			}
			a.mu.Lock()
			changed := a.status.Error != err.Error()
			a.status.Error = err.Error()
			a.mu.Unlock()
			if changed {
				log.Printf("activity storage: %v", err)
			}
			// Retain this batch and retry transient storage errors in order.
			select {
			case <-a.stopping:
				a.closeErr = err
				return
			case <-time.After(250 * time.Millisecond):
			}
		}
	}
}

func (a *activityStore) close() error {
	a.mu.Lock()
	if !a.closed {
		a.closed = true
		close(a.stopping)
		close(a.queue)
	}
	a.mu.Unlock()
	<-a.done
	return errors.Join(a.closeErr, a.db.Close())
}

func (a *activityStore) page(ctx context.Context, before int64, limit int) (ActivityPage, error) {
	p := ActivityPage{Events: []Event{}}
	rows, err := a.db.QueryContext(ctx, `SELECT id,time,text,session FROM activity WHERE (?=0 OR id<?) ORDER BY id DESC LIMIT ?`, before, before, limit+1)
	if err != nil {
		return p, err
	}
	defer rows.Close()
	for rows.Next() {
		var e Event
		var timestamp string
		if err = rows.Scan(&e.ID, &timestamp, &e.Text, &e.Session); err != nil {
			return p, err
		}
		e.Time, err = time.Parse(time.RFC3339Nano, timestamp)
		if err != nil {
			return p, fmt.Errorf("invalid activity timestamp: %w", err)
		}
		p.Events = append(p.Events, e)
	}
	if err = rows.Err(); err != nil {
		return p, err
	}
	if len(p.Events) > limit {
		p.Events = p.Events[:limit]
		p.NextBefore = p.Events[len(p.Events)-1].ID
	}
	p.Storage = a.statusSnapshot()
	return p, nil
}

// OpenActivity restores recent history and queues this startup's events after it.
// Call once before serving requests or starting the control loop.
func (s *Server) OpenActivity(path string, limit int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.activity != nil {
		return fmt.Errorf("activity storage is already open")
	}
	a, recent, last, err := openActivity(path, s.simulation, limit)
	if err != nil {
		return fmt.Errorf("open activity: %w", err)
	}
	startup := s.events
	s.activity = a
	s.events = recent
	s.nextEventID = last
	for i := len(startup) - 1; i >= 0; i-- {
		e := startup[i]
		s.nextEventID++
		e.ID = s.nextEventID
		s.appendEvent(e)
	}
	return nil
}

func (s *Server) appendEvent(e Event) {
	s.events = append([]Event{e}, s.events...)
	if len(s.events) > recentActivityLimit {
		s.events = s.events[:recentActivityLimit]
	}
	if s.activity != nil {
		s.activity.enqueue(e)
	}
}

func (s *Server) activityStatus() ActivityStatus {
	if s.activity != nil {
		return s.activity.statusSnapshot()
	}
	return ActivityStatus{Retained: len(s.events), Limit: recentActivityLimit}
}

func (s *Server) activityHandler(w http.ResponseWriter, r *http.Request) {
	limit := 50
	before := int64(0)
	if value := r.URL.Query().Get("limit"); value != "" {
		v, err := strconv.Atoi(value)
		if err != nil || v < 1 || v > 100 {
			writeJSON(w, 400, map[string]string{"error": "limit must be 1..100"})
			return
		}
		limit = v
	}
	if value := r.URL.Query().Get("before"); value != "" {
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil || v < 1 {
			writeJSON(w, 400, map[string]string{"error": "before must be a positive event ID"})
			return
		}
		before = v
	}
	s.mu.Lock()
	a := s.activity
	if a == nil {
		p := ActivityPage{Events: []Event{}, Storage: s.activityStatus()}
		for _, e := range s.events {
			if before == 0 || e.ID < before {
				p.Events = append(p.Events, e)
			}
		}
		if len(p.Events) > limit {
			p.Events = p.Events[:limit]
			p.NextBefore = p.Events[len(p.Events)-1].ID
		}
		s.mu.Unlock()
		writeJSON(w, 200, p)
		return
	}
	s.mu.Unlock()
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	p, err := a.page(ctx, before, limit)
	if err != nil {
		writeJSON(w, 503, map[string]string{"error": "Activity history is temporarily unavailable"})
		return
	}
	writeJSON(w, 200, p)
}
