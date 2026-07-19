//go:build windows

package platform

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	// user32 procs
	createWindowExW     = user32.NewProc("CreateWindowExW")
	registerClassExW    = user32.NewProc("RegisterClassExW")
	defWindowProcW      = user32.NewProc("DefWindowProcW")
	postQuitMessage     = user32.NewProc("PostQuitMessage")
	showWindow          = user32.NewProc("ShowWindow")
	updateWindow        = user32.NewProc("UpdateWindow")
	getMessageW         = user32.NewProc("GetMessageW")
	translateMessage    = user32.NewProc("TranslateMessage")
	dispatchMessageW    = user32.NewProc("DispatchMessageW")
	setWindowPos        = user32.NewProc("SetWindowPos")
	getDC               = user32.NewProc("GetDC")
	releaseDC           = user32.NewProc("ReleaseDC")
	loadCursorW         = user32.NewProc("LoadCursorW")
	setCursor           = user32.NewProc("SetCursor")
	getModuleHandleW    = kernel32.NewProc("GetModuleHandleW")
	setCapture          = user32.NewProc("SetCapture")
	releaseCapture      = user32.NewProc("ReleaseCapture")
	getCursorPos        = user32.NewProc("GetCursorPos")
	getWindowRect       = user32.NewProc("GetWindowRect")
	postMessageW        = user32.NewProc("PostMessageW")
	beginPaint          = user32.NewProc("BeginPaint")
	endPaint            = user32.NewProc("EndPaint")
	updateLayeredWindow = user32.NewProc("UpdateLayeredWindow")
	destroyWindow       = user32.NewProc("DestroyWindow")
	getSystemMetrics    = user32.NewProc("GetSystemMetrics")
	// 菜单相关
	createPopupMenu  = user32.NewProc("CreatePopupMenu")
	appendMenuW      = user32.NewProc("AppendMenuW")
	trackPopupMenu   = user32.NewProc("TrackPopupMenu")
	createPopupMenu_ = user32.NewProc("CreatePopupMenu")
	destroyMenu      = user32.NewProc("DestroyMenu")

	// gdi32 procs
	createSolidBrush   = gdi32.NewProc("CreateSolidBrush")
	createPen          = gdi32.NewProc("CreatePen")
	roundRect          = gdi32.NewProc("RoundRect")
	selectObject       = gdi32.NewProc("SelectObject")
	deleteObject       = gdi32.NewProc("DeleteObject")
	setTextColor       = gdi32.NewProc("SetTextColor")
	setBkMode          = gdi32.NewProc("SetBkMode")
	textOutW           = gdi32.NewProc("TextOutW")
	createCompatibleDC = gdi32.NewProc("CreateCompatibleDC")
	createDIBSection   = gdi32.NewProc("CreateDIBSection")
	deleteDC           = gdi32.NewProc("DeleteDC")
	bitBlt             = gdi32.NewProc("BitBlt")
	createFontW        = gdi32.NewProc("CreateFontW")
)

const (
	WS_EX_LAYERED    = 0x80000
	WS_EX_TOPMOST    = 0x00000008
	WS_EX_TOOLWINDOW = 0x00000080
	WS_POPUP         = 0x80000000
	WS_VISIBLE       = 0x10000000

	LWA_ALPHA = 0x00000002

	GWL_EXSTYLE int32 = -20

	WM_DESTROY         = 0x0002
	WM_PAINT           = 0x000F
	WM_LBUTTONDOWN     = 0x0201
	WM_LBUTTONUP       = 0x0202
	WM_MOUSEMOVE       = 0x0200
	WM_RBUTTONDOWN     = 0x0204
	WM_RBUTTONUP       = 0x0205
	WM_NCHITTEST       = 0x0084
	WM_NCLBUTTONDOWN   = 0x00A1
	WM_NCLBUTTONUP     = 0x00A2
	WM_NCLBUTTONDBLCLK = 0x00A3
	WM_NCRBUTTONDOWN   = 0x00A4
	WM_NCRBUTTONUP     = 0x00A5
	WM_NCMOUSEMOVE     = 0x00A0
	WM_LBUTTONDBLCLK   = 0x0203
	WM_SETCURSOR       = 0x0020
	WM_UPDATE_DISPLAY  = 0x0401
	WM_COMMAND         = 0x0111

	// 菜单相关
	MF_STRING     = 0x00000000
	MF_POPUP      = 0x00000010
	MF_SEPARATOR  = 0x00000800
	TPM_LEFTALIGN = 0x0000

	// 菜单 ID
	ID_SIZE_INCREASE   = 1001
	ID_SIZE_DECREASE   = 1002
	ID_EXIT            = 1003
	ID_EXPRESSION_BASE = 2000 // 表情菜单 ID 从 2000 开始

	HTCAPTION = 2
	HTCLIENT  = 1

	VK_LBUTTON = 0x01

	SW_SHOW = 5

	SWP_NOZORDER = 0x0004
	SWP_NOSIZE   = 0x0001

	COLOR_WINDOW = 5

	CS_HREDRAW = 0x0002
	CS_VREDRAW = 0x0001

	IDC_ARROW = 32512

	AC_SRC_OVER  = 0x00
	AC_SRC_ALPHA = 0x01
	ULW_ALPHA    = 0x02

	DIB_RGB_COLORS = 0

	SM_CXSCREEN = 0
	SM_CYSCREEN = 1

	SRCCOPY = 0x00CC0020

	// 字体相关
	DEFAULT_CHARSET     = 1
	OUT_DEFAULT_PRECIS  = 0
	CLIP_DEFAULT_PRECIS = 0
	DEFAULT_QUALITY     = 0
	DEFAULT_PITCH       = 0
	FF_DONTCARE         = 0

	// 字体粗细
	FW_NORMAL = 400
)

