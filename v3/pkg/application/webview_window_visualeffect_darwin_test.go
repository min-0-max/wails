//go:build darwin && !ios && !server

package application

import (
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application/internal/mainthreadharness"
)

// The visual effect view that stands in for the liquid glass takes the
// material and blending of its style, or the material it is given, follows
// the active state of its window, is not emphasized, and keeps the appearance
// of its window.
func TestLiquidGlassVisualEffectView(t *testing.T) {
	if !appKitReady {
		t.Fatal("AppKit is not available")
	}
	osMajor := mainthreadharness.ProbeVisualEffect(0, -1).OSMajor
	light, dark := mainthreadharness.MaterialHUDWindow, mainthreadharness.MaterialFullScreenUI
	if osMajor >= 15 {
		light, dark = mainthreadharness.MaterialUnderPageBackground, mainthreadharness.MaterialHeaderView
	}
	tests := []struct {
		name         string
		style        int
		material     int
		wantMaterial int
		wantBlending int
	}{
		{"automatic", 0, -1, mainthreadharness.MaterialContentBackground, mainthreadharness.BlendingWithinWindow},
		{"light", 1, -1, light, mainthreadharness.BlendingBehindWindow},
		{"dark", 2, -1, dark, mainthreadharness.BlendingBehindWindow},
		{"vibrant", 3, -1, mainthreadharness.MaterialHUDWindow, mainthreadharness.BlendingBehindWindow},
		{"given material", 2, mainthreadharness.MaterialSidebar, mainthreadharness.MaterialSidebar, mainthreadharness.BlendingBehindWindow},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := mainthreadharness.ProbeVisualEffect(test.style, test.material)
			if got.Material != test.wantMaterial || got.BlendingMode != test.wantBlending {
				t.Errorf("material %d with blending %d, want %d with %d", got.Material, got.BlendingMode, test.wantMaterial, test.wantBlending)
			}
			if got.State != mainthreadharness.StateFollowsWindowActiveState || got.Emphasized || got.HasAppearance {
				t.Errorf("state %d, emphasized %v, own appearance %v; want the window's active state, not emphasized, no appearance of its own", got.State, got.Emphasized, got.HasAppearance)
			}
		})
	}
}
