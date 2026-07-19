//go:build windows

package windows

import (
	"syscall"
	"unsafe"

	"desktopPet/platform"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	createWindowExW            = user32.NewProc("CreateWindowExW")
	registerClassExW           = user32.NewProc("RegisterClassExW")
	defWindowProcW             = user32.NewProc("DefWindowProcW")
	postQuitMessage            = user32.NewProc("PostQuitMessage")
	showWindow                 = user32.NewProc("ShowWindow")
	updateWindow               = user32.NewProc("UpdateWindow")
	getMessageW                = user32.NewProc("GetMessageW")
	translateMessage           = user32.NewProc("TranslateMessage")
	dispatchMessageW           = user32.NewProc("DispatchMessageW")
	setLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
	setWindowPos               = user32.NewProc("SetWindowPos")
	getDC                      = user32.NewProc("GetDC")
	releaseDC                  = user32.NewProc("ReleaseDC")
	loadCursorW                = user32.NewProc("LoadCursorW")
	getModuleHandleW           = kernel32.NewProc("GetModuleHandleW")

	createCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	deleteDC               = gdi32.NewProc("DeleteDC")
	createCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	deleteObject           = gdi32.NewProc("DeleteObject")
	selectObject           = gdi32.NewProc("SelectObject")
	setDIBitsToDevice      = gdi32.NewProc("SetDIBitsToDevice")
	bitBlt                 = gdi32.NewProc("BitBlt")
	setBkColor             = gdi32.NewProc("SetBkColor")
	extTextOutW            = gdi32.NewProc("ExtTextOutW")
	createSolidBrush       = gdi32.NewProc("CreateSolidBrush")
	createPen              = gdi32.NewProc("CreatePen")
	roundRect              = gdi32.NewProc("RoundRect")
	setTextColor           = gdi32.NewProc("SetTextColor")
	setBkMode              = gdi32.NewProc("SetBkMode")
	textOutW               = gdi32.NewProc("TextOutW")
	moveToEx               = gdi32.NewProc("MoveToEx")
	lineTo                 = gdi32.NewProc("LineTo")
	closeFigure            = gdi32.NewProc("CloseFigure")
	fillPath               = gdi32.NewProc("FillPath")
)

const (
	WS_EX_LAYERED     = 0x80000
	WS_EX_TOPMOST     = 0x00000008
	WS_EX_TOOLWINDOW  = 0x00000080
	WS_POPUP          = 0x80000000

	LWA_COLORKEY      = 0x00000001
	LWA_ALPHA         = 0x00000002

	WM_DESTROY        = 0x0002
	WM_PAINT          = 0x000F
	WM_LBUTTONDOWN    = 0x0201
	WM_LBUTTONUP      = 0x0202
	WM_MOUSEMOVE      = 0x0020
	WM_RBUTTONDOWN    = 0x0204

	SW_SHOW           = 5

	SWP_NOZORDER      = 0x0004
	SWP_NOSIZE        = 0x0001
	SWP_NOMOVE        = 0x0002

	COLOR_WINDOW      = 5

	CS_HREDRAW        = 0x0002
	CS_VREDRAW        = 0x0001
	CS_DBLCLKS        = 0x0008

	IDC_ARROW         = 32512

	SRCCOPY           = 0xCC0020

	ETO_OPAQUE        = 0x0002

	BI_RGB            = 0
	DIB_RGB_COLORS    = 0

	SM_CXSCREEN       = 0
	SM_CYSCREEN       = 1
)

type HWND uintptr
type HDC uintptr

type POINT struct {
	X, Y int32
}

type SIZE struct {
	Cx, Cy int32
}

type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground HBRUSH
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

type HBRUSH uintptr

type MSG struct {
	Hwnd    HWND
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type BITMAPINFOHEADER struct {
	BiSize        uint32
	BiWidth       int32
	BiHeight      int32
	BiPlanes      uint16
	BiBitCount    uint16
	BiCompression uint32
	BiSizeImage   uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed     uint32
	BiClrImportant uint32
}

type BITMAPINFO struct {
	BmiHeader BITMAPINFOHEADER
	BmiColors [1]uint32
}

type Window struct {
	Hwnd         HWND
	Width        int
	Height       int
	X            int
	Y            int
	OnClick      platform.MouseEventHandler
	OnRightClick platform.MouseEventHandler
	OnMouseDown  platform.MouseEventHandler
	OnMouseUp    platform.MouseEventHandler
	OnMouseMove  platform.MouseEventHandler
	running      bool
}

func NewPlatform() platform.Platform {
	return &Window{}
}

func (w *Window) Init() error {
	return nil
}

func (w *Window) CreateWindow(title string, x, y, width, height int) error {
	w.Width = width
	w.Height = height
	w.X = x
	w.Y = y

	err := w.registerWindowClass()
	if err != nil {
		return err
	}

	w.Hwnd, err = w.createWindow(title)
	if err != nil {
		return err
	}

	return nil
}

func (w *Window) registerWindowClass() error {
	className, _ := syscall.UTF16PtrFromString("DesktopPetClass")

	hCursor, _, _ := loadCursorW.Call(0, uintptr(IDC_ARROW))

	wc := WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEX{})),
		Style:         CS_HREDRAW | CS_VREDRAW | CS_DBLCLKS,
		LpfnWndProc:   syscall.NewCallback(w.wndProc),
		HInstance:     w.getModuleHandle(),
		HCursor:       hCursor,
		HbrBackground: HBRUSH(COLOR_WINDOW + 1),
		LpszClassName: className,
	}

	if _, _, err := registerClassExW.Call(uintptr(unsafe.Pointer(&wc))); err != nil && err.Error() != "The operation completed successfully." {
		return err
	}

	return nil
}

