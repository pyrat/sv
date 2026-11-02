package rope

import (
	_ "embed"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

//go:embed ui.html
var page string

func (c *Controller) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /api/trace.csv", func(w http.ResponseWriter, r *http.Request) {
		t := c.Snapshot().LastMove
		if t == nil {
			http.Error(w, "no move trace recorded", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="depthpilot-`+t.ID+`.csv"`)
		_ = t.CSV(w)
	})
	m.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, page)
	})
	m.HandleFunc("GET /api/state", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(c.Snapshot())
	})
	for _, op := range []string{"reference", "move", "stop"} {
		m.HandleFunc("POST /api/"+op, func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
				http.Error(w, "application/json required", 415)
				return
			}
			if origin := r.Header.Get("Origin"); origin != "" {
				u, e := url.Parse(origin)
				if e != nil || u.Host != r.Host {
					http.Error(w, "same-origin requests required", 403)
					return
				}
			}
			if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
				http.Error(w, "cross-site requests rejected", 403)
				return
			}
			var err error
			if op == "move" {
				var body struct {
					X     *float64 `json:"x"`
					Z     *float64 `json:"z"`
					Speed *float64 `json:"speed"`
				}
				d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
				d.DisallowUnknownFields()
				if e := d.Decode(&body); e != nil || body.X == nil || body.Z == nil || body.Speed == nil {
					http.Error(w, "x, z and speed are required numbers", 400)
					return
				}
				if d.Decode(new(any)) != io.EOF {
					http.Error(w, "one JSON object required", 400)
					return
				}
				err = c.Move(Point{*body.X, *body.Z}, *body.Speed)
			} else if op == "reference" {
				err = c.Reference()
			} else {
				err = c.Stop()
			}
			if err != nil {
				http.Error(w, err.Error(), 409)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(c.Snapshot())
		})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'unsafe-inline'; style-src 'unsafe-inline'; frame-ancestors 'none'")
		m.ServeHTTP(w, r)
	})
}
