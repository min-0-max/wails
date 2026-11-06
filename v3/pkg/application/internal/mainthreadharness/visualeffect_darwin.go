//go:build darwin && !ios && !server

package mainthreadharness

/*
#cgo CFLAGS: -mmacosx-version-min=12.0 -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>

// Defined in package application.
extern void wailsRunOnMainSync(void (^call)(void));
extern NSVisualEffectView* wailsNewVisualEffectView(int style, int material, double cornerRadius);

typedef struct {
	long material;
	long blendingMode;
	long state;
	int emphasized;
	int hasAppearance;
	int osMajor;
} wails_visual_effect_t;

static wails_visual_effect_t wailsProbeVisualEffect(int style, int material) {
	__block wails_visual_effect_t result = {0};
	wailsRunOnMainSync(^{
		NSVisualEffectView *view = wailsNewVisualEffectView(style, material, 0);
		result.material = (long)view.material;
		result.blendingMode = (long)view.blendingMode;
		result.state = (long)view.state;
		result.emphasized = view.emphasized;
		result.hasAppearance = view.appearance != nil;
		result.osMajor = (int)[[NSProcessInfo processInfo] operatingSystemVersion].majorVersion;
	});
	return result;
}
*/
import "C"

// The NSVisualEffectView values that the probe reports.
var (
	MaterialHeaderView          = int(C.NSVisualEffectMaterialHeaderView)
	MaterialHUDWindow           = int(C.NSVisualEffectMaterialHUDWindow)
	MaterialFullScreenUI        = int(C.NSVisualEffectMaterialFullScreenUI)
	MaterialContentBackground   = int(C.NSVisualEffectMaterialContentBackground)
	MaterialUnderPageBackground = int(C.NSVisualEffectMaterialUnderPageBackground)
	MaterialSidebar             = int(C.NSVisualEffectMaterialSidebar)

	BlendingBehindWindow = int(C.NSVisualEffectBlendingModeBehindWindow)
	BlendingWithinWindow = int(C.NSVisualEffectBlendingModeWithinWindow)

	StateFollowsWindowActiveState = int(C.NSVisualEffectStateFollowsWindowActiveState)
)

// VisualEffect is what the visual effect view that stands in for the liquid
// glass is set to.
type VisualEffect struct {
	Material      int
	BlendingMode  int
	State         int
	Emphasized    bool
	HasAppearance bool
	// OSMajor is the major version of the running macOS.
	OSMajor int
}

// ProbeVisualEffect creates the visual effect view for a liquid glass style
// and material, without adding it to a window, and reads it back.
func ProbeVisualEffect(style, material int) VisualEffect {
	r := C.wailsProbeVisualEffect(C.int(style), C.int(material))
	return VisualEffect{
		Material:      int(r.material),
		BlendingMode:  int(r.blendingMode),
		State:         int(r.state),
		Emphasized:    r.emphasized != 0,
		HasAppearance: r.hasAppearance != 0,
		OSMajor:       int(r.osMajor),
	}
}
