package window

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

var (
	user32         = syscall.NewLazyDLL("user32.dll")
	createWindowEx = user32.NewProc("CreateWindowExW")
	setLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
	getWindowLongPtr = user32.NewProc("GetWindowLongPtrW")
	setWindowLongPtr = user32.NewProc("SetWindowLongPtrW")
	showWindow      = user32.NewProc("ShowWindow")
	updateWindow    = user32.NewProc("UpdateWindow")
	dispatchMessage = user32.NewProc("DispatchMessageW")
	getMessage      = user32.NewProc("GetMessageW")
	translateMessage = user32.NewProc("TranslateMessage")
	registerClassEx = user32.NewProc("RegisterClassExW")
	defWindowProc   = user32.NewProc("DefWindowProcW")
	setCursorPos    = user32.NewProc("SetCursorPos")
	getCursorPos    = user32.NewProc("GetCursorPos")
)

const (
	WS_EX_LAYERED   = 0x80000
	WS_EX_TOPMOST   = 0x00000008
	WS_EX_TOOLWINDOW = 0x00000080
	WS_POPUP        = 0x80000000
	WS_VISIBLE      = 0x10000000

	LWA_COLORKEY = 0x00000001
	LWA_ALPHA    = 0x00000002

	GWL_EXSTYLE = -20

	WM_DESTROY   = 0x0002
	WM_PAINT     = 0x000F
	WM_MOUSEMOVE = 0x0020
	WM_LBUTTONDOWN = 0x0201
	WM_LBUTTONUP   = 0x0202
	WM_RBUTTONDOWN = 0x0204
	WM_MOUSELEAVE = 0x02A3
)

type Window struct {
	hwnd     win.HWND
	width    int
	height   int
	x        int
	y        int
	visible  bool
	transparentColor win.COLORREF
}

func NewWindow(title string, x, y, width, height int) (*Window, error) {
	w := &Window{
		width:    width,
		height:   height,
		x:        x,
		y:        y,
		transparentColor: win.RGB(0, 0, 0),
	}

	err := w.registerWindowClass()
	if err != nil {
		return nil, err
	}

	w.hwnd, err = w.createWindow(title)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Window) registerWindowClass() error {
	className, _ := syscall.UTF16PtrFromString("DesktopPetClass")

	wc := win.WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(win.WNDCLASSEX{})),
		Style:         win.CS_HREDRAW | win.CS_VREDRAW | win.CS_DBLCLKS,
		LpfnWndProc:   syscall.NewCallback(w.wndProc),
		HInstance:     win.GetModuleHandle(nil),
		HCursor:       win.LoadCursor(0, win.IDC_ARROW),
		LpszClassName: className,
	}

	_, _, err := registerClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	if err != nil && err.Error() != "The operation completed successfully." {
		return fmt.Errorf("registerClassEx failed: %v", err)
	}

	return nil
}

func (w *Window) createWindow(title string) (win.HWND, error) {
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	className, _ := syscall.UTF16PtrFromString("DesktopPetClass")

	exStyle := WS_EX_LAYERED | WS_EX_TOPMOST | WS_EX_TOOLWINDOW
	style := WS_POPUP | WS_VISIBLE

	hwnd, _, err := createWindowEx.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(style),
		uintptr(w.x),
		uintptr(w.y),
		uintptr(w.width),
		uintptr(w.height),
		0,
		0,
		uintptr(win.GetModuleHandle(nil)),
		0,
	)

	if hwnd == 0 {
		return 0, fmt.Errorf("createWindowEx failed: %v", err)
	}

	_, _, err = setLayeredWindowAttributes.Call(
		hwnd,
		uintptr(w.transparentColor),
		255,
		uintptr(LWA_ALPHA),
	)
	if err != nil && err.Error() != "The operation completed successfully." {
		return 0, fmt.Errorf("setLayeredWindowAttributes failed: %v", err)
	}

	return win.HWND(hwnd), nil
}

func (w *Window) wndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_DESTROY:
		win.PostQuitMessage(0)
		return 0
	case WM_PAINT:
		return w.handlePaint(hwnd)
	case WM_LBUTTONDOWN:
		return w.handleMouseDown(hwnd, wParam, lParam)
	case WM_LBUTTONUP:
		return w.handleMouseUp(hwnd, wParam, lParam)
	case WM_MOUSEMOVE:
		return w.handleMouseMove(hwnd, wParam, lParam)
	case WM_RBUTTONDOWN:
		return w.handleRightClick(hwnd, wParam, lParam)
	}

	result, _, _ := defWindowProc.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return result
}

func (w *Window) handlePaint(hwnd win.HWND) uintptr {
	var ps win.PAINTSTRUCT
	hdc := win.BeginPaint(hwnd, &ps)
	
	win.EndPaint(hwnd, &ps)
	return 0
}

func (w *Window) handleMouseDown(hwnd win.HWND, wParam, lParam uintptr) uintptr {
	return 0
}

func (w *Window) handleMouseUp(hwnd win.HWND, wParam, lParam uintptr) uintptr {
	return 0
}

func (w *Window) handleMouseMove(hwnd win.HWND, wParam, lParam uintptr) uintptr {
	return 0
}

func (w *Window) handleRightClick(hwnd win.HWND, wParam, lParam uintptr) uintptr {
	return 0
}

func (w *Window) Show() error {
	if w.hwnd == 0 {
		return fmt.Errorf("window not created")
	}

	_, _, err := showWindow.Call(uintptr(w.hwnd), uintptr(win.SW_SHOW))
	if err != nil && err.Error() != "The operation completed successfully." {
		return fmt.Errorf("showWindow failed: %v", err)
	}

	w.visible = true
	return nil
}

func (w *Window) Hide() error {
	if w.hwnd == 0 {
		return fmt.Errorf("window not created")
	}

	_, _, err := showWindow.Call(uintptr(w.hwnd), uintptr(win.SW_HIDE))
	if err != nil && err.Error() != "The operation completed successfully." {
		return fmt.Errorf("hideWindow failed: %v", err)
	}

	w.visible = false
	return nil
}

func (w *Window) Update() error {
	if w.hwnd == 0 {
		return fmt.Errorf("window not created")
	}

	_, _, err := updateWindow.Call(uintptr(w.hwnd))
	if err != nil && err.Error() != "The operation completed successfully." {
		return fmt.Errorf("updateWindow failed: %v", err)
	}

	return nil
}

func (w *Window) SetPosition(x, y int) {
	w.x = x
	w.y = y
	win.SetWindowPos(w.hwnd, 0, x, y, 0, 0, win.SWP_NOZORDER|win.SWP_NOSIZE)
}

func (w *Window) GetPosition() (int, int) {
	return w.x, w.y
}

func (w *Window) GetSize() (int, int) {
	return w.width, w.height
}

func (w *Window) SetSize(width, height int) {
	w.width = width
	w.height = height
	win.SetWindowPos(w.hwnd, 0, 0, 0, width, height, win.SWP_NOZORDER|win.SWP_NOMOVE)
}

func (w *Window) RunMessageLoop() {
	var msg win.MSG
	for {
		ret, _, _ := getMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 {
			break
		}
		translateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		dispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func (w *Window) Close() {
	if w.hwnd != 0 {
		win.DestroyWindow(w.hwnd)
		w.hwnd = 0
	}
}
