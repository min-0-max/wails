//go:build darwin && !ios && !server

package application

import (
	"slices"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application/internal/mainthreadharness"
)

var openDialogFilterPaths = []string{"/tmp/archive.tar.gz", "/tmp/picture.png", "/tmp/photo.jpg", "/tmp/notes.txt"}

// A filter whose extensions all have content types limits the panel to those
// content types, and the delegate enables the files with those extensions.
func TestOpenDialogFilterWithContentTypes(t *testing.T) {
	if !appKitReady {
		t.Fatal("AppKit is not available")
	}
	got := mainthreadharness.ProbeOpenPanelFilter([]string{"png", "jpg"}, openDialogFilterPaths)
	if want := []string{"public.png", "public.jpeg"}; !slices.Equal(got.ContentTypes, want) {
		t.Errorf("the panel allows the content types %q, want %q", got.ContentTypes, want)
	}
	if want := []bool{false, true, true, false}; !slices.Equal(got.Enabled, want) {
		t.Errorf("the delegate enables %v of %q, want %v", got.Enabled, openDialogFilterPaths, want)
	}
}

// A filter with an extension that has no content type, such as tar.gz, leaves
// the panel without a type restriction and lets the delegate filter by the
// suffix. A restriction to file types would replace the content types of the
// other extensions and hide the files that the filter asks for.
func TestOpenDialogFilterWithAnExtensionWithoutContentType(t *testing.T) {
	if !appKitReady {
		t.Fatal("AppKit is not available")
	}
	got := mainthreadharness.ProbeOpenPanelFilter([]string{"tar.gz", "png"}, openDialogFilterPaths)
	if got.Restricted || len(got.ContentTypes) != 0 {
		t.Errorf("the panel allows the content types %q and restricts file types %v, want no restriction", got.ContentTypes, got.Restricted)
	}
	if want := []bool{true, true, false, false}; !slices.Equal(got.Enabled, want) {
		t.Errorf("the delegate enables %v of %q, want %v", got.Enabled, openDialogFilterPaths, want)
	}
}
