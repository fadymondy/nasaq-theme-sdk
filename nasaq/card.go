package nasaq

import (
	"context"

	"github.com/fadymondy/nasaq-theme-sdk/theme"
)

// ThemeCard is one catalog theme as an app's Themes page shows it. The name
// matches the npm package's ThemeCard, so an OpenAPI export keeps the schema.
type ThemeCard struct {
	ID            string     `json:"id"`
	Version       string     `json:"version"`
	Name          theme.Text `json:"name"`
	Description   theme.Text `json:"description,omitempty"`
	Category      string     `json:"category,omitempty"`
	Free          bool       `json:"free"`
	Thumbnail     string     `json:"thumbnail,omitempty"`
	ThumbnailDark string     `json:"thumbnail_dark,omitempty"`
	DemoURL       string     `json:"demo_url,omitempty"`
	// CustomizeURL opens the theme in Nasaq's customizer, in the linked org.
	CustomizeURL string `json:"customize_url"`
}

// Cards is the app's catalog as ThemeCards, the customize links in locale and
// in org (the site's CircleXO organisation; "" when it is not linked).
func (c *Client) Cards(ctx context.Context, locale, org string) ([]ThemeCard, error) {
	list, err := c.Catalog(ctx)
	out := make([]ThemeCard, 0, len(list))
	for _, t := range list {
		out = append(out, ThemeCard{
			ID: t.ID, Version: t.Version, Name: t.Name, Description: t.Description, Category: t.Category,
			Free: t.Free, Thumbnail: t.Previews.Thumbnail, ThumbnailDark: t.Previews.ThumbnailDark,
			DemoURL: t.DemoURL, CustomizeURL: c.CustomizeURL(ctx, locale, t.ID, org),
		})
	}
	return out, err
}