func (w *Window) getModuleHandle() uintptr {
	mod, _, _ := getModuleHandleW.Call(0)
	return mod
}

func (w *Window) createWindow(title string) (HWND, error) {
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	className, _ := syscall.UTF16PtrFromString("DesktopPetClass")

	exStyle := WS_EX_LAYERED | WS_EX_TOPMOST | WS_EX_TOOLWINDOW
	style := WS_POPUP

	hwnd, _, err := createWindowExW.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(style),
		uintptr(w.X),
		uintptr(w.Y),
		uintptr(w.Width),
		uintptr(w.Height),
		0,
		0,
		w.getModuleHandle(),
		0,
	)

	if hwnd == 0 {
		return 0, err
	}

	setLayeredWindowAttributes.Call(
		hwnd,
		uintptr(0),
		uintptr(255),
		uintptr(LWA_ALPHA),
	)

	return HWND(hwnd), nil
}

func (w *Window) wndProc(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_DESTROY:
		w.running = false
		postQuitMessage.Call(0)
		return 0
	case WM_PAINT:
		return w.handlePaint(hwnd)
	case WM_LBUTTONDOWN:
		x := int(lParam & 0xFFFF)
		y := int((lParam >> 16) & 0xFFFF)
		if w.OnMouseDown != nil {
			w.OnMouseDown(platform.MouseEvent{X: x, Y: y})
		}
		return 0
	case WM_LBUTTONUP:
		x := int(lParam & 0xFFFF)
		y := int((lParam >> 16) & 0xFFFF)
		if w.OnMouseUp != nil {
			w.OnMouseUp(platform.MouseEvent{X: x, Y: y})
		}
		return 0
	case WM_MOUSEMOVE:
		x := int(lParam & 0xFFFF)
		y := int((lParam >> 16) & 0xFFFF)
		if w.OnMouseMove != nil {
			w.OnMouseMove(platform.MouseEvent{X: x, Y: y})
		}
		return 0
	case WM_RBUTTONDOWN:
		x := int(lParam & 0xFFFF)
		y := int((lParam >> 16) & 0xFFFF)
		if w.OnRightClick != nil {
			w.OnRightClick(platform.MouseEvent{X: x, Y: y})
		}
		return 0
	}

	result, _, _ := defWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return result
}

func (w *Window) handlePaint(hwnd HWND) uintptr {
	hdc, _, _ := getDC.Call(uintptr(hwnd))
	releaseDC.Call(uintptr(hwnd), hdc)
	return 0
}

func (w *Window) Show() {
	showWindow.Call(uintptr(w.Hwnd), uintptr(SW_SHOW))
}

