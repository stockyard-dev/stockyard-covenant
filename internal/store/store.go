package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct{ db *sql.DB }

type Policy struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Body         string  `json:"body"`
	Category     string  `json:"category,omitempty"`
	Version      int     `json:"version"`
	Status       string  `json:"status"` // draft, active, retired
	Owner        string  `json:"owner,omitempty"`
	RequiresAck  bool    `json:"requires_ack"`
	AckDeadline  string  `json:"ack_deadline,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
	AckCount     int     `json:"ack_count"`
	MemberCount  int     `json:"member_count"`
	CompliancePct float64 `json:"compliance_pct"`
	EvidenceCount int    `json:"evidence_count"`
}

type Member struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email,omitempty"`
	Department string `json:"department,omitempty"`
	CreatedAt  string `json:"created_at"`
	AckCount   int    `json:"ack_count"`
	PendingAck int    `json:"pending_ack"`
}

type Acknowledgment struct {
	ID            string `json:"id"`
	PolicyID      string `json:"policy_id"`
	PolicyTitle   string `json:"policy_title,omitempty"`
	PolicyVersion int    `json:"policy_version"`
	MemberID      string `json:"member_id"`
	MemberName    string `json:"member_name,omitempty"`
	AckedAt       string `json:"acked_at"`
}

type Evidence struct {
	ID          string `json:"id"`
	PolicyID    string `json:"policy_id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	FileName    string `json:"file_name,omitempty"`
	CollectedAt string `json:"collected_at"`
	CollectedBy string `json:"collected_by,omitempty"`
	CreatedAt   string `json:"created_at"`
}

type PolicyVersion struct {
	ID        string `json:"id"`
	PolicyID  string `json:"policy_id"`
	Version   int    `json:"version"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	ChangedBy string `json:"changed_by,omitempty"`
	CreatedAt string `json:"created_at"`
}

