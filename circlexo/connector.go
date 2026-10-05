// Package circlexo keeps a site in step with the theme its owner customises and
// publishes in Nasaq UI Studio, through the owner's CircleXO organisation.
//
// The owner signs in with CircleXO once (ConnectStart / ConnectFinish); the
// site is bound to a Nasaq project (Bind), and whatever is published there is
// pulled back (Pull): on bind, when Nasaq says it changed, and on a periodic
// sweep (StartSweep). Nothing Nasaq pushes is trusted: a push only triggers a
// pull, made with the site's own CircleXO grant.
//
// The app supplies where links live (Store; SQLStore covers Postgres), how
// refresh tokens are sealed (Sealer), and how a site's look is read and saved
// (Site).
package circlexo

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	cxo "github.com/circlexo/circlexo-go"
	"github.com/circlexo/circlexo-go/oidc"
	"github.com/circlexo/circlexo-go/service"

	"github.com/fadymondy/nasaq-theme-sdk/nasaq"
	"github.com/fadymondy/nasaq-theme-sdk/theme"
)

// Link is a site's CircleXO connection, without the token.
type Link struct {
	OrgID       string     `json:"org_id"`
	OrgName     string     `json:"org_name,omitempty"`
	ProjectID   string     `json:"project_id,omitempty"`
	ProjectName string     `json:"project_name,omitempty"`
	LinkedBy    string     `json:"linked_by,omitempty"`
	LinkedAt    time.Time  `json:"linked_at"`
	LastPullAt  *time.Time `json:"last_pull_at,omitempty"`
	LastError   string     `json:"last_error,omitempty"`
}

// GrantLost is Link.LastError when the hub refused the stored grant: the owner
// must connect again, and the sweep skips the site until then.
const GrantLost = "grant_lost"

// Store keeps each site's link. site is the app's own key for it (a tenant id,
// a store id). Every method is called with the app's privileged connection:
// the refresh token is a secret.
type Store interface {
	// SaveLink links site to l.OrgID with the sealed refresh token, clearing
	// LastError. Linking to a different organisation forgets the project; the
	// same one keeps it.
	SaveLink(ctx context.Context, site string, l Link, sealedRefresh string) error
	// Link is site's link, or nil when it has none.
	Link(ctx context.Context, site string) (*Link, error)
	// DeleteLink forgets the link and its token.
	DeleteLink(ctx context.Context, site string) error
	// WithRefresh calls fn with site's sealed refresh token held exclusively
	// (the hub rotates it). When fn returns a non-empty next without an error,
	// next replaces it. It returns theme.ErrNotLinked when site has no link.
	WithRefresh(ctx context.Context, site string, fn func(sealed string) (next string, err error)) error
	// SetProject records the Nasaq project site is bound to.
	SetProject(ctx context.Context, site, projectID, projectName string) error
	// SetPulled records a pull now, with its error ("" when it went well).
	SetPulled(ctx context.Context, site, lastError string) error
	// SetError records an error without a pull (GrantLost).
	SetError(ctx context.Context, site, lastError string) error
	// Bound is every site bound to a project whose grant is not lost.
	Bound(ctx context.Context) ([]string, error)
}

// Sealer seals refresh tokens at rest.
type Sealer interface {
	Configured() bool
	Seal(plaintext string) (string, error)
	Unseal(sealed string) (string, error)
}

// Site is the app's side: a site's current look, and saving a new one. SetLook
// may refuse (a theme.Error, say) when the look would break the site.
type Site interface {
	Look(ctx context.Context, site string) (theme.Look, error)
	SetLook(ctx context.Context, site string, l theme.Look) error
}

// Connector connects sites to Nasaq through CircleXO.
type Connector struct {
	// Hub is the app's CircleXO client config, read per call (an operator may
	// set it at run time). An error or an invalid config is ErrNotConfigured.
	Hub func(context.Context) (cxo.Config, error)
	// NasaqAudience is Nasaq's app id on the hub: the audience the owner's
	// token is exchanged for.
	NasaqAudience func(context.Context) string
	Nasaq         *nasaq.Client
	Store         Store
	Vault         Sealer
	Site          Site
	// HTTP is used for the hub; nil means the config's own client.
	HTTP *http.Client
	// Now is the clock (tests); nil means time.Now.
	Now func() time.Time
}

func (c *Connector) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

// Configured reports whether a site can connect CircleXO.
func (c *Connector) Configured(ctx context.Context) bool {
	_, err := c.hub(ctx)
	return err == nil
}

