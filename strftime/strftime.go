// Package strftime wraps the C stdlib strftime and strptime functions.
package strftime

/*
#include <stdlib.h>
#include <time.h>
#include <errno.h>
int read_errno() { return errno; }
*/
import "C"
import (
	"syscall"
	"time"
	"unsafe"
)

// Strftime wraps the C Strftime function
func Strftime(format string, t time.Time) (string, error) {
	_, off := t.Zone()
	tz := t.Location()
	var (
		secs    = t.Sub(time.Date(1970, 1, 1, 0, 0, 0, 0, tz)).Seconds() - float64(off)
		tt      = C.time_t(secs)
		tm      = C.struct_tm{}
		cFormat = C.CString(format)
		cOut    [256]C.char
	)
	defer C.free(unsafe.Pointer(cFormat)) // nolint: gas
	C.localtime_r(&tt, &tm)
	size := C.strftime(&cOut[0], C.size_t(len(cOut)), cFormat, &tm)
	// This is not thread-safe. Even if we add a mutex, it is not thread-safe
	// relative to other C stdlib calls that don't use it.
	// OTOH, I can't find a format string that C strftime thinks is illegal,
	// so maybe this is a non-issue
	if size == 0 {
		errno := C.read_errno()
		return "", syscall.Errno(errno)
	}
	return C.GoString(&cOut[0]), nil
}

// Strptime wraps the C strptime function
// func Strptime(format, s string) (time.Time, error) {
// 	var (
// 		tm      = C.struct_tm{}
// 		cFormat = C.CString(format)
// 		cin     = C.CString(s)
// 	)
// 	defer C.free(unsafe.Pointer(cin))     // nolint: gas
// 	defer C.free(unsafe.Pointer(cFormat)) // nolint: gas
// 	ptr := C.strptime(cin, cFormat, &tm)
// 	if ptr == nil {
// 		var zero time.Time
// 		return zero, &time.ParseError{
// 			Layout:     format,
// 			Value:      s,
// 			LayoutElem: format,
// 			ValueElem:  s,
// 		}
// 	}
// 	return time.Date(int(tm.tm_year)+1900, time.Month(tm.tm_mon+1), int(tm.tm_mday), int(tm.tm_hour), int(tm.tm_min), int(tm.tm_sec), 0, time.UTC), nil
// }
