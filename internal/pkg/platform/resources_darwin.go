//go:build darwin
// +build darwin

package platform

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreFoundation

#include <CoreFoundation/CoreFoundation.h>
*/
import "C"
import (
	"unsafe"
)

// ResourcesPath returns the path to the app's resources directory.
func ResourcesPath() string {
	// get main bundle
	mainBundle := C.CFBundleGetMainBundle()

	// get resources URL
	resourcesURL := C.CFBundleCopyResourcesDirectoryURL(mainBundle)

	// get resources path
	var resourcesPath [4096]C.UInt8
	uchar := C.CFURLGetFileSystemRepresentation(resourcesURL, 1 /* true */, (*C.UInt8)(&resourcesPath[0]), 4096)
	if uchar == 0 {
		return ""
	}

	return C.GoString((*C.char)(unsafe.Pointer(&resourcesPath[0])))
}
