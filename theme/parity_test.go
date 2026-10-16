package theme

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
)

func TestSharedParityFixture(t *testing.T) {
	data, err := os.ReadFile("../testdata/parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Contrast []struct {
			A, B  string
			Ratio float64
		}
		Theme struct {
			Input  SiteTheme
			Public PublicSiteTheme
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Contrast) == 0 {
		t.Fatal("contrast fixture must not be empty")
	}
	for _, pair := range fixture.Contrast {
		t.Run(pair.A+"_"+pair.B, func(t *testing.T) {
			for _, colors := range [][2]string{{pair.A, pair.B}, {pair.B, pair.A}} {
				if got := Contrast(colors[0], colors[1]); math.Abs(got-pair.Ratio) > 1e-10 {
					t.Fatalf("contrast = %.15g, want %.15g", got, pair.Ratio)
				}
			}
		})
	}
	if got := fixture.Theme.Input.Public(); !reflect.DeepEqual(got, fixture.Theme.Public) {
		t.Fatalf("public theme = %+v, want %+v", got, fixture.Theme.Public)
	}
}
