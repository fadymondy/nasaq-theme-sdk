package circlexo

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/fadymondy/nasaq-theme-sdk/theme"
)

// MemStore is a Store in memory, for tests and single-process tools. The zero
// value is ready.
type MemStore struct {
	mu    sync.Mutex
	links map[string]memLink
}

type memLink struct {
	Link
	sealed string
}

func (m *MemStore) SaveLink(_ context.Context, site string, l Link, sealedRefresh string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.links == nil {
		m.links = map[string]memLink{}
	}
	if old, ok := m.links[site]; ok && old.OrgID == l.OrgID {
		l.ProjectID, l.ProjectName = old.ProjectID, old.ProjectName
	} else {
		l.ProjectID, l.ProjectName = "", ""
	}
	l.LastError, l.LastPullAt = "", nil
	if l.LinkedAt.IsZero() {
		l.LinkedAt = time.Now()
	}
	m.links[site] = memLink{Link: l, sealed: sealedRefresh}
	return nil
}

func (m *MemStore) Link(_ context.Context, site string) (*Link, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	l, ok := m.links[site]
	if !ok {
		return nil, nil
	}
	out := l.Link
	return &out, nil
}

func (m *MemStore) DeleteLink(_ context.Context, site string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.links, site)
	return nil
}

// Sealed is site's sealed refresh token (tests).
func (m *MemStore) Sealed(site string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.links[site].sealed
}

func (m *MemStore) WithRefresh(_ context.Context, site string, fn func(string) (string, error)) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	l, ok := m.links[site]
	if !ok {
		return theme.ErrNotLinked
	}
	next, err := fn(l.sealed)
	if err != nil {
		return err
	}
	if next != "" {
		l.sealed = next
		m.links[site] = l
	}
	return nil
}

func (m *MemStore) update(site string, f func(*Link)) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if l, ok := m.links[site]; ok {
		f(&l.Link)
		m.links[site] = l
	}
	return nil
}

func (m *MemStore) SetProject(_ context.Context, site, id, name string) error {
	return m.update(site, func(l *Link) { l.ProjectID, l.ProjectName = id, name })
}

func (m *MemStore) SetPulled(_ context.Context, site, lastError string) error {
	now := time.Now()
	return m.update(site, func(l *Link) { l.LastPullAt, l.LastError = &now, lastError })
}

func (m *MemStore) SetError(_ context.Context, site, lastError string) error {
	return m.update(site, func(l *Link) { l.LastError = lastError })
}

func (m *MemStore) Bound(context.Context) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []string
	for id, l := range m.links {
		if l.ProjectID != "" && l.LastError != GrantLost {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out, nil
}
