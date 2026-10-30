//go:build darwin && !ios && !server

#include <stdbool.h>
#include <stddef.h>
#include <sys/types.h>
#include "mainthread_darwin.h"

// The native code reaches Go only through these functions. Each one runs its
// Go handler with wailsRunOnWorker, so that a callback which AppKit or WebKit
// makes on the main thread runs its Go code on the worker while the main
// thread waits in C, and an answer that the caller expects is returned
// synchronously. The Go handlers have the name of their function with the
// prefix go.

extern void goProcessApplicationEvent(unsigned int eventID, void *data);
extern void goProcessWindowEvent(unsigned int windowID, unsigned int eventID);
extern void goProcessMessage(unsigned int windowID, unsigned long long callbackLifetimeID, char *message, char *origin, bool isMainFrame);
extern void goProcessURLRequest(unsigned int windowID, void *wkUrlSchemeTask);
extern void goCancelURLRequest(void *wkUrlSchemeTask);
extern char *goAcceleratorFromKeyPress(unsigned short keyCode, unsigned long modifiers, unsigned int character, int hasCharacter);
extern void goProcessWindowKeyDownEvent(unsigned int windowID, unsigned long long callbackLifetimeID, char *acceleratorString);
extern bool goProcessWindowKeyEquivalent(unsigned int windowID, unsigned long long callbackLifetimeID, char *acceleratorString);
extern void goProcessDragItems(unsigned int windowID, unsigned long long callbackLifetimeID, char **arr, int length, int x, int y);
extern void goMacosOnDragEnter(unsigned int windowID);
extern void goMacosOnDragExit(unsigned int windowID);
extern void goMacosOnDragOver(unsigned int windowID, int x, int y);
extern void goProcessMenuItemClick(unsigned int menuID);
extern bool goShouldQuitApplication(void);
extern void goCleanup(void);
extern void goHandleOpenFile(char *filePath);
extern void goHandleOpenURL(char *url);
extern void goSystrayClickCallback(long id, int buttonID);
extern int goSystrayPreClickCallback(long id, int buttonID);
extern void goDialogCallback(int id, int buttonPressed);
extern void goOpenFileDialogCallback(unsigned int id, char *path);
extern void goOpenFileDialogCallbackEnd(unsigned int id);
extern void goSaveFileDialogCallback(unsigned int id, char *path);
extern int goResolveMediaCapturePermission(unsigned int windowID, bool needAudio, bool needVideo);
extern void goGlobalShortcutCallback(int id);
extern bool goWindowShouldUnconditionallyClose(unsigned int windowID);
extern bool goWindowIsHidden(unsigned int windowID);
extern void goHandleSecondInstanceData(char *message);

void processApplicationEvent(unsigned int eventID, void *data) {
	wailsRunOnWorker(^{
		goProcessApplicationEvent(eventID, data);
	});
}

void processWindowEvent(unsigned int windowID, unsigned int eventID) {
	wailsRunOnWorker(^{
		goProcessWindowEvent(windowID, eventID);
	});
}

void processMessage(unsigned int windowID, unsigned long long callbackLifetimeID, const char *message, const char *origin, bool isMainFrame) {
	wailsRunOnWorker(^{
		goProcessMessage(windowID, callbackLifetimeID, (char *)message, (char *)origin, isMainFrame);
	});
}

void processURLRequest(unsigned int windowID, void *wkUrlSchemeTask) {
	wailsRunOnWorker(^{
		goProcessURLRequest(windowID, wkUrlSchemeTask);
	});
}

void cancelURLRequest(void *wkUrlSchemeTask) {
	wailsRunOnWorker(^{
		goCancelURLRequest(wkUrlSchemeTask);
	});
}

char *acceleratorFromKeyPress(unsigned short keyCode, unsigned long modifiers, unsigned int character, int hasCharacter) {
	__block char *accelerator = NULL;
	wailsRunOnWorker(^{
		accelerator = goAcceleratorFromKeyPress(keyCode, modifiers, character, hasCharacter);
	});
	return accelerator;
}