func (c *Connector) hub(ctx context.Context) (cxo.Config, error) {
	if c.Hub == nil {
		return cxo.Config{}, theme.ErrNotConfigured
	}
	cfg, err := c.Hub(ctx)
	if err != nil {
		return cxo.Config{}, theme.ErrNotConfigured
	}
	if c.Vault == nil || !c.Vault.Configured() {
		return cxo.Config{}, theme.ErrNoVault
	}
	if c.HTTP != nil {
		cfg.HTTPClient = c.HTTP
	}
	if err := cfg.Validate(); err != nil {
		return cxo.Config{}, theme.ErrNotConfigured
	}
	return cfg, nil
}

// connectState travels through the hub sealed: which site, which account, and
// the PKCE verifier and nonce for the code.
type connectState struct {
	Site     string `json:"t"`
	User     string `json:"u"`
	Nonce    string `json:"n"`
	Verifier string `json:"v"`
	Redirect string `json:"r"`
	Exp      int64  `json:"e"`
}

func random(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// ConnectTTL is how long a sign-in may take.
const ConnectTTL = 15 * time.Minute

// ConnectStart is the CircleXO sign-in URL for user, an admin of site. redirect
// is the page the hub returns to (registered on the app's hub client); the
// caller checks it is one of the app's own.
func (c *Connector) ConnectStart(ctx context.Context, site, user, redirect string) (string, error) {
	cfg, err := c.hub(ctx)
	if err != nil {
		return "", err
	}
	st := connectState{Site: site, User: user, Nonce: random(24), Verifier: random(48), Redirect: redirect,
		Exp: c.now().Add(ConnectTTL).Unix()}
	raw, _ := json.Marshal(st)
	sealed, err := c.Vault.Seal(string(raw))
	if err != nil {
		return "", theme.ErrNoVault
	}
	// The RP's own state cookie is not used: the state is sealed above, so the
	// key only satisfies the constructor.
	rp, err := oidc.New(cfg, redirect, []byte(random(24))[:32])
	if err != nil {
		return "", theme.ErrNotConfigured
	}
	// URL-safe, so it survives a hub that does not escape it on the way back.
	u, err := rp.AuthURL(ctx, base64.RawURLEncoding.EncodeToString([]byte(sealed)), st.Nonce, st.Verifier, "", "")
	if err != nil {
		return "", fmt.Errorf("%w: %v", theme.ErrUpstream, err)
	}
	return u, nil
}

// ConnectFinish trades the hub's code for the owner's tokens and links the site
// to their organisation. The sign-in must have been started by the same
// account, for the same site.
func (c *Connector) ConnectFinish(ctx context.Context, site, user, code, state string) (Link, error) {
	cfg, err := c.hub(ctx)
	if err != nil {
		return Link{}, err
	}
	var st connectState
	sealed, err := base64.RawURLEncoding.DecodeString(state)
	raw := ""
	if err == nil {
		raw, err = c.Vault.Unseal(string(sealed))
	}
	if err != nil || json.Unmarshal([]byte(raw), &st) != nil || c.now().Unix() > st.Exp || st.User == "" || st.User != user || st.Site != site {
		return Link{}, theme.ErrState
	}
	svc, err := service.New(cfg)
	if err != nil {
		return Link{}, theme.ErrNotConfigured
	}
	tok, err := svc.Post(ctx, url.Values{"grant_type": {"authorization_code"}, "code": {code},
		"redirect_uri": {st.Redirect}, "code_verifier": {st.Verifier}})
	if err != nil {
		var oe *service.OAuthError
		if errors.As(err, &oe) && oe.Code == "invalid_grant" {
			return Link{}, fmt.Errorf("%w (%v)", theme.ErrState, err)
		}
		return Link{}, fmt.Errorf("%w: %v", theme.ErrUpstream, err)
	}
	v := cxo.NewVerifier(cfg)
	idc, err := v.VerifyIDToken(ctx, tok.IDToken, cfg.ClientID, st.Nonce)
	if err != nil {
		return Link{}, fmt.Errorf("%w (id token: %v)", theme.ErrState, err)
	}
	atc, err := v.VerifyAccessToken(ctx, tok.AccessToken)
	if err != nil {
		return Link{}, fmt.Errorf("%w (access token: %v)", theme.ErrState, err)
	}
	if atc.Subject != idc.Subject {
		return Link{}, fmt.Errorf("%w (the tokens name different accounts)", theme.ErrState)
	}
	if atc.OrgID == "" {
		return Link{}, theme.ErrNoOrg
	}
	if tok.RefreshToken == "" {
		return Link{}, fmt.Errorf("%w: the hub issued no refresh token", theme.ErrUpstream)
	}
	sealedRT, err := c.Vault.Seal(tok.RefreshToken)
	if err != nil {
		return Link{}, theme.ErrNoVault
	}
	link := Link{OrgID: atc.OrgID, OrgName: orgName(ctx, cfg, svc, atc.OrgID), LinkedBy: user, LinkedAt: c.now()}
	if err := c.Store.SaveLink(ctx, site, link, sealedRT); err != nil {
		return Link{}, fmt.Errorf("circlexo: save link: %w", err)
	}
	l, err := c.Store.Link(ctx, site)
	if err != nil {
		return Link{}, err
	}
	if l == nil {
		return link, nil
	}
	return *l, nil
}

// orgName is the organisation's display name, best effort.
func orgName(ctx context.Context, cfg cxo.Config, svc *service.Client, org string) string {
	o, err := cxo.NewClient(cfg, service.NewSource(svc)).Org(ctx, org)
	if err != nil {
		return ""
	}
	return o.Name
}

// Link is site's link, or nil when it has none.
func (c *Connector) Link(ctx context.Context, site string) (*Link, error) {
	return c.Store.Link(ctx, site)
}

// Disconnect forgets the link and the token. The theme stays as it is.
func (c *Connector) Disconnect(ctx context.Context, site string) error {
	return c.Store.DeleteLink(ctx, site)
}

var errGrantLost = errors.New("grant lost")

// NasaqToken refreshes site's grant and exchanges it for a token Nasaq accepts.
func (c *Connector) NasaqToken(ctx context.Context, site string) (string, error) {
	cfg, err := c.hub(ctx)
	if err != nil {
		return "", err
	}
	svc, err := service.New(cfg)
	if err != nil {
		return "", theme.ErrNotConfigured
	}
	var access string
	err = c.Store.WithRefresh(ctx, site, func(sealed string) (string, error) {
		rt, err := c.Vault.Unseal(sealed)
		if err != nil || rt == "" {
			return "", theme.ErrGrantLost
		}
		tok, err := svc.Post(ctx, url.Values{"grant_type": {"refresh_token"}, "refresh_token": {rt}})
		if err != nil {
			var oe *service.OAuthError
			if errors.As(err, &oe) && oe.Code == "invalid_grant" {
				return "", errGrantLost
			}
			return "", fmt.Errorf("%w: refresh: %v", theme.ErrUpstream, err)
		}
		access = tok.AccessToken
		if tok.RefreshToken == "" || tok.RefreshToken == rt {
			return "", nil
		}
		next, err := c.Vault.Seal(tok.RefreshToken)
		if err != nil {
			return "", theme.ErrNoVault
		}
		return next, nil
	})
	if errors.Is(err, errGrantLost) {
		_ = c.Store.SetError(ctx, site, GrantLost)
		return "", theme.ErrGrantLost
	}
	if err != nil {
		return "", err
	}
	aud := ""
	if c.NasaqAudience != nil {
		aud = c.NasaqAudience(ctx)
	}
	ex, err := svc.Exchange(ctx, access, service.AccessTokenType, aud)
	if err != nil {
		return "", fmt.Errorf("%w: exchange: %v", theme.ErrUpstream, err)
	}
	return ex.AccessToken, nil
}

// Projects lists the Nasaq projects in site's organisation.
func (c *Connector) Projects(ctx context.Context, site string) ([]nasaq.Project, error) {
	tok, err := c.NasaqToken(ctx, site)
	if err != nil {
		return nil, err
	}
	ps, err := c.Nasaq.Projects(ctx, tok)
	if err != nil {
		return nil, nasaq.Refusal(err)
	}
	return ps, nil
}

// Bind makes projectID site's Nasaq project: Nasaq learns where the app renders
// the site (b; its App defaults to the Nasaq client's), and the published theme
// is pulled at once.
func (c *Connector) Bind(ctx context.Context, site, projectID string, b nasaq.Binding) (*Link, error) {
	tok, err := c.NasaqToken(ctx, site)
	if err != nil {
		return nil, err
	}
	ps, err := c.Nasaq.Projects(ctx, tok)
	if err != nil {
		return nil, nasaq.Refusal(err)
	}
	var proj *nasaq.Project
	for i := range ps {
		if ps[i].ID == projectID {
			proj = &ps[i]
		}
	}
	if proj == nil {
		return nil, theme.ErrUnknownProject
	}
	if err := c.Nasaq.Bind(ctx, tok, projectID, b); err != nil {
		return nil, nasaq.Refusal(err)
	}
	if err := c.Store.SetProject(ctx, site, projectID, proj.Label()); err != nil {
		return nil, err
	}
	if _, err := c.Pull(ctx, site); err != nil && !NothingPublished(err) {
		return nil, err
	}
	return c.Store.Link(ctx, site)
}

var errNothingPublished = errors.New("circlexo: the Nasaq project has no published theme")

// NothingPublished reports a pull that found no published theme on the project.
func NothingPublished(err error) bool { return errors.Is(err, errNothingPublished) }

// Pull reads the theme published on site's Nasaq project and applies it.
// changed is false when the look is already that; a project with nothing
// published yet changes nothing (NothingPublished).
func (c *Connector) Pull(ctx context.Context, site string) (changed bool, err error) {
	defer func() {
		msg := ""
		switch {
		case errors.Is(err, theme.ErrGrantLost):
			msg = GrantLost // the sweep skips it until the owner connects again
		case err != nil && !NothingPublished(err):
			msg = err.Error()
			if len(msg) > 300 {
				msg = msg[:300]
			}
		}
		_ = c.Store.SetPulled(ctx, site, msg)
	}()
	l, err := c.Store.Link(ctx, site)
	if err != nil {
		return false, err
	}
	if l == nil {
		return false, theme.ErrNotLinked
	}
	if l.ProjectID == "" {
		return false, theme.ErrNoProject
	}
	tok, err := c.NasaqToken(ctx, site)
	if err != nil {
		return false, err
	}
	core, err := c.Nasaq.Core(ctx, tok, l.ProjectID)
	if err != nil {
		return false, nasaq.Refusal(err)
	}
	if core == nil {
		return false, errNothingPublished
	}
	cur, err := c.Site.Look(ctx, site)
	if err != nil {
		return false, err
	}
	next, err := theme.Map(cur, theme.Source{ID: core.Theme, Version: core.Version, ProjectID: l.ProjectID}, core.Tokens, c.now())
	if err != nil {
		return false, err
	}
	if theme.Same(cur, next) {
		return false, nil
	}
	return true, c.Site.SetLook(ctx, site, next)
}

// ---- Nasaq's "it changed" signal and the sweep -----------------------------------

// PullGap is the least time between two pulls a push can trigger for one site.
const PullGap = 20 * time.Second

// Limiter rate-limits push-triggered pulls per site. The zero value is ready.
type Limiter struct {
	// Gap is the least time between two pulls; 0 means PullGap.
	Gap time.Duration

	mu   sync.Mutex
	last map[string]time.Time
}

// Allow reports whether site may pull now, and if so counts it.
func (l *Limiter) Allow(site string, now time.Time) bool {
	gap := l.Gap
	if gap == 0 {
		gap = PullGap
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if t, ok := l.last[site]; ok && now.Sub(t) < gap {
		return false
	}
	if l.last == nil {
		l.last = map[string]time.Time{}
	}
	l.last[site] = now
	return true
}

// StartSweep pulls every bound site's theme every so often, so a change
// published in Nasaq arrives even when its push did not. changed is called for
// each site whose look changed (the app drops its cache). It stops with ctx.
func (c *Connector) StartSweep(ctx context.Context, log *slog.Logger, every time.Duration, changed func(site string)) {
	go func() {
		t := time.NewTicker(every)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
			c.Sweep(ctx, log, changed)
		}
	}()
}

// Sweep pulls every bound site's theme once.
func (c *Connector) Sweep(ctx context.Context, log *slog.Logger, changed func(site string)) {
	if !c.Configured(ctx) {
		return
	}
	ids, err := c.Store.Bound(ctx)
	if err != nil && log != nil {
		log.Warn("circlexo: sweep", "err", err)
	}
	for _, id := range ids {
		ch, err := c.Pull(ctx, id)
		switch {
		case err != nil && !NothingPublished(err):
			if log != nil {
				log.Warn("circlexo: pull", "site", id, "err", err)
			}
		case ch:
			if changed != nil {
				changed(id)
			}
			if log != nil {
				log.Info("circlexo: pulled a new theme", "site", id)
			}
		}
	}
}