type HWND uintptr
type HDC uintptr
type HBRUSH uintptr

type POINT struct{ X, Y int32 }
type SIZE struct{ Cx, Cy int32 }
type RECT struct{ Left, Top, Right, Bottom int32 }

type BLENDFUNCTION struct {
	BlendOp             byte
	BlendFlags          byte
	SourceConstantAlpha byte
	AlphaFormat         byte
}

type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
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

type MSG struct {
	Hwnd    HWND
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type PAINTSTRUCT struct {
	Hdc         HDC
	FErase      int32
	RcPaint     RECT
	FPaint      int32
	RgbReserved uint32
}

type windowsPlatform struct {
	Hwnd               HWND
	Width              int
	Height             int
	X                  int
	Y                  int
	OnClick            MouseEventHandler
	OnRightClick       MouseEventHandler
	OnMouseDown        MouseEventHandler
	OnMouseUp          MouseEventHandler
	OnMouseMove        MouseEventHandler
	OnExpressionChange ExpressionChangeHandler
	running            bool
	currentAction      string
	currentFrame       int
	currentBubble      string
	// 图片数据
	basePixels []byte
	memDC      uintptr
	hBitmap    uintptr
	bmpPtr     unsafe.Pointer
	hasImage   bool
	// 多图片支持
	images       map[string][]byte // 图片名称 -> 像素数据
	imageOrder   []string          // 图片切换顺序
	currentIdx   int               // 当前图片索引
	imageWidth   int
	imageHeight  int
	bubbleHeight int // 动态气泡高度
	// 缩放
	scale float64 // 当前缩放比例
	// 右键拖动/单击
	dragStartX int32
	dragStartY int32
	winStartX  int32
	winStartY  int32
	dragging   bool
	isDrag     bool // 是否是拖动（vs 单击）
}

func (w *windowsPlatform) Init() error {
	return nil
}

func (w *windowsPlatform) CreateWindow(title string, x, y, width, height int) error {
	w.X = x
	w.Y = y

	// 尝试加载所有宠物图片
	if err := w.loadAllImages(); err == nil {
		width = w.Width
		height = w.Height
		fmt.Printf("已加载宠物图片，尺寸: %dx%d，共 %d 张\n", width, height, len(w.images))
	} else {
		w.Width = width
		w.Height = height
		fmt.Printf("未找到宠物图片，使用默认尺寸: %dx%d\n", width, height)
	}
	w.X = x
	w.Y = y

	if err := w.registerWindowClass(); err != nil {
		return err
	}

	hwnd, err := w.createWindow(title, width, height)
	if err != nil {
		return err
	}
	w.Hwnd = hwnd

	// 创建 DIB section 和内存 DC
	if err := w.createDIBSection(); err != nil {
		return err
	}

	return nil
}

// 加载单个图片并返回像素数据（统一尺寸，带偏移）
func (w *windowsPlatform) loadImageByName(name string, targetWidth, targetHeight int, offsetY int) ([]byte, error) {
	exePath, _ := os.Executable()
	paths := []string{
		filepath.Join("assets", name+".png"),
		filepath.Join(filepath.Dir(exePath), "assets", name+".png"),
		filepath.Join(filepath.Dir(exePath), name+".png"),
	}

	var img image.Image
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		img, _ = png.Decode(f)
		f.Close()
		if img != nil {
			break
		}
	}
	if img == nil {
		return nil, fmt.Errorf("未找到图片: %s", name)
	}

	bounds := img.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()

	// 使用目标尺寸
	width := targetWidth
	height := targetHeight

	// 计算缩放比例（保持宽高比）
	scaleW := float64(width) / float64(origW)
	scaleH := float64(height) / float64(origH)
	scale := scaleW
	if scaleH < scaleW {
		scale = scaleH
	}

	// 计算实际绘制区域（居中）
	drawW := int(float64(origW) * scale)
	drawH := int(float64(origH) * scale)
	offsetX := (width - drawW) / 2

	// 转换为 BGRA 格式，统一尺寸（包含偏移区域）
	pixels := make([]byte, width*(height+offsetY)*4)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := ((y+offsetY)*width + x) * 4
			// 检查是否在绘制区域内
			if x >= offsetX && x < offsetX+drawW && y >= (height-drawH)/2 && y < (height-drawH)/2+drawH {
				srcX := int(float64(x-offsetX) / scale)
				srcY := int(float64(y-(height-drawH)/2) / scale)
				if srcX >= origW {
					srcX = origW - 1
				}
				if srcY >= origH {
					srcY = origH - 1
				}
				r32, g32, b32, a32 := img.At(srcX, srcY).RGBA()
				a := uint8(a32 >> 8)
				if a > 0 {
					pixels[idx+0] = uint8(b32 >> 8) // Blue
					pixels[idx+1] = uint8(g32 >> 8) // Green
					pixels[idx+2] = uint8(r32 >> 8) // Red
					pixels[idx+3] = a               // Alpha
				}
			}
			// 其他区域保持透明（alpha = 0）
		}
	}

	return pixels, nil
}