void processWindowKeyDownEvent(unsigned int windowID, unsigned long long callbackLifetimeID, const char *acceleratorString) {
	wailsRunOnWorker(^{
		goProcessWindowKeyDownEvent(windowID, callbackLifetimeID, (char *)acceleratorString);
	});
}

bool processWindowKeyEquivalent(unsigned int windowID, unsigned long long callbackLifetimeID, const char *acceleratorString) {
	__block bool consumed = false;
	wailsRunOnWorker(^{
		consumed = goProcessWindowKeyEquivalent(windowID, callbackLifetimeID, (char *)acceleratorString);
	});
	return consumed;
}

void processDragItems(unsigned int windowID, unsigned long long callbackLifetimeID, char **arr, int length, int x, int y) {
	wailsRunOnWorker(^{
		goProcessDragItems(windowID, callbackLifetimeID, arr, length, x, y);
	});
}

void macosOnDragEnter(unsigned int windowID) {
	wailsRunOnWorker(^{
		goMacosOnDragEnter(windowID);
	});
}

void macosOnDragExit(unsigned int windowID) {
	wailsRunOnWorker(^{
		goMacosOnDragExit(windowID);
	});
}

void macosOnDragOver(unsigned int windowID, int x, int y) {
	wailsRunOnWorker(^{
		goMacosOnDragOver(windowID, x, y);
	});
}

void processMenuItemClick(unsigned int menuID) {
	wailsRunOnWorker(^{
		goProcessMenuItemClick(menuID);
	});
}

bool shouldQuitApplication(void) {
	__block bool quit = false;
	wailsRunOnWorker(^{
		quit = goShouldQuitApplication();
	});
	return quit;
}

void cleanup(void) {
	wailsRunOnWorker(^{
		goCleanup();
	});
}

void HandleOpenFile(char *filePath) {
	wailsRunOnWorker(^{
		goHandleOpenFile(filePath);
	});
}

void HandleOpenURL(char *url) {
	wailsRunOnWorker(^{
		goHandleOpenURL(url);
	});
}

void systrayClickCallback(long id, int buttonID) {
	wailsRunOnWorker(^{
		goSystrayClickCallback(id, buttonID);
	});
}

int systrayPreClickCallback(long id, int buttonID) {
	__block int action = 0;
	wailsRunOnWorker(^{
		action = goSystrayPreClickCallback(id, buttonID);
	});
	return action;
}

void dialogCallback(int id, int buttonPressed) {
	wailsRunOnWorker(^{
		goDialogCallback(id, buttonPressed);
	});
}

void openFileDialogCallback(uint id, char *path) {
	wailsRunOnWorker(^{
		goOpenFileDialogCallback(id, path);
	});
}

void openFileDialogCallbackEnd(uint id) {
	wailsRunOnWorker(^{
		goOpenFileDialogCallbackEnd(id);
	});
}

void saveFileDialogCallback(uint id, char *path) {
	wailsRunOnWorker(^{
		goSaveFileDialogCallback(id, path);
	});
}

int resolveMediaCapturePermission(unsigned int windowID, bool needAudio, bool needVideo) {
	__block int decision = 0;
	wailsRunOnWorker(^{
		decision = goResolveMediaCapturePermission(windowID, needAudio, needVideo);
	});
	return decision;
}

void globalShortcutCallback(int id) {
	wailsRunOnWorker(^{
		goGlobalShortcutCallback(id);
	});
}

bool windowShouldUnconditionallyClose(unsigned int windowID) {
	__block bool unconditionallyClose = false;
	wailsRunOnWorker(^{
		unconditionallyClose = goWindowShouldUnconditionallyClose(windowID);
	});
	return unconditionallyClose;
}

bool windowIsHidden(unsigned int windowID) {
	__block bool hidden = false;
	wailsRunOnWorker(^{
		hidden = goWindowIsHidden(windowID);
	});
	return hidden;
}

void handleSecondInstanceData(char *message) {
	wailsRunOnWorker(^{
		goHandleSecondInstanceData(message);
	});
}
