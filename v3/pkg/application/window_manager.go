package application

import "sync"

type windowLifetime struct {
	id uint64

	mu      sync.Mutex
	retired bool
}

func (l *windowLifetime) retire() {
	if l == nil {
		return
	}
	l.mu.Lock()
	l.retired = true
	l.mu.Unlock()
}

func (l *windowLifetime) admit() bool {
	if l == nil {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return !l.retired
}

var windowLifetimeSequence uint64
var windowLifetimeSequenceLock sync.Mutex

func nextWindowLifetimeID() uint64 {
	windowLifetimeSequenceLock.Lock()
	defer windowLifetimeSequenceLock.Unlock()
	windowLifetimeSequence++
	return uint64(1)<<63 | windowLifetimeSequence
}

func allocatedWindowLifetimeID(id uint64) bool {
	const prefix = uint64(1) << 63
	if id&prefix == 0 {
		return false
	}
	sequence := id &^ prefix
	windowLifetimeSequenceLock.Lock()
	defer windowLifetimeSequenceLock.Unlock()
	return sequence > 0 && sequence <= windowLifetimeSequence
}

// WindowManager manages all window-related operations
type WindowManager struct {
	app *App
}

// newWindowManager creates a new WindowManager instance
func newWindowManager(app *App) *WindowManager {
	return &WindowManager{
		app: app,
		// Use the app's existing windows map - don't create a new one
	}
}

// GetByName returns a window by name and whether it exists
func (wm *WindowManager) GetByName(name string) (Window, bool) {
	wm.app.windowsLock.RLock()
	defer wm.app.windowsLock.RUnlock()

	for _, window := range wm.app.windows {
		if window.Name() == name {
			return window, true
		}
	}
	return nil, false
}

// Get is an alias for GetByName for consistency
func (wm *WindowManager) Get(name string) (Window, bool) {
	return wm.GetByName(name)
}

// GetByID returns a window by ID and whether it exists
func (wm *WindowManager) GetByID(id uint) (Window, bool) {
	wm.app.windowsLock.RLock()
	defer wm.app.windowsLock.RUnlock()

	window, exists := wm.app.windows[id]
	return window, exists
}

// OnCreate registers a callback to be called when a window is created
func (wm *WindowManager) OnCreate(callback func(Window)) {
	wm.app.windowCreatedCallbacks = append(wm.app.windowCreatedCallbacks, callback)
}

// New creates a new webview window
func (wm *WindowManager) New() *WebviewWindow {
	return wm.NewWithOptions(WebviewWindowOptions{})
}

// NewWithOptions creates a new webview window with options
func (wm *WindowManager) NewWithOptions(windowOptions WebviewWindowOptions) *WebviewWindow {
	newWindow := NewWindow(windowOptions)
	wm.addAndRun(newWindow)
	return newWindow
}

func (wm *WindowManager) addAndRun(newWindow *WebviewWindow) {
	lifetimeID := nextWindowLifetimeID()
	newWindow.callbackLifetimeID = lifetimeID
	wm.app.windowsLock.Lock()
	if wm.app.windowLifetimes == nil {
		wm.app.windowLifetimes = make(map[uint]*windowLifetime)
	}
	wm.app.windows[newWindow.ID()] = newWindow
	wm.app.windowLifetimes[newWindow.ID()] = &windowLifetime{id: lifetimeID}
	wm.app.windowsLock.Unlock()

	for _, hook := range wm.app.windowCreatedCallbacks {
		hook(newWindow)
	}

	wm.app.runOrDeferToAppRun(newWindow)
}

// Current returns the current active window (may be nil)
func (wm *WindowManager) Current() Window {
	if wm.app.impl == nil {
		return nil
	}
	id := wm.app.impl.getCurrentWindowID()
	wm.app.windowsLock.RLock()
	defer wm.app.windowsLock.RUnlock()
	return wm.app.windows[id]
}

// Add adds a window to the manager
func (wm *WindowManager) Add(window Window) {
	lifetimeID := nextWindowLifetimeID()
	if webviewWindow, ok := window.(*WebviewWindow); ok {
		webviewWindow.callbackLifetimeID = lifetimeID
	}
	wm.addWindow(window, lifetimeID)
}

func (wm *WindowManager) addWindow(window Window, lifetimeID uint64) {
	wm.app.windowsLock.Lock()
	defer wm.app.windowsLock.Unlock()
	if wm.app.windowLifetimes == nil {
		wm.app.windowLifetimes = make(map[uint]*windowLifetime)
	}
	wm.app.windows[window.ID()] = window
	wm.app.windowLifetimes[window.ID()] = &windowLifetime{id: lifetimeID}

	// Call registered callbacks
	for _, callback := range wm.app.windowCreatedCallbacks {
		callback(window)
	}
}

func (wm *WindowManager) lifetimeID(windowID uint) uint64 {
	wm.app.windowsLock.RLock()
	defer wm.app.windowsLock.RUnlock()
	if lifetime := wm.app.windowLifetimes[windowID]; lifetime != nil {
		return lifetime.id
	}
	return 0
}

// Retire closes callback admission for one window lifetime before native
// teardown begins. Remove remains a manager-only operation.
func (wm *WindowManager) Retire(windowID uint, lifetimeID uint64) {
	wm.app.windowsLock.RLock()
	lifetime := wm.app.windowLifetimes[windowID]
	wm.app.windowsLock.RUnlock()
	if lifetime != nil && lifetime.id == lifetimeID {
		lifetime.retire()
	}
}

// Remove removes a window from the manager by ID
func (wm *WindowManager) Remove(windowID uint) {
	wm.app.windowsLock.Lock()
	defer wm.app.windowsLock.Unlock()
	delete(wm.app.windows, windowID)
	delete(wm.app.windowLifetimes, windowID)
}

func (wm *WindowManager) removeLifetime(windowID uint, lifetimeID uint64) {
	wm.app.windowsLock.Lock()
	defer wm.app.windowsLock.Unlock()
	lifetime := wm.app.windowLifetimes[windowID]
	if lifetime == nil || lifetime.id != lifetimeID {
		return
	}
	delete(wm.app.windows, windowID)
	delete(wm.app.windowLifetimes, windowID)
}

// RemoveByName removes a window from the manager by name
func (wm *WindowManager) RemoveByName(name string) bool {
	window, exists := wm.GetByName(name)
	if exists {
		wm.Remove(window.ID())
		return true
	}
	return false
}

// Internal methods for backward compatibility
func (wm *WindowManager) add(window Window) {
	wm.Add(window)
}

func (wm *WindowManager) remove(windowID uint) {
	wm.Remove(windowID)
}

// GetAll returns all windows
func (wm *WindowManager) GetAll() []Window {
	wm.app.windowsLock.RLock()
	defer wm.app.windowsLock.RUnlock()

	windows := make([]Window, 0, len(wm.app.windows))
	for _, window := range wm.app.windows {
		windows = append(windows, window)
	}
	return windows
}
