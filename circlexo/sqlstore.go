package circlexo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"

	"github.com/fadymondy/nasaq-theme-sdk/theme"
)

// Schema is the Postgres table SQLStore expects, for a site keyed by uuid. Copy
// it into the app's migrations, renaming the table and key column as needed
// (and adding the foreign key to the app's sites). Keep the table out of reach
// of the request path: it holds refresh tokens.
const Schema = `
CREATE TABLE IF NOT EXISTS circlexo_links (
    site_id       uuid        PRIMARY KEY,
    org_id        text        NOT NULL,
    org_name      text        NOT NULL DEFAULT '',
    project_id    text        NOT NULL DEFAULT '',
    project_name  text        NOT NULL DEFAULT '',
    linked_by     text        NOT NULL DEFAULT '',
    refresh_token text        NOT NULL DEFAULT '',
    linked_at     timestamptz NOT NULL DEFAULT now(),
    last_pull_at  timestamptz,
    last_error    text        NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS circlexo_links_project ON circlexo_links (project_id) WHERE project_id <> '';
`

// SQLStore is a Store on a Postgres table shaped like Schema.
type SQLStore struct {
	DB *sql.DB
	// Table and Key name the table and its site column; empty means
	// "circlexo_links" and "site_id".
	Table, Key string
	// KeyType is the key's SQL type, for the casts; empty means "uuid".
	KeyType string
}

var identRE = regexp.MustCompile(`^[a-z_][a-z0-9_]{0,62}$`)

func (s *SQLStore) names() (table, key, typ string) {
	table, key, typ = s.Table, s.Key, s.KeyType
	if table == "" {
		table = "circlexo_links"
	}
	if key == "" {
		key = "site_id"
	}
	if typ == "" {
		typ = "uuid"
	}
	for _, n := range []string{table, key, typ} {
		if !identRE.MatchString(n) {
			panic(fmt.Sprintf("circlexo: SQLStore name %q is not a plain identifier", n))
		}
	}
	return table, key, typ
}

// q expands {t}, {k} and {ty} in a statement.
func (s *SQLStore) q(stmt string) string {
	t, k, ty := s.names()
	out := []byte{}
	for i := 0; i < len(stmt); i++ {
		switch {
		case stmt[i] == '{' && i+2 < len(stmt) && stmt[i+1] == 't' && stmt[i+2] == '}':
			out = append(out, t...)
			i += 2
		case stmt[i] == '{' && i+2 < len(stmt) && stmt[i+1] == 'k' && stmt[i+2] == '}':
			out = append(out, k...)
			i += 2
		case stmt[i] == '{' && i+3 < len(stmt) && stmt[i+1:i+4] == "ty}":
			out = append(out, ty...)
			i += 3
		default:
			out = append(out, stmt[i])
		}
	}
	return string(out)
}

func (s *SQLStore) SaveLink(ctx context.Context, site string, l Link, sealedRefresh string) error {
	_, err := s.DB.ExecContext(ctx, s.q(`
		INSERT INTO {t} ({k}, org_id, org_name, linked_by, refresh_token, linked_at)
		VALUES ($1::{ty}, $2, $3, $4, $5, now())
		ON CONFLICT ({k}) DO UPDATE SET
			org_id = EXCLUDED.org_id, org_name = EXCLUDED.org_name, linked_by = EXCLUDED.linked_by,
			refresh_token = EXCLUDED.refresh_token, linked_at = now(), last_error = '',
			project_id   = CASE WHEN {t}.org_id = EXCLUDED.org_id THEN {t}.project_id ELSE '' END,
			project_name = CASE WHEN {t}.org_id = EXCLUDED.org_id THEN {t}.project_name ELSE '' END`),
		site, l.OrgID, l.OrgName, l.LinkedBy, sealedRefresh)
	return err
}

func (s *SQLStore) Link(ctx context.Context, site string) (*Link, error) {
	var l Link
	var pulled sql.NullTime
	err := s.DB.QueryRowContext(ctx, s.q(`
		SELECT org_id, org_name, project_id, project_name, linked_by, linked_at, last_pull_at, last_error
		FROM {t} WHERE {k} = $1::{ty}`), site).
		Scan(&l.OrgID, &l.OrgName, &l.ProjectID, &l.ProjectName, &l.LinkedBy, &l.LinkedAt, &pulled, &l.LastError)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("circlexo: load link: %w", err)
	}
	if pulled.Valid {
		l.LastPullAt = &pulled.Time
	}
	return &l, nil
}

func (s *SQLStore) DeleteLink(ctx context.Context, site string) error {
	_, err := s.DB.ExecContext(ctx, s.q(`DELETE FROM {t} WHERE {k} = $1::{ty}`), site)
	return err
}

func (s *SQLStore) WithRefresh(ctx context.Context, site string, fn func(sealed string) (string, error)) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // a no-op after Commit
	var sealed string
	err = tx.QueryRowContext(ctx, s.q(`SELECT refresh_token FROM {t} WHERE {k} = $1::{ty} FOR UPDATE`), site).Scan(&sealed)
	if errors.Is(err, sql.ErrNoRows) {
		return theme.ErrNotLinked
	}
	if err != nil {
		return err
	}
	next, err := fn(sealed)
	if err != nil {
		return err
	}
	if next != "" && next != sealed {
		if _, err := tx.ExecContext(ctx, s.q(`UPDATE {t} SET refresh_token = $2 WHERE {k} = $1::{ty}`), site, next); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLStore) SetProject(ctx context.Context, site, projectID, projectName string) error {
	_, err := s.DB.ExecContext(ctx, s.q(`UPDATE {t} SET project_id = $2, project_name = $3 WHERE {k} = $1::{ty}`), site, projectID, projectName)
	return err
}

func (s *SQLStore) SetPulled(ctx context.Context, site, lastError string) error {
	_, err := s.DB.ExecContext(ctx, s.q(`UPDATE {t} SET last_pull_at = now(), last_error = $2 WHERE {k} = $1::{ty}`), site, lastError)
	return err
}

func (s *SQLStore) SetError(ctx context.Context, site, lastError string) error {
	_, err := s.DB.ExecContext(ctx, s.q(`UPDATE {t} SET last_error = $2 WHERE {k} = $1::{ty}`), site, lastError)
	return err
}

func (s *SQLStore) Bound(ctx context.Context) ([]string, error) {
	rows, err := s.DB.QueryContext(ctx, s.q(`SELECT {k}::text FROM {t} WHERE project_id <> '' AND last_error <> $1`), GrantLost)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