func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	dsn := filepath.Join(dataDir, "covenant.db") + "?_journal_mode=WAL&_busy_timeout=5000"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	for _, q := range []string{
		`CREATE TABLE IF NOT EXISTS policies (
			id TEXT PRIMARY KEY, title TEXT NOT NULL, body TEXT DEFAULT '',
			category TEXT DEFAULT '', version INTEGER DEFAULT 1,
			status TEXT DEFAULT 'draft', owner TEXT DEFAULT '',
			requires_ack INTEGER DEFAULT 1, ack_deadline TEXT DEFAULT '',
			created_at TEXT DEFAULT (datetime('now')),
			updated_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS policy_versions (
			id TEXT PRIMARY KEY, policy_id TEXT NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
			version INTEGER NOT NULL, title TEXT DEFAULT '', body TEXT DEFAULT '',
			changed_by TEXT DEFAULT '', created_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS members (
			id TEXT PRIMARY KEY, name TEXT NOT NULL,
			email TEXT DEFAULT '', department TEXT DEFAULT '',
			created_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS acknowledgments (
			id TEXT PRIMARY KEY, policy_id TEXT NOT NULL,
			policy_version INTEGER NOT NULL, member_id TEXT NOT NULL,
			acked_at TEXT DEFAULT (datetime('now')),
			UNIQUE(policy_id, policy_version, member_id)
		)`,
		`CREATE TABLE IF NOT EXISTS evidence (
			id TEXT PRIMARY KEY, policy_id TEXT NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
			title TEXT NOT NULL, description TEXT DEFAULT '',
			file_name TEXT DEFAULT '', collected_at TEXT DEFAULT '',
			collected_by TEXT DEFAULT '', created_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE INDEX IF NOT EXISTS idx_ack_policy ON acknowledgments(policy_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ack_member ON acknowledgments(member_id)`,
		`CREATE INDEX IF NOT EXISTS idx_versions_policy ON policy_versions(policy_id)`,
		`CREATE INDEX IF NOT EXISTS idx_evidence_policy ON evidence(policy_id)`,
	} {
		if _, err := db.Exec(q); err != nil {
			return nil, fmt.Errorf("migrate: %w", err)
		}
	}
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }
func genID() string        { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string          { return time.Now().UTC().Format(time.RFC3339) }

// ── Policies ──

func (d *DB) hydratePolicy(p *Policy) {
	var totalMembers int
	d.db.QueryRow(`SELECT COUNT(*) FROM members`).Scan(&totalMembers)
	p.MemberCount = totalMembers

	// Count acknowledgments for current version
	d.db.QueryRow(`SELECT COUNT(*) FROM acknowledgments WHERE policy_id=? AND policy_version=?`,
		p.ID, p.Version).Scan(&p.AckCount)

	if p.RequiresAck && totalMembers > 0 {
		p.CompliancePct = float64(p.AckCount) / float64(totalMembers) * 100
	} else if !p.RequiresAck {
		p.CompliancePct = 100
	}

	d.db.QueryRow(`SELECT COUNT(*) FROM evidence WHERE policy_id=?`, p.ID).Scan(&p.EvidenceCount)
}

func (d *DB) CreatePolicy(p *Policy) error {
	p.ID = genID()
	p.CreatedAt = now()
	p.UpdatedAt = p.CreatedAt
	if p.Status == "" {
		p.Status = "draft"
	}
	p.Version = 1
	ra := 1
	if !p.RequiresAck {
		ra = 0
	}
	_, err := d.db.Exec(`INSERT INTO policies (id,title,body,category,version,status,owner,requires_ack,ack_deadline,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		p.ID, p.Title, p.Body, p.Category, p.Version, p.Status, p.Owner, ra, p.AckDeadline, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return err
	}
	// Save version 1
	d.db.Exec(`INSERT INTO policy_versions (id,policy_id,version,title,body,changed_by,created_at) VALUES (?,?,?,?,?,?,?)`,
		genID(), p.ID, 1, p.Title, p.Body, p.Owner, p.CreatedAt)
	return nil
}

func (d *DB) scanPolicy(s interface{ Scan(...any) error }) *Policy {
	var p Policy
	var ra int
	if err := s.Scan(&p.ID, &p.Title, &p.Body, &p.Category, &p.Version, &p.Status, &p.Owner, &ra, &p.AckDeadline, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil
	}
	p.RequiresAck = ra == 1
	d.hydratePolicy(&p)
	return &p
}

const policyCols = `id,title,body,category,version,status,owner,requires_ack,ack_deadline,created_at,updated_at`

func (d *DB) GetPolicy(id string) *Policy {
	return d.scanPolicy(d.db.QueryRow(`SELECT `+policyCols+` FROM policies WHERE id=?`, id))
}

func (d *DB) ListPolicies(status, category string) []Policy {
	where := []string{"1=1"}
	args := []any{}
	if status != "" && status != "all" {
		where = append(where, "status=?")
		args = append(args, status)
	}
	if category != "" {
		where = append(where, "category=?")
		args = append(args, category)
	}
	w := strings.Join(where, " AND ")
	rows, err := d.db.Query(`SELECT `+policyCols+` FROM policies WHERE `+w+` ORDER BY category, title`, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []Policy
	for rows.Next() {
		if p := d.scanPolicy(rows); p != nil {
			out = append(out, *p)
		}
	}
	return out
}

func (d *DB) UpdatePolicy(id string, p *Policy, actor string) error {
	old := d.GetPolicy(id)
	if old == nil {
		return fmt.Errorf("not found")
	}

	// If body changed, bump version
	newVersion := old.Version
	if p.Body != old.Body {
		newVersion = old.Version + 1
		d.db.Exec(`INSERT INTO policy_versions (id,policy_id,version,title,body,changed_by,created_at) VALUES (?,?,?,?,?,?,?)`,
			genID(), id, newVersion, p.Title, p.Body, actor, now())
	}

	ra := 1
	if !p.RequiresAck {
		ra = 0
	}
	_, err := d.db.Exec(`UPDATE policies SET title=?,body=?,category=?,version=?,status=?,owner=?,requires_ack=?,ack_deadline=?,updated_at=? WHERE id=?`,
		p.Title, p.Body, p.Category, newVersion, p.Status, p.Owner, ra, p.AckDeadline, now(), id)
	return err
}

func (d *DB) DeletePolicy(id string) error {
	d.db.Exec(`DELETE FROM acknowledgments WHERE policy_id=?`, id)
	d.db.Exec(`DELETE FROM evidence WHERE policy_id=?`, id)
	d.db.Exec(`DELETE FROM policy_versions WHERE policy_id=?`, id)
	_, err := d.db.Exec(`DELETE FROM policies WHERE id=?`, id)
	return err
}

func (d *DB) ListPolicyVersions(policyID string) []PolicyVersion {
	rows, err := d.db.Query(`SELECT id,policy_id,version,title,body,changed_by,created_at FROM policy_versions WHERE policy_id=? ORDER BY version DESC`, policyID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []PolicyVersion
	for rows.Next() {
		var v PolicyVersion
		if err := rows.Scan(&v.ID, &v.PolicyID, &v.Version, &v.Title, &v.Body, &v.ChangedBy, &v.CreatedAt); err != nil {
			continue
		}
		out = append(out, v)
	}
	return out
}

// ── Members ──

func (d *DB) CreateMember(m *Member) error {
	m.ID = genID()
	m.CreatedAt = now()
	_, err := d.db.Exec(`INSERT INTO members (id,name,email,department,created_at) VALUES (?,?,?,?,?)`,
		m.ID, m.Name, m.Email, m.Department, m.CreatedAt)
	return err
}

func (d *DB) hydrateMember(m *Member) {
	d.db.QueryRow(`SELECT COUNT(*) FROM acknowledgments WHERE member_id=?`, m.ID).Scan(&m.AckCount)
	// Pending: active policies requiring ack where member hasn't acked current version
	d.db.QueryRow(`SELECT COUNT(*) FROM policies WHERE status='active' AND requires_ack=1 AND id NOT IN (SELECT policy_id FROM acknowledgments WHERE member_id=? AND policy_version=policies.version)`, m.ID).Scan(&m.PendingAck)
}

func (d *DB) GetMember(id string) *Member {
	var m Member
	if err := d.db.QueryRow(`SELECT id,name,email,department,created_at FROM members WHERE id=?`, id).Scan(
		&m.ID, &m.Name, &m.Email, &m.Department, &m.CreatedAt); err != nil {
		return nil
	}
	d.hydrateMember(&m)
	return &m
}

func (d *DB) ListMembers() []Member {
	rows, err := d.db.Query(`SELECT id,name,email,department,created_at FROM members ORDER BY name`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []Member
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.ID, &m.Name, &m.Email, &m.Department, &m.CreatedAt); err != nil {
			continue
		}
		d.hydrateMember(&m)
		out = append(out, m)
	}
	return out
}

func (d *DB) DeleteMember(id string) error {
	d.db.Exec(`DELETE FROM acknowledgments WHERE member_id=?`, id)
	_, err := d.db.Exec(`DELETE FROM members WHERE id=?`, id)
	return err
}

// ── Acknowledgments ──

func (d *DB) Acknowledge(policyID, memberID string) error {
	p := d.GetPolicy(policyID)
	if p == nil {
		return fmt.Errorf("policy not found")
	}
	_, err := d.db.Exec(`INSERT OR REPLACE INTO acknowledgments (id,policy_id,policy_version,member_id,acked_at) VALUES (?,?,?,?,?)`,
		genID(), policyID, p.Version, memberID, now())
	return err
}

func (d *DB) ListAcknowledgments(policyID string) []Acknowledgment {
	rows, err := d.db.Query(`SELECT a.id, a.policy_id, a.policy_version, a.member_id, a.acked_at, COALESCE(m.name,'') FROM acknowledgments a LEFT JOIN members m ON a.member_id=m.id WHERE a.policy_id=? ORDER BY a.acked_at DESC`, policyID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []Acknowledgment
	for rows.Next() {
		var a Acknowledgment
		if err := rows.Scan(&a.ID, &a.PolicyID, &a.PolicyVersion, &a.MemberID, &a.AckedAt, &a.MemberName); err != nil {
			continue
		}
		out = append(out, a)
	}
	return out
}

// Who hasn't acknowledged the current version of a policy
func (d *DB) PendingAcks(policyID string) []Member {
	p := d.GetPolicy(policyID)
	if p == nil {
		return nil
	}
	rows, err := d.db.Query(`SELECT id,name,email,department,created_at FROM members WHERE id NOT IN (SELECT member_id FROM acknowledgments WHERE policy_id=? AND policy_version=?) ORDER BY name`, policyID, p.Version)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []Member
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.ID, &m.Name, &m.Email, &m.Department, &m.CreatedAt); err != nil {
			continue
		}
		out = append(out, m)
	}
	return out
}

// ── Evidence ──

func (d *DB) CreateEvidence(e *Evidence) error {
	e.ID = genID()
	e.CreatedAt = now()
	if e.CollectedAt == "" {
		e.CollectedAt = now()
	}
	_, err := d.db.Exec(`INSERT INTO evidence (id,policy_id,title,description,file_name,collected_at,collected_by,created_at) VALUES (?,?,?,?,?,?,?,?)`,
		e.ID, e.PolicyID, e.Title, e.Description, e.FileName, e.CollectedAt, e.CollectedBy, e.CreatedAt)
	return err
}

func (d *DB) ListEvidence(policyID string) []Evidence {
	rows, err := d.db.Query(`SELECT id,policy_id,title,description,file_name,collected_at,collected_by,created_at FROM evidence WHERE policy_id=? ORDER BY collected_at DESC`, policyID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []Evidence
	for rows.Next() {
		var e Evidence
		if err := rows.Scan(&e.ID, &e.PolicyID, &e.Title, &e.Description, &e.FileName, &e.CollectedAt, &e.CollectedBy, &e.CreatedAt); err != nil {
			continue
		}
		out = append(out, e)
	}
	return out
}

func (d *DB) DeleteEvidence(id string) error {
	_, err := d.db.Exec(`DELETE FROM evidence WHERE id=?`, id)
	return err
}

// ── Categories ──

func (d *DB) Categories() []string {
	rows, err := d.db.Query(`SELECT DISTINCT category FROM policies WHERE category != '' ORDER BY category`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var c string
		rows.Scan(&c)
		out = append(out, c)
	}
	return out
}

// ── Stats ──

type Stats struct {
	Policies       int     `json:"policies"`
	Active         int     `json:"active"`
	Members        int     `json:"members"`
	Acknowledgments int    `json:"acknowledgments"`
	Evidence       int     `json:"evidence"`
	OverallCompliance float64 `json:"overall_compliance"`
}

func (d *DB) Stats() Stats {
	var s Stats
	d.db.QueryRow(`SELECT COUNT(*) FROM policies`).Scan(&s.Policies)
	d.db.QueryRow(`SELECT COUNT(*) FROM policies WHERE status='active'`).Scan(&s.Active)
	d.db.QueryRow(`SELECT COUNT(*) FROM members`).Scan(&s.Members)
	d.db.QueryRow(`SELECT COUNT(*) FROM acknowledgments`).Scan(&s.Acknowledgments)
	d.db.QueryRow(`SELECT COUNT(*) FROM evidence`).Scan(&s.Evidence)

	// Overall compliance: avg compliance across active policies that require ack
	policies := d.ListPolicies("active", "")
	if len(policies) > 0 {
		total := 0.0
		count := 0
		for _, p := range policies {
			if p.RequiresAck {
				total += p.CompliancePct
				count++
			}
		}
		if count > 0 {
			s.OverallCompliance = total / float64(count)
		}
	}
	return s
}