// 加载所有图片
func (w *windowsPlatform) loadAllImages() error {
	// 图片列表（按顺序）
	imageNames := []string{
		"主体", "挥手", "奔跑", "跳跃", "打哈欠",
		"害羞", "生气", "委屈", "震惊", "喝饮料", "睡觉",
	}

	w.images = make(map[string][]byte)
	w.imageOrder = imageNames
	w.currentIdx = 0
	w.scale = 1.0 // 初始化缩放比例

	// 统一显示尺寸（宽度不变，高度增加给气泡留空间）
	displayWidth := 200
	displayHeight := 300
	bubbleHeight := 80 // 气泡区域高度（调整到合适的距离）

	// 加载所有图片（统一尺寸）
	for _, name := range imageNames {
		pixels, err := w.loadImageByName(name, displayWidth, displayHeight, bubbleHeight)
		if err != nil {
			fmt.Printf("加载图片失败: %s, %v\n", name, err)
			continue
		}
		w.images[name] = pixels
	}

	// 设置尺寸和当前图片（窗口高度 = 图片高度 + 气泡高度）
	w.Width = displayWidth
	w.Height = displayHeight + bubbleHeight
	w.imageWidth = displayWidth
	w.imageHeight = displayHeight
	w.basePixels = w.images[imageNames[0]]
	w.hasImage = true

	return nil
}

func (w *windowsPlatform) createDIBSection() error {
	screenDC, _, _ := getDC.Call(0)
	defer releaseDC.Call(0, screenDC)

	w.memDC, _, _ = createCompatibleDC.Call(screenDC)
	if w.memDC == 0 {
		return fmt.Errorf("CreateCompatibleDC failed")
	}

	bmi := BITMAPINFOHEADER{
		BiSize:        uint32(unsafe.Sizeof(BITMAPINFOHEADER{})),
		BiWidth:       int32(w.Width),
		BiHeight:      -int32(w.Height), // 负值 = 从上到下
		BiPlanes:      1,
		BiBitCount:    32,
		BiCompression: 0, // BI_RGB
	}

	var bmpPtr unsafe.Pointer
	hBmp, _, _ := createDIBSection.Call(
		uintptr(w.memDC),
		uintptr(unsafe.Pointer(&bmi)),
		uintptr(DIB_RGB_COLORS),
		uintptr(unsafe.Pointer(&bmpPtr)),
		0, 0,
	)
	if hBmp == 0 || bmpPtr == nil {
		return fmt.Errorf("CreateDIBSection failed")
	}
	w.hBitmap = hBmp
	w.bmpPtr = bmpPtr

	// 将位图选入内存 DC
	selectObject.Call(uintptr(w.memDC), hBmp)

	return nil
}

func (w *windowsPlatform) registerWindowClass() error {
	className, _ := syscall.UTF16PtrFromString("DesktopPetClass")
	hCursor, _, _ := loadCursorW.Call(0, uintptr(IDC_ARROW))

	wc := WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEX{})),
		Style:         CS_HREDRAW | CS_VREDRAW,
		LpfnWndProc:   syscall.NewCallback(w.wndProc),
		HInstance:     w.getModuleHandle(),
		HCursor:       hCursor,
		HbrBackground: HBRUSH(COLOR_WINDOW + 1),
		LpszClassName: className,
	}

	registerClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	return nil
}

func (w *windowsPlatform) getModuleHandle() uintptr {
	mod, _, _ := getModuleHandleW.Call(0)
	return mod
}

func (w *windowsPlatform) createWindow(title string, width, height int) (HWND, error) {
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	className, _ := syscall.UTF16PtrFromString("DesktopPetClass")

	exStyle := uintptr(WS_EX_TOPMOST | WS_EX_TOOLWINDOW | WS_EX_LAYERED)
	style := uintptr(WS_POPUP | WS_VISIBLE)

	hwnd, _, err := createWindowExW.Call(
		exStyle,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(titlePtr)),
		style,
		uintptr(w.X), uintptr(w.Y),
		uintptr(width), uintptr(height),
		0, 0, w.getModuleHandle(), 0,
	)
	if hwnd == 0 {
		return 0, err
	}
	return HWND(hwnd), nil
}

