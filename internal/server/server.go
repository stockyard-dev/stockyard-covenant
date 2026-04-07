package server

import (
	"encoding/json"
	"github.com/stockyard-dev/stockyard-covenant/internal/store"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Server struct {
	db      *store.DB
	mux     *http.ServeMux
	limits  Limits
	dataDir string
	pCfg    map[string]json.RawMessage
}

func New(db *store.DB, limits Limits, dataDir string) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), limits: limits, dataDir: dataDir}

	// Policies
	s.mux.HandleFunc("GET /api/policies", s.listPolicies)
	s.mux.HandleFunc("POST /api/policies", s.createPolicy)
	s.mux.HandleFunc("GET /api/policies/{id}", s.getPolicy)
	s.mux.HandleFunc("PUT /api/policies/{id}", s.updatePolicy)
	s.mux.HandleFunc("DELETE /api/policies/{id}", s.deletePolicy)
	s.mux.HandleFunc("GET /api/policies/{id}/versions", s.listVersions)
	s.mux.HandleFunc("GET /api/policies/{id}/acks", s.listAcks)
	s.mux.HandleFunc("GET /api/policies/{id}/pending", s.pendingAcks)
	s.mux.HandleFunc("GET /api/policies/{id}/evidence", s.listEvidence)
	s.mux.HandleFunc("POST /api/policies/{id}/evidence", s.createEvidence)

	// Members
	s.mux.HandleFunc("GET /api/members", s.listMembers)
	s.mux.HandleFunc("POST /api/members", s.createMember)
	s.mux.HandleFunc("GET /api/members/{id}", s.getMember)
	s.mux.HandleFunc("DELETE /api/members/{id}", s.deleteMember)

	// Acknowledge
	s.mux.HandleFunc("POST /api/acknowledge", s.acknowledge)

	// Evidence
	s.mux.HandleFunc("DELETE /api/evidence/{id}", s.deleteEvidence)

	// Meta
	s.mux.HandleFunc("GET /api/categories", s.categories)
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)

	// Dashboard
	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)
	s.mux.HandleFunc("GET /", s.root)
	s.mux.HandleFunc("GET /api/tier", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"tier": s.limits.Tier, "upgrade_url": "https://stockyard.dev/covenant/"})
	})

	s.loadPersonalConfig()
	s.mux.HandleFunc("GET /api/config", s.configHandler)
	s.mux.HandleFunc("GET /api/extras/{resource}", s.listExtras)
	s.mux.HandleFunc("GET /api/extras/{resource}/{id}", s.getExtras)
	s.mux.HandleFunc("PUT /api/extras/{resource}/{id}", s.putExtras)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func (s *Server) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/ui", http.StatusFound)
}

// ── Policies ──

func (s *Server) listPolicies(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	category := r.URL.Query().Get("category")
	writeJSON(w, 200, map[string]any{"policies": orEmpty(s.db.ListPolicies(status, category))})
}

func (s *Server) createPolicy(w http.ResponseWriter, r *http.Request) {
	var p store.Policy
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeErr(w, 400, "invalid json")
		return
	}
	if p.Title == "" {
		writeErr(w, 400, "title required")
		return
	}
	p.RequiresAck = true // default
	if err := s.db.CreatePolicy(&p); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 201, s.db.GetPolicy(p.ID))
}

func (s *Server) getPolicy(w http.ResponseWriter, r *http.Request) {
	p := s.db.GetPolicy(r.PathValue("id"))
	if p == nil {
		writeErr(w, 404, "not found")
		return
	}
	writeJSON(w, 200, p)
}