func (w *Window) Run() {
	w.running = true
	var msg MSG
	for w.running {
		ret, _, _ := getMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 {
			break
		}
		translateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		dispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func (w *Window) Close() {
	if w.Hwnd != 0 {
		var destroyWindow = user32.NewProc("DestroyWindow")
		destroyWindow.Call(uintptr(w.Hwnd))
		w.Hwnd = 0
	}
}

func (w *Window) Navigate(url string) {
}

func (w *Window) SetSize(width, height int) {
	w.Width = width
	w.Height = height
	setWindowPos.Call(uintptr(w.Hwnd), 0, 0, 0, uintptr(width), uintptr(height), uintptr(SWP_NOZORDER|SWP_NOMOVE))
}

func (w *Window) SetPosition(x, y int) {
	w.X = x
	w.Y = y
	setWindowPos.Call(uintptr(w.Hwnd), 0, uintptr(x), uintptr(y), 0, 0, uintptr(SWP_NOZORDER|SWP_NOSIZE))
}

func (w *Window) GetSize() (int, int) {
	return w.Width, w.Height
}

func (w *Window) GetPosition() (int, int) {
	return w.X, w.Y
}

func (w *Window) ShowNotification(message string) {
}

func (w *Window) SetOnClick(handler platform.MouseEventHandler) {
	w.OnClick = handler
}

func (w *Window) SetOnRightClick(handler platform.MouseEventHandler) {
	w.OnRightClick = handler
}

func (w *Window) SetOnMouseDown(handler platform.MouseEventHandler) {
	w.OnMouseDown = handler
}

func (w *Window) SetOnMouseUp(handler platform.MouseEventHandler) {
	w.OnMouseUp = handler
}

func (w *Window) SetOnMouseMove(handler platform.MouseEventHandler) {
	w.OnMouseMove = handler
}

func (w *Window) Eval(code string) {
}

func (w *Window) UpdatePet(action string, frame int) {
	w.draw(action, frame)
}

func (w *Window) ShowBubble(text string) {
	w.drawBubble(text)
}

func (w *Window) HideBubble() {
}

func (w *Window) SetDragging(dragging bool) {
}

func (w *Window) GetDC() HDC {
	hdc, _, _ := getDC.Call(uintptr(w.Hwnd))
	return HDC(hdc)
}

func (w *Window) ReleaseDC(hdc HDC) {
	releaseDC.Call(uintptr(w.Hwnd), uintptr(hdc))
}

func (w *Window) draw(action string, frame int) {
	hdc := w.GetDC()
	if hdc == 0 {
		return
	}
	defer w.ReleaseDC(hdc)

	setBkColor.Call(uintptr(hdc), uintptr(RGB(0, 0, 0)))
	extTextOutW.Call(uintptr(hdc), 0, 0, uintptr(ETO_OPAQUE), 0, 0, 0, 0)

	colors := map[string][2]uint32{
		"idle":    {RGB(255, 182, 193), RGB(255, 160, 176)},
		"jump":    {RGB(135, 206, 250), RGB(173, 216, 230)},
		"head_pat": {RGB(255, 218, 185), RGB(255, 228, 225)},
		"wave":    {RGB(144, 238, 144), RGB(152, 251, 152)},
		"happy":   {RGB(255, 215, 0), RGB(255, 255, 0)},
	}

	colorList, ok := colors[action]
	if !ok {
		colorList = colors["idle"]
	}

	brush, _, _ := createSolidBrush.Call(uintptr(colorList[frame%2]))
	if brush != 0 {
		oldBrush, _, _ := selectObject.Call(uintptr(hdc), brush)
		roundRect.Call(uintptr(hdc), 0, 0, uintptr(w.Width), uintptr(w.Height), uintptr(10), uintptr(10))
		selectObject.Call(uintptr(hdc), oldBrush)
		deleteObject.Call(brush)
	}

	w.drawBubble("")
}

func (w *Window) drawBubble(text string) {
	if text == "" {
		return
	}

	hdc := w.GetDC()
	if hdc == 0 {
		return
	}
	defer w.ReleaseDC(hdc)

	width := 150
	height := 60
	x := w.Width/2 - width/2
	y := -height

	brush, _, _ := createSolidBrush.Call(uintptr(RGB(255, 255, 255)))
	if brush == 0 {
		return
	}
	defer deleteObject.Call(brush)

	pen, _, _ := createPen.Call(uintptr(0), uintptr(1), uintptr(RGB(0, 0, 0)))
	if pen == 0 {
		return
	}
	defer deleteObject.Call(pen)

	oldBrush, _, _ := selectObject.Call(uintptr(hdc), brush)
	oldPen, _, _ := selectObject.Call(uintptr(hdc), pen)
	defer selectObject.Call(uintptr(hdc), oldBrush)
	defer selectObject.Call(uintptr(hdc), oldPen)

	roundRect.Call(uintptr(hdc), uintptr(x), uintptr(y), uintptr(x+width), uintptr(y+height), uintptr(10), uintptr(10))

	setTextColor.Call(uintptr(hdc), uintptr(RGB(0, 0, 0)))
	setBkMode.Call(uintptr(hdc), uintptr(1))

	textPtr, _ := syscall.UTF16PtrFromString(text)
	textOutW.Call(uintptr(hdc), uintptr(x+10), uintptr(y+15), uintptr(unsafe.Pointer(textPtr)), uintptr(len(text)))

	moveToEx.Call(uintptr(hdc), uintptr(x+width/2), uintptr(y+height), 0)
	lineTo.Call(uintptr(hdc), uintptr(x+width/2+10), uintptr(y+height+10))
	lineTo.Call(uintptr(hdc), uintptr(x+width/2-10), uintptr(y+height+10))
	closeFigure.Call(uintptr(hdc))
	fillPath.Call(uintptr(hdc))
}

func RGB(r, g, b byte) uint32 {
	return uint32(r) | uint32(g)<<8 | uint32(b)<<16
}

func GetSystemMetrics(index int) int {
	getSystemMetrics := user32.NewProc("GetSystemMetrics")
	ret, _, _ := getSystemMetrics.Call(uintptr(index))
	return int(ret)
}
