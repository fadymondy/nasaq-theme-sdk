package product

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ThemePath is where a product receives published Studio themes.
const ThemePath = "/api/nasaq/theme"

// ErrBadToken is returned by a Verify func for a missing, unknown or revoked link token.
var ErrBadToken = errors.New("product: bad link token")

// Applied is what a product reports after applying a package.
type Applied struct {
	Theme   string `json:"theme"`
	Version string `json:"version"`
	// Draft is true when the product staged the theme instead of publishing it.
	Draft bool `json:"draft,omitempty"`
}

// Receiver serves PUT ThemePath. The product supplies Verify (link token -> its tenant) and
// Apply (put the package on that tenant's storefront). A link token is minted by the product
// for one store and pasted into Studio, which sends it as a bearer token.
type Receiver struct {
	Product  string
	Verify   func(ctx context.Context, r *http.Request, token string) (tenant string, err error)
	Apply    func(ctx context.Context, tenant string, p Package) (Applied, error)
	MaxBytes int64 // default 256 KiB
}

func (rc Receiver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.Header().Set("Allow", http.MethodPut)
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed", "الطريقة غير مسموحة")
		return
	}
	tok := bearer(r)
	if tok == "" {
		writeErr(w, http.StatusUnauthorized, "missing link token", "رمز الربط مفقود")
		return
	}
	tenant, err := rc.Verify(r.Context(), r, tok)
	if err != nil || tenant == "" {
		writeErr(w, http.StatusUnauthorized, "invalid link token", "رمز الربط غير صالح")
		return
	}
	max := rc.MaxBytes
	if max <= 0 {
		max = 256 << 10
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, max))
	if err != nil {
		writeErr(w, http.StatusRequestEntityTooLarge, "theme package too large", "حزمة القالب كبيرة جدا")
		return
	}
	var p Package
	if err := json.Unmarshal(body, &p); err != nil {
		writeErr(w, http.StatusBadRequest, "theme package is not valid JSON", "حزمة القالب ليست JSON صالحا")
		return
	}
	if rc.Product != "" && p.Product != rc.Product {
		writeErr(w, http.StatusUnprocessableEntity, fmt.Sprintf("this theme is built for %q, not %q", p.Product, rc.Product), "هذا القالب مصمم لمنتج آخر")
		return
	}
	if err := p.Validate(); err != nil {
		writeErr(w, http.StatusUnprocessableEntity, err.Error(), "حزمة القالب غير صالحة")
		return
	}
	res, err := rc.Apply(r.Context(), tenant, p)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could not apply the theme", "تعذر تطبيق القالب")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"applied": true, "theme": res.Theme, "version": res.Version, "draft": res.Draft})
}

func bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if len(h) > 7 && strings.EqualFold(h[:7], "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}

func writeErr(w http.ResponseWriter, code int, en, ar string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": en, "errorAr": ar})
}

// HashToken is how a product stores a link token (never the token itself).
func HashToken(tok string) string {
	s := sha256.Sum256([]byte(tok))
	return hex.EncodeToString(s[:])
}

// TokenMatches compares a token with a stored hash in constant time.
func TokenMatches(tok, hash string) bool {
	return subtle.ConstantTimeCompare([]byte(HashToken(tok)), []byte(hash)) == 1
}

// Push sends a package to a product (Studio side). base is the product's storefront or API
// origin; a 404/405 means the product does not take Studio themes (ErrNotLinked).
func Push(ctx context.Context, c *http.Client, base, token string, p Package) (Applied, error) {
	if c == nil {
		c = &http.Client{Timeout: 15 * time.Second}
	}
	b, err := json.Marshal(p)
	if err != nil {
		return Applied{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, strings.TrimRight(base, "/")+ThemePath, bytes.NewReader(b))
	if err != nil {
		return Applied{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := c.Do(req)
	if err != nil {
		return Applied{}, err
	}
	defer res.Body.Close()
	rb, _ := io.ReadAll(io.LimitReader(res.Body, 64<<10))
	switch {
	case res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusMethodNotAllowed:
		return Applied{}, ErrNotLinked
	case res.StatusCode >= 300:
		var e struct{ Error string }
		_ = json.Unmarshal(rb, &e)
		if e.Error == "" {
			e.Error = res.Status
		}
		return Applied{}, &PushError{Status: res.StatusCode, Msg: e.Error}
	}
	var a Applied
	if err := json.Unmarshal(rb, &a); err != nil {
		return Applied{}, err
	}
	return a, nil
}

// ErrNotLinked means the product has no theme receiver at that address.
var ErrNotLinked = errors.New("product: the app does not receive Nasaq themes")

// PushError is a refusal from the product.
type PushError struct {
	Status int
	Msg    string
}

func (e *PushError) Error() string {
	return fmt.Sprintf("product: push refused (%d): %s", e.Status, e.Msg)
}