func (s *Server) updatePolicy(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ex := s.db.GetPolicy(id)
	if ex == nil {
		writeErr(w, 404, "not found")
		return
	}
	var req struct {
		Title       string `json:"title"`
		Body        string `json:"body"`
		Category    string `json:"category"`
		Status      string `json:"status"`
		Owner       string `json:"owner"`
		RequiresAck *bool  `json:"requires_ack"`
		AckDeadline string `json:"ack_deadline"`
		Actor       string `json:"actor"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	p := store.Policy{
		Title:       req.Title,
		Body:        req.Body,
		Category:    req.Category,
		Status:      req.Status,
		Owner:       req.Owner,
		RequiresAck: ex.RequiresAck,
		AckDeadline: req.AckDeadline,
	}
	if p.Title == "" {
		p.Title = ex.Title
	}
	if p.Body == "" {
		p.Body = ex.Body
	}
	if p.Category == "" {
		p.Category = ex.Category
	}
	if p.Status == "" {
		p.Status = ex.Status
	}
	if p.Owner == "" {
		p.Owner = ex.Owner
	}
	if req.RequiresAck != nil {
		p.RequiresAck = *req.RequiresAck
	}
	if err := s.db.UpdatePolicy(id, &p, req.Actor); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, s.db.GetPolicy(id))
}

func (s *Server) deletePolicy(w http.ResponseWriter, r *http.Request) {
	s.db.DeletePolicy(r.PathValue("id"))
	writeJSON(w, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) listVersions(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"versions": orEmpty(s.db.ListPolicyVersions(r.PathValue("id")))})
}

func (s *Server) listAcks(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"acknowledgments": orEmpty(s.db.ListAcknowledgments(r.PathValue("id")))})
}

func (s *Server) pendingAcks(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"pending": orEmpty(s.db.PendingAcks(r.PathValue("id")))})
}

// ── Members ──

func (s *Server) listMembers(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"members": orEmpty(s.db.ListMembers())})
}

func (s *Server) createMember(w http.ResponseWriter, r *http.Request) {
	var m store.Member
	json.NewDecoder(r.Body).Decode(&m)
	if m.Name == "" {
		writeErr(w, 400, "name required")
		return
	}
	s.db.CreateMember(&m)
	writeJSON(w, 201, s.db.GetMember(m.ID))
}

func (s *Server) getMember(w http.ResponseWriter, r *http.Request) {
	m := s.db.GetMember(r.PathValue("id"))
	if m == nil {
		writeErr(w, 404, "not found")
		return
	}
	writeJSON(w, 200, m)
}

func (s *Server) deleteMember(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteMember(r.PathValue("id"))
	writeJSON(w, 200, map[string]string{"deleted": "ok"})
}

// ── Acknowledge ──

func (s *Server) acknowledge(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PolicyID string `json:"policy_id"`
		MemberID string `json:"member_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.PolicyID == "" || req.MemberID == "" {
		writeErr(w, 400, "policy_id and member_id required")
		return
	}
	if err := s.db.Acknowledge(req.PolicyID, req.MemberID); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"acknowledged": "ok"})
}

// ── Evidence ──

func (s *Server) listEvidence(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"evidence": orEmpty(s.db.ListEvidence(r.PathValue("id")))})
}

func (s *Server) createEvidence(w http.ResponseWriter, r *http.Request) {
	policyID := r.PathValue("id")
	if s.db.GetPolicy(policyID) == nil {
		writeErr(w, 404, "policy not found")
		return
	}
	var e store.Evidence
	json.NewDecoder(r.Body).Decode(&e)
	if e.Title == "" {
		writeErr(w, 400, "title required")
		return
	}
	e.PolicyID = policyID
	s.db.CreateEvidence(&e)
	writeJSON(w, 201, e)
}

func (s *Server) deleteEvidence(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteEvidence(r.PathValue("id"))
	writeJSON(w, 200, map[string]string{"deleted": "ok"})
}

// ── Meta ──

func (s *Server) categories(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"categories": orEmpty(s.db.Categories())})
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, s.db.Stats())
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	st := s.db.Stats()
	writeJSON(w, 200, map[string]any{
		"status":     "ok",
		"service":    "covenant",
		"policies":   st.Policies,
		"compliance": st.OverallCompliance,
	})
}

func orEmpty[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// ─── personalization (auto-added) ──────────────────────────────────

func (s *Server) loadPersonalConfig() {
	path := filepath.Join(s.dataDir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("%s: warning: could not parse config.json: %v", "covenant", err)
		return
	}
	s.pCfg = cfg
	log.Printf("%s: loaded personalization from %s", "covenant", path)
}

func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	if s.pCfg == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.pCfg)
}

func (s *Server) listExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	all := s.db.AllExtras(resource)
	out := make(map[string]json.RawMessage, len(all))
	for id, data := range all {
		out[id] = json.RawMessage(data)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (s *Server) getExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	id := r.PathValue("id")
	data := s.db.GetExtras(resource, id)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(data))
}

func (s *Server) putExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	id := r.PathValue("id")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"read body"}`, 400)
		return
	}
	var probe map[string]any
	if err := json.Unmarshal(body, &probe); err != nil {
		http.Error(w, `{"error":"invalid json"}`, 400)
		return
	}
	if err := s.db.SetExtras(resource, id, string(body)); err != nil {
		http.Error(w, `{"error":"save failed"}`, 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":"saved"}`))
}
