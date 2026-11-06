//go:build darwin && !ios

package events

/*
#cgo CFLAGS: -mmacosx-version-min=12.0 -x objective-c
#cgo LDFLAGS: -framework Cocoa

#include "events_darwin.h"
#include <stdlib.h>
#include <stdbool.h>

bool hasListener[MAX_EVENTS] = {false};

void registerListener(unsigned int event) {
	hasListener[event] = true;
}

bool hasListeners(unsigned int event) {
	//return hasListener[event];
	return true;
}

*/
import "C"
