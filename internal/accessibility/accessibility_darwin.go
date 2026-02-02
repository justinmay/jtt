package accessibility

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreFoundation -framework ApplicationServices
#include <CoreFoundation/CoreFoundation.h>
#include <ApplicationServices/ApplicationServices.h>

bool checkAccessibility(bool prompt) {
    NSDictionary *options = @{(__bridge id)kAXTrustedCheckOptionPrompt: @(prompt)};
    return AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options);
}
*/
import "C"

func CheckAccessibility(prompt bool) bool {
	return bool(C.checkAccessibility(C.bool(prompt)))
}