func (w *windowsPlatform) wndProc(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_DESTROY:
		w.running = false
		postQuitMessage.Call(0)
		return 0

	case WM_PAINT:
		var ps PAINTSTRUCT
		beginPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&ps)))
		endPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&ps)))
		w.updateDisplay()
		return 0

	case WM_NCHITTEST:
		// 返回 HTCLIENT，让左键和右键点击都可以被处理
		return HTCLIENT

	case WM_LBUTTONDOWN:
		// 左键单击：切换图片
		if w.OnClick != nil {
			x := int(int16(lParam & 0xFFFF))
			y := int(int16((lParam >> 16) & 0xFFFF))
			w.OnClick(MouseEvent{X: x, Y: y})
		}
		return 0

	case WM_RBUTTONDOWN:
		// 右键按下：记录起始位置
		var pt POINT
		getCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
		w.dragStartX = pt.X
		w.dragStartY = pt.Y
		var rect RECT
		getWindowRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&rect)))
		w.winStartX = rect.Left
		w.winStartY = rect.Top
		w.dragging = true
		w.isDrag = false // 初始化为非拖动
		setCapture.Call(uintptr(hwnd))
		return 0

	case WM_MOUSEMOVE:
		// 右键拖动检测
		if w.dragging {
			var pt POINT
			getCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
			dx := pt.X - w.dragStartX
			dy := pt.Y - w.dragStartY
			// 如果移动距离超过 5 像素，认为是拖动
			if dx*dx+dy*dy > 25 {
				w.isDrag = true
			}
			// 如果是拖动，更新位置（使用 SWP_NOREDRAW 避免重绘阻塞）
			if w.isDrag {
				newX := w.winStartX + (pt.X - w.dragStartX)
				newY := w.winStartY + (pt.Y - w.dragStartY)
				// SWP_NOREDRAW = 0x0008, SWP_NOZORDER = 0x0004, SWP_NOSIZE = 0x0001
				setWindowPos.Call(uintptr(hwnd), 0, uintptr(newX), uintptr(newY), 0, 0, uintptr(0x0008|SWP_NOZORDER|SWP_NOSIZE))
			}
		}
		return 0

	case WM_RBUTTONUP:
		// 右键松开
		if w.dragging {
			releaseCapture.Call()
			w.dragging = false
			// 如果不是拖动，则显示菜单
			if !w.isDrag {
				w.showContextMenu(hwnd)
			}
		}
		return 0

	case WM_COMMAND:
		// 处理菜单命令
		menuID := int(wParam & 0xFFFF)
		w.handleMenuCommand(menuID)
		return 0

	case WM_UPDATE_DISPLAY:
		w.updateDisplay()
		return 0
	}

	result, _, _ := defWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return result
}

func (w *windowsPlatform) updateDisplay() {
	if w.Hwnd == 0 || w.bmpPtr == nil {
		return
	}

	pixelCount := w.Width * w.Height * 4

	if w.hasImage {
		// 复制基础像素到 DIB section
		dest := unsafe.Slice((*byte)(w.bmpPtr), pixelCount)
		copy(dest, w.basePixels)

		// 在 DIB section 上绘制气泡
		if w.currentBubble != "" {
			w.drawBubbleOnPixels(dest)
		}

		// 使用 UpdateLayeredWindow 显示（支持逐像素透明）
		size := SIZE{Cx: int32(w.Width), Cy: int32(w.Height)}
		ptSrc := POINT{X: 0, Y: 0}
		blend := BLENDFUNCTION{
			BlendOp:             AC_SRC_OVER,
			BlendFlags:          0,
			SourceConstantAlpha: 255,
			AlphaFormat:         AC_SRC_ALPHA,
		}

		updateLayeredWindow.Call(
			uintptr(w.Hwnd),
			0,
			0, // 不改变位置
			uintptr(unsafe.Pointer(&size)),
			uintptr(w.memDC),
			uintptr(unsafe.Pointer(&ptSrc)),
			0,
			uintptr(unsafe.Pointer(&blend)),
			uintptr(ULW_ALPHA),
		)
	} else {
		// 无图片时使用纯色绘制
		hdc := w.getDC()
		if hdc == 0 {
			return
		}
		defer w.releaseDC(hdc)

		colors := map[string]uint32{
			"idle":     RGB(255, 182, 193),
			"jump":     RGB(135, 206, 250),
			"head_pat": RGB(255, 218, 185),
			"wave":     RGB(144, 238, 144),
			"happy":    RGB(255, 215, 0),
		}
		color := colors[w.currentAction]
		if color == 0 {
			color = colors["idle"]
		}

		brush, _, _ := createSolidBrush.Call(uintptr(color))
		if brush != 0 {
			oldBrush, _, _ := selectObject.Call(uintptr(hdc), brush)
			roundRect.Call(uintptr(hdc), 0, 0, uintptr(w.Width), uintptr(w.Height), 10, 10)
			selectObject.Call(uintptr(hdc), oldBrush)
			deleteObject.Call(brush)
		}
	}
}

