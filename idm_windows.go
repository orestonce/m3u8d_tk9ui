//go:build windows && idm

package main

import (
	"github.com/ZenLiuCN/fn"
	. "modernc.org/tk9.0"
	"regexp"
	"syscall"
	"unsafe"
)

func IDMConfig(s *TKApp) bool {
	regUrl := fn.Panic1(regexp.Compile(s.Get(regIDMUrl)))
	regName := fn.Panic1(regexp.Compile(s.Get(regIDMName)))
	s.locator = s.tab0.TButton(Txt(s.Get(btnFindIDMWindow)), Command(func() {
		if hwnd := findWindow(s.Get(txtIDMPromptClass), s.Get(txtIDMPromptTitle)); hwnd != 0 {
			var sx = 0
			enumChildWindows(hwnd, func(w HWND) bool {
				ci, ok := getClassName(w)
				if !ok {
					return true
				}
				if ci == "Edit" {
					x := getEditText(w)
					if regUrl.MatchString(x) {
						s.uri.Configure(Textvariable(x))
						sx++
					} else if xs := regName.FindStringSubmatch(x); xs != nil {
						s.file.Configure(Textvariable(xs[len(xs)-1]))
						sx++
					}
				}
				return sx < 2
			})
			if cc := s.Get(txtIDMCancel); cc != "" {
				enumChildWindows(hwnd, func(w HWND) bool {
					ci, ok := getClassName(w)
					if !ok {
						return true
					}
					if ci == "Button" {
						x := getEditText(w)
						if x == cc {
							postMessage(w, WM_LBUTTONDOWN, 0, 0)
							postMessage(w, WM_LBUTTONUP, 0, 0)
							return false
						}
					}
					return true
				})
			}
		}
	}))
	Tooltip(s.locator, s.Get(tipIDMConfig))
	return true
}

type HWND uintptr

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	pFindWindow       = user32.NewProc("FindWindowW")
	pEnumChildWindows = user32.NewProc("EnumChildWindows")
	pGetClassName     = user32.NewProc("GetClassNameW")
	pSendMessage      = user32.NewProc("SendMessageW")
	pPostMessage      = user32.NewProc("PostMessageW")
)

const (
	WM_GETTEXTLENGTH = 14
	WM_GETTEXT       = 13
	WM_LBUTTONDOWN   = 513
	WM_LBUTTONUP     = 514
)

func findWindow(className, windowName string) HWND {
	var class, window uintptr
	if className != "" {
		class = uintptr(unsafe.Pointer(fn.Panic1(syscall.UTF16PtrFromString(className))))
	}
	if windowName != "" {
		window = uintptr(unsafe.Pointer(fn.Panic1(syscall.UTF16PtrFromString(windowName))))
	}
	ret, _, _ := pFindWindow.Call(class, window)
	return HWND(ret)
}
func enumChildWindows(parent HWND, callback func(window HWND) bool) bool {
	f := syscall.NewCallback(func(w, _ uintptr) uintptr {
		if callback(HWND(w)) {
			return 1
		}
		return 0
	})
	ret, _, _ := pEnumChildWindows.Call(uintptr(parent), f, 0)
	return ret != 0
}
func getClassName(window HWND) (string, bool) {
	var output [256]uint16
	ret, _, _ := pGetClassName.Call(
		uintptr(window),
		uintptr(unsafe.Pointer(&output[0])),
		uintptr(len(output)),
	)
	return syscall.UTF16ToString(output[:]), ret != 0
}
func getEditText(w HWND) string {
	n := sendMessage(w, WM_GETTEXTLENGTH, 0, 0)
	buf := make([]uint16, n+1)
	n = sendMessage(w, WM_GETTEXT, uintptr(n+1), uintptr(unsafe.Pointer(&buf[0])))
	return syscall.UTF16ToString(buf[:n])
}

func sendMessage(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr {
	ret, _, _ := pSendMessage.Call(
		uintptr(hwnd),
		uintptr(msg),
		wParam,
		lParam,
	)
	return ret
}

func postMessage(hwnd HWND, msg uint32, wParam, lParam uintptr) bool {
	ret, _, _ := pPostMessage.Call(
		uintptr(hwnd),
		uintptr(msg),
		wParam,
		lParam,
	)
	return ret != 0
}
