//go:build darwin && !ios && !server

package application

import (
	"slices"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application/internal/mainthreadharness"
)

// The file drop view of a window registers file URLs, the pasteboard type
// that dragged files arrive as.
func TestFileDropRegistersFileURLs(t *testing.T) {
	if !appKitReady {
		t.Fatal("AppKit is not available")
	}
	if got, want := mainthreadharness.FileDropTypes(), []string{"public.file-url"}; !slices.Equal(got, want) {
		t.Errorf("the file drop view registers %q, want %q", got, want)
	}
}

// The drop handlers read the path of every dropped file, in order, and see no
// files on a pasteboard that holds only web URLs.
func TestFileDropReadsTheDroppedFiles(t *testing.T) {
	if !appKitReady {
		t.Fatal("AppKit is not available")
	}
	tests := []struct {
		name      string
		urls      []string
		wantFiles bool
		wantPaths []string
	}{
		{
			name:      "files",
			urls:      []string{"file:///tmp/a%20b.txt", "file:///tmp/c.png", "file:///tmp/folder/"},
			wantFiles: true,
			wantPaths: []string{"/tmp/a b.txt", "/tmp/c.png", "/tmp/folder"},
		},
		{
			name:      "a file and a web URL",
			urls:      []string{"https://example.com/page", "file:///tmp/c.png"},
			wantFiles: true,
			wantPaths: []string{"/tmp/c.png"},
		},
		{
			name: "web URLs",
			urls: []string{"https://example.com/page"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hasFiles, paths := mainthreadharness.PasteboardFiles(test.urls)
			if hasFiles != test.wantFiles {
				t.Errorf("the pasteboard holds files = %v, want %v", hasFiles, test.wantFiles)
			}
			if !slices.Equal(paths, test.wantPaths) {
				t.Errorf("the drop handlers read %q, want %q", paths, test.wantPaths)
			}
		})
	}
}