func (w *windowsPlatform) drawBubbleOnPixels(pixels []byte) {
	// 计算文字尺寸和气泡大小
	text := w.currentBubble
	if len(text) == 0 {
		return
	}

	// 根据缩放比例调整文字和气泡大小
	scale := w.scale
	if scale < 0.1 {
		scale = 1.0
	}

	// 根据文字长度计算气泡尺寸（缩放后）
	charWidth := int(18 * scale)  // 中文字符宽度估算（增加间距）
	lineHeight := int(28 * scale) // 行高（增加）
	maxCharsPerLine := 10         // 每行最大字符数
	padding := int(20 * scale)    // 内边距（增加）

	// 计算需要的行数
	textLen := len([]rune(text))
	lines := (textLen + maxCharsPerLine - 1) / maxCharsPerLine
	if lines < 1 {
		lines = 1
	}

	// 计算气泡尺寸
	minBubbleW := int(100 * scale)
	bubbleW := max(len([]rune(text))*charWidth+padding*2, minBubbleW)
	maxBubbleW := int(200 * scale)
	if bubbleW > maxBubbleW {
		bubbleW = maxBubbleW
	}
	bubbleH := lines*lineHeight + padding*2

	// 气泡位置（显示在图片上方的空白区域）
	bubbleX := (w.Width - bubbleW) / 2
	if bubbleX < 0 {
		bubbleX = 2
	}
	// 气泡显示在窗口顶部（上方空白区域）
	bubbleY := 10

	// 圆角半径（缩放后）
	radius := int(12 * scale)

	// 绘制阴影（偏移根据缩放）
	shadowOffset := int(3 * scale)
	shadowColor := [4]byte{180, 180, 180, 100} // 灰色半透明阴影
	for y := bubbleY + shadowOffset; y < bubbleY+bubbleH+shadowOffset && y < w.Height; y++ {
		for x := bubbleX + shadowOffset; x < bubbleX+bubbleW+shadowOffset && x < w.Width; x++ {
			if x < 0 || y < 0 {
				continue
			}
			// 圆角判断
			dx := 0
			dy := 0
			if x-bubbleX-shadowOffset < radius {
				dx = radius - (x - bubbleX - shadowOffset)
			} else if x-bubbleX-shadowOffset >= bubbleW-radius {
				dx = (x - bubbleX - shadowOffset) - (bubbleW - radius) + 1
			}
			if y-bubbleY-shadowOffset < radius {
				dy = radius - (y - bubbleY - shadowOffset)
			} else if y-bubbleY-shadowOffset >= bubbleH-radius {
				dy = (y - bubbleY - shadowOffset) - (bubbleH - radius) + 1
			}
			if dx*dx+dy*dy <= radius*radius {
				idx := (y*w.Width + x) * 4
				pixels[idx+0] = blendChannel(pixels[idx+0], shadowColor[0], shadowColor[3])
				pixels[idx+1] = blendChannel(pixels[idx+1], shadowColor[1], shadowColor[3])
				pixels[idx+2] = blendChannel(pixels[idx+2], shadowColor[2], shadowColor[3])
				if pixels[idx+3] < shadowColor[3] {
					pixels[idx+3] = shadowColor[3]
				}
			}
		}
	}

	// 绘制气泡背景（渐变白色）
	for y := bubbleY; y < bubbleY+bubbleH && y < w.Height; y++ {
		for x := bubbleX; x < bubbleX+bubbleW && x < w.Width; x++ {
			if x < 0 || y < 0 {
				continue
			}
			// 圆角判断
			dx := 0
			dy := 0
			if x-bubbleX < radius {
				dx = radius - (x - bubbleX)
			} else if x-bubbleX >= bubbleW-radius {
				dx = (x - bubbleX) - (bubbleW - radius) + 1
			}
			if y-bubbleY < radius {
				dy = radius - (y - bubbleY)
			} else if y-bubbleY >= bubbleH-radius {
				dy = (y - bubbleY) - (bubbleH - radius) + 1
			}
			if dx*dx+dy*dy <= radius*radius {
				// 渐变效果：从上到下略微变深
				gradient := uint8(255 - (y-bubbleY)*20/bubbleH)
				idx := (y*w.Width + x) * 4
				pixels[idx+0] = blendChannel(pixels[idx+0], gradient, 240) // B
				pixels[idx+1] = blendChannel(pixels[idx+1], gradient, 240) // G
				pixels[idx+2] = blendChannel(pixels[idx+2], 255, 240)      // R (淡粉色)
				pixels[idx+3] = 255                                        // A
			}
		}
	}

	// 绘制气泡边框（淡蓝色，宽度根据缩放）
	borderWidth := max(1, int(2*scale))
	borderColor := [4]byte{100, 149, 237, 255} // 矢车菊蓝
	for y := bubbleY; y < bubbleY+bubbleH && y < w.Height; y++ {
		for x := bubbleX; x < bubbleX+bubbleW && x < w.Width; x++ {
			if x < 0 || y < 0 {
				continue
			}
			// 检查是否是边框（距离边界 borderWidth 像素内）
			isBorder := false
			// 圆角判断
			dx := 0
			dy := 0
			if x-bubbleX < radius {
				dx = radius - (x - bubbleX)
			} else if x-bubbleX >= bubbleW-radius {
				dx = (x - bubbleX) - (bubbleW - radius) + 1
			}
			if y-bubbleY < radius {
				dy = radius - (y - bubbleY)
			} else if y-bubbleY >= bubbleH-radius {
				dy = (y - bubbleY) - (bubbleH - radius) + 1
			}
			// 计算到边缘的距离
			insideRadius := dx*dx + dy*dy
			if insideRadius <= radius*radius {
				// 检查是否在边框范围内
				borderDist := min(
					min(x-bubbleX, bubbleX+bubbleW-x),
					min(y-bubbleY, bubbleY+bubbleH-y),
				)
				if borderDist >= 0 && borderDist < borderWidth {
					isBorder = true
				}
			}
			if isBorder {
				idx := (y*w.Width + x) * 4
				pixels[idx+0] = borderColor[0] // B
				pixels[idx+1] = borderColor[1] // G
				pixels[idx+2] = borderColor[2] // R
				pixels[idx+3] = borderColor[3] // A
			}
		}
	}

	// 绘制文字 - 使用 GDI 在内存 DC 上绘制
	oldBmp, _, _ := selectObject.Call(uintptr(w.memDC), w.hBitmap)
	defer selectObject.Call(uintptr(w.memDC), oldBmp)

	// 创建缩放后的字体
	fontSize := int32(24 * scale)
	if fontSize < 8 {
		fontSize = 8
	}
	fontName, _ := syscall.UTF16PtrFromString("微软雅黑")
	hFont, _, _ := createFontW.Call(
		uintptr(fontSize),                  // nHeight
		0,                                  // nWidth (0 = 自动)
		0,                                  // nEscapement
		0,                                  // nOrientation
		uintptr(FW_NORMAL),                 // fnWeight
		0,                                  // fdwItalic
		0,                                  // fdwUnderline
		0,                                  // fdwStrikeOut
		uintptr(DEFAULT_CHARSET),           // fdwCharSet
		uintptr(OUT_DEFAULT_PRECIS),        // fdwOutputPrecision
		uintptr(CLIP_DEFAULT_PRECIS),       // fdwClipPrecision
		uintptr(DEFAULT_QUALITY),           // fdwQuality
		uintptr(DEFAULT_PITCH|FF_DONTCARE), // fdwPitchAndFamily
		uintptr(unsafe.Pointer(fontName)),
	)
	var oldFont uintptr
	if hFont != 0 {
		oldFont, _, _ = selectObject.Call(uintptr(w.memDC), hFont)
		defer func() {
			selectObject.Call(uintptr(w.memDC), oldFont)
			deleteObject.Call(hFont)
		}()
	}

	// 设置透明背景模式
	setBkMode.Call(uintptr(w.memDC), 1) // TRANSPARENT
	// 文字颜色：深灰色
	setTextColor.Call(uintptr(w.memDC), uintptr(RGB(50, 50, 50)))

	// 清空文字绘制区域（防止残留字符）
	clearY1 := bubbleY
	clearY2 := bubbleY + bubbleH
	if clearY1 < 0 {
		clearY1 = 0
	}
	if clearY2 > w.Height {
		clearY2 = w.Height
	}
	for y := clearY1; y < clearY2; y++ {
		for x := bubbleX; x < bubbleX+bubbleW && x < w.Width; x++ {
			if x < 0 {
				continue
			}
			idx := (y*w.Width + x) * 4
			// 只清空文字区域（不是整个气泡）
			if pixels[idx+3] > 0 {
				// 保留背景色，只重置 alpha 以便重新绘制
			}
		}
	}

	// 分行绘制文字
	runes := []rune(text)
	lineIndex := 0
	for i := 0; i < len(runes); i += maxCharsPerLine {
		end := min(i+maxCharsPerLine, len(runes))
		line := string(runes[i:end])
		textPtr, _ := syscall.UTF16PtrFromString(line)
		textX := int32(bubbleX + padding)
		textY := int32(bubbleY + padding + lineIndex*lineHeight)
		// 使用 UTF16 字符数（不是字节数）
		textOutW.Call(uintptr(w.memDC), uintptr(textX), uintptr(textY), uintptr(unsafe.Pointer(textPtr)), uintptr(len(runes[i:end])))
		lineIndex++
	}

	// 修复 GDI 绘制文字后的 alpha 通道
	for y := bubbleY; y < bubbleY+bubbleH && y < w.Height; y++ {
		for x := bubbleX; x < bubbleX+bubbleW && x < w.Width; x++ {
			if x < 0 {
				continue
			}
			idx := (y*w.Width + x) * 4
			// 如果像素有颜色但 alpha 为 0，设置为不透明
			if pixels[idx+3] == 0 && (pixels[idx+0] != 0 || pixels[idx+1] != 0 || pixels[idx+2] != 0) {
				pixels[idx+3] = 255
			}
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func blendChannel(dst, src uint8, alpha uint8) uint8 {
	a := uint16(alpha)
	return uint8((uint16(dst)*(255-a) + uint16(src)*a) / 255)
}

func RGB(r, g, b byte) uint32 {
	return uint32(r) | uint32(g)<<8 | uint32(b)<<16
}

func (w *windowsPlatform) getDC() HDC {
	hdc, _, _ := getDC.Call(uintptr(w.Hwnd))
	return HDC(hdc)
}

func (w *windowsPlatform) releaseDC(hdc HDC) {
	releaseDC.Call(uintptr(w.Hwnd), uintptr(hdc))
}

func (w *windowsPlatform) Show() {
	if w.Hwnd == 0 {
		return
	}
	showWindow.Call(uintptr(w.Hwnd), uintptr(SW_SHOW))
	updateWindow.Call(uintptr(w.Hwnd))
	w.updateDisplay()
}

func (w *windowsPlatform) Run() {
	w.running = true
	var msg MSG
	for w.running {
		ret, _, _ := getMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(ret) == 0 {
			// WM_QUIT
			break
		}
		if int32(ret) == -1 {
			// 错误
			break
		}
		translateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		dispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func (w *windowsPlatform) Close() {
	if w.Hwnd != 0 {
		destroyWindow.Call(uintptr(w.Hwnd))
		w.Hwnd = 0
	}
	if w.hBitmap != 0 {
		deleteObject.Call(w.hBitmap)
		w.hBitmap = 0
	}
	if w.memDC != 0 {
		deleteDC.Call(w.memDC)
		w.memDC = 0
	}
}

func (w *windowsPlatform) Navigate(url string) {}

func (w *windowsPlatform) SetSize(width, height int) {
	w.Width = width
	w.Height = height
	setWindowPos.Call(uintptr(w.Hwnd), 0, 0, 0, uintptr(width), uintptr(height), uintptr(SWP_NOZORDER))
}

func (w *windowsPlatform) SetPosition(x, y int) {
	w.X = x
	w.Y = y
	setWindowPos.Call(uintptr(w.Hwnd), 0, uintptr(x), uintptr(y), 0, 0, uintptr(SWP_NOSIZE))
}

func (w *windowsPlatform) GetSize() (int, int) {
	return w.Width, w.Height
}

func (w *windowsPlatform) GetPosition() (int, int) {
	return w.X, w.Y
}

func (w *windowsPlatform) ShowNotification(message string) {}

func (w *windowsPlatform) SetOnClick(handler MouseEventHandler)      { w.OnClick = handler }
func (w *windowsPlatform) SetOnRightClick(handler MouseEventHandler) { w.OnRightClick = handler }
func (w *windowsPlatform) SetOnMouseDown(handler MouseEventHandler)  { w.OnMouseDown = handler }
func (w *windowsPlatform) SetOnMouseUp(handler MouseEventHandler)    { w.OnMouseUp = handler }
func (w *windowsPlatform) SetOnMouseMove(handler MouseEventHandler)  { w.OnMouseMove = handler }
func (w *windowsPlatform) SetOnExpressionChange(handler ExpressionChangeHandler) {
	w.OnExpressionChange = handler
}

func (w *windowsPlatform) Eval(code string) {}

func (w *windowsPlatform) UpdatePet(action string, frame int) {
	w.currentAction = action
	w.currentFrame = frame
	if w.Hwnd != 0 {
		postMessageW.Call(uintptr(w.Hwnd), uintptr(WM_UPDATE_DISPLAY), 0, 0)
	}
}

func (w *windowsPlatform) ShowBubble(text string) {
	w.currentBubble = text
	if w.Hwnd != 0 {
		postMessageW.Call(uintptr(w.Hwnd), uintptr(WM_UPDATE_DISPLAY), 0, 0)
	}
}

func (w *windowsPlatform) HideBubble() {
	w.currentBubble = ""
	if w.Hwnd != 0 {
		postMessageW.Call(uintptr(w.Hwnd), uintptr(WM_UPDATE_DISPLAY), 0, 0)
	}
}

// 切换到下一张图片
func (w *windowsPlatform) SwitchToNextImage() string {
	if len(w.imageOrder) == 0 {
		return ""
	}
	w.currentIdx = (w.currentIdx + 1) % len(w.imageOrder)
	w.switchToImage(w.currentIdx)
	return w.imageOrder[w.currentIdx]
}

// 获取当前图片名称
func (w *windowsPlatform) GetCurrentImageName() string {
	if len(w.imageOrder) == 0 || w.currentIdx >= len(w.imageOrder) {
		return ""
	}
	return w.imageOrder[w.currentIdx]
}

func (w *windowsPlatform) SetDragging(dragging bool) {}

// 显示右键菜单
func (w *windowsPlatform) showContextMenu(hwnd HWND) {
	// 创建主菜单
	hMenu, _, _ := createPopupMenu.Call()
	if hMenu == 0 {
		return
	}

	// 创建"调整大小"子菜单
	hSizeMenu, _, _ := createPopupMenu.Call()
	if hSizeMenu != 0 {
		increasePtr, _ := syscall.UTF16PtrFromString("  🔍 放大 (+)  ")
		decreasePtr, _ := syscall.UTF16PtrFromString("  🔍 缩小 (-)  ")
		appendMenuW.Call(hSizeMenu, MF_STRING, uintptr(ID_SIZE_INCREASE), uintptr(unsafe.Pointer(increasePtr)))
		appendMenuW.Call(hSizeMenu, MF_STRING, uintptr(ID_SIZE_DECREASE), uintptr(unsafe.Pointer(decreasePtr)))

		sizePtr, _ := syscall.UTF16PtrFromString("📐 调整大小")
		appendMenuW.Call(hMenu, MF_STRING|MF_POPUP, hSizeMenu, uintptr(unsafe.Pointer(sizePtr)))
	}

	// 添加分隔线
	appendMenuW.Call(hMenu, MF_SEPARATOR, 0, 0)

	// 创建"表情"子菜单
	hExprMenu, _, _ := createPopupMenu.Call()
	if hExprMenu != 0 {
		for i, name := range w.imageOrder {
			// 添加空格增加间距
			displayName := "  " + name + "  "
			namePtr, _ := syscall.UTF16PtrFromString(displayName)
			appendMenuW.Call(hExprMenu, MF_STRING, uintptr(ID_EXPRESSION_BASE+i), uintptr(unsafe.Pointer(namePtr)))
		}

		exprPtr, _ := syscall.UTF16PtrFromString("😊 表情")
		appendMenuW.Call(hMenu, MF_STRING|MF_POPUP, hExprMenu, uintptr(unsafe.Pointer(exprPtr)))
	}

	// 添加分隔线
	appendMenuW.Call(hMenu, MF_SEPARATOR, 0, 0)

	// 添加退出按钮
	exitPtr, _ := syscall.UTF16PtrFromString("❌ 退出")
	appendMenuW.Call(hMenu, MF_STRING, uintptr(ID_EXIT), uintptr(unsafe.Pointer(exitPtr)))

	// 获取鼠标位置
	var pt POINT
	getCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

	// 显示菜单
	// TrackPopupMenu 参数: flags, x, y, reserved, hwnd, prcRect
	// TPM_LEFTALIGN | TPM_LEFTBUTTON = 0x0000 | 0x0000
	trackPopupMenu.Call(
		hMenu,
		uintptr(TPM_LEFTALIGN),
		uintptr(pt.X),
		uintptr(pt.Y),
		0,
		uintptr(hwnd),
		0,
	)

	// 销毁菜单
	destroyMenu.Call(hMenu)
}

// 处理菜单命令
func (w *windowsPlatform) handleMenuCommand(menuID int) {
	switch {
	case menuID == ID_SIZE_INCREASE:
		// 放大
		w.scale *= 1.2
		if w.scale > 3.0 {
			w.scale = 3.0
		}
		w.resizeImage()

	case menuID == ID_SIZE_DECREASE:
		// 缩小
		w.scale /= 1.2
		if w.scale < 0.3 {
			w.scale = 0.3
		}
		w.resizeImage()

	case menuID == ID_EXIT:
		// 退出程序
		w.running = false
		postQuitMessage.Call(0)

	case menuID >= ID_EXPRESSION_BASE && menuID < ID_EXPRESSION_BASE+len(w.imageOrder):
		// 切换到指定表情
		idx := menuID - ID_EXPRESSION_BASE
		name := w.imageOrder[idx]
		w.switchToImage(idx)
		// 触发回调，更新气泡文字
		if w.OnExpressionChange != nil {
			w.OnExpressionChange(name)
		}
	}
}

// 调整图片大小
func (w *windowsPlatform) resizeImage() {
	if len(w.imageOrder) == 0 {
		return
	}

	// 基础尺寸
	baseWidth := 200
	baseImageHeight := 300
	baseBubbleHeight := 80 // 气泡区域高度

	// 计算新尺寸
	newWidth := int(float64(baseWidth) * w.scale)
	newImageHeight := int(float64(baseImageHeight) * w.scale)
	newBubbleHeight := int(float64(baseBubbleHeight) * w.scale)
	newTotalHeight := newImageHeight + newBubbleHeight

	// 重新加载当前图片
	currentName := w.imageOrder[w.currentIdx]
	pixels, err := w.loadImageByName(currentName, newWidth, newImageHeight, newBubbleHeight)
	if err != nil {
		fmt.Printf("调整大小失败: %v\n", err)
		return
	}

	// 更新尺寸
	w.Width = newWidth
	w.Height = newTotalHeight
	w.imageWidth = newWidth
	w.imageHeight = newImageHeight
	w.basePixels = pixels

	// 重新创建 DIB section
	if w.hBitmap != 0 {
		deleteObject.Call(w.hBitmap)
		w.hBitmap = 0
	}
	if w.memDC != 0 {
		deleteDC.Call(w.memDC)
		w.memDC = 0
	}
	w.createDIBSection()

	// 更新窗口大小
	setWindowPos.Call(uintptr(w.Hwnd), 0, 0, 0, uintptr(newWidth), uintptr(newTotalHeight), uintptr(SWP_NOZORDER|0x0002)) // SWP_NOMOVE

	// 更新显示
	postMessageW.Call(uintptr(w.Hwnd), uintptr(WM_UPDATE_DISPLAY), 0, 0)
}

// 切换到指定图片
func (w *windowsPlatform) switchToImage(idx int) {
	if idx < 0 || idx >= len(w.imageOrder) {
		return
	}

	w.currentIdx = idx
	name := w.imageOrder[idx]

	// 计算当前气泡高度
	bubbleHeight := w.Height - w.imageHeight

	// 加载当前尺寸的图片（使用图片高度，不是窗口总高度）
	pixels, err := w.loadImageByName(name, w.imageWidth, w.imageHeight, bubbleHeight)
	if err != nil {
		fmt.Printf("切换图片失败: %v\n", err)
		return
	}

	w.basePixels = pixels
	w.hasImage = true

	// 更新显示
	postMessageW.Call(uintptr(w.Hwnd), uintptr(WM_UPDATE_DISPLAY), 0, 0)
}
