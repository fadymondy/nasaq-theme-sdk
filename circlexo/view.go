package circlexo

import (
	"context"

	"github.com/fadymondy/nasaq-theme-sdk/theme"
)

// ThemeStatus is what an app's Themes page reads: the site's theme and its
// CircleXO link. The names match the npm package's types.
type ThemeStatus struct {
	Theme  theme.SiteTheme `json:"theme"`
	Action theme.ColorPair `json:"action"`
	// DefaultTheme is the platform's default Nasaq theme; empty means "reset"
	// goes back to the app's own look.
	DefaultTheme string        `json:"default_theme,omitempty"`
	CircleXO     ThemeCircleXO `json:"circlexo"`
}

// ThemeCircleXO is the CircleXO side of ThemeStatus.
type ThemeCircleXO struct {
	// Configured is false until the operator sets the app up on the hub.
	Configured bool  `json:"configured"`
	Link       *Link `json:"link,omitempty"`
}

// ThemeConnectView answers the connect call.
type ThemeConnectView struct {
	// AuthURL is set by action=start: send the browser there.
	AuthURL string       `json:"auth_url,omitempty"`
	Status  *ThemeStatus `json:"status,omitempty"`
}

// Status builds site's ThemeStatus from its look and its link.
func (c *Connector) Status(ctx context.Context, site, defaultTheme string) (ThemeStatus, error) {
	look, err := c.Site.Look(ctx, site)
	if err != nil {
		return ThemeStatus{}, err
	}
	out := ThemeStatus{Theme: look.Theme, Action: look.Action, DefaultTheme: defaultTheme}
	out.CircleXO.Configured = c.Configured(ctx)
	if c.Store != nil {
		if out.CircleXO.Link, err = c.Store.Link(ctx, site); err != nil {
			return ThemeStatus{}, err
		}
	}
	return out, nil
}
