//go:build darwin

package platform

/*
#cgo CFLAGS: -fobjc-arc -Wno-unused-variable -Wno-unused-function
#cgo LDFLAGS: -framework Cocoa -framework CoreGraphics

#include <stdlib.h>
#include <stdint.h>

// 桥接函数定义在 darwin_bridge.m 中（单一翻译单元，避免 ObjC 类符号重复）。
// 此处仅声明，供 Go 通过 cgo 调用，链接时解析到 .m 中的定义。
//
// 注意：C 侧只持有整数句柄（intptr_t），不持有任何 Go 指针，以符合 cgo 指针规则。

extern void setupApp(void);
extern int  getScreenHeight(void);
extern void *createPetWindow(intptr_t handle, int x, int y, int w, int h, void **outView);
extern void *createMenuTarget(intptr_t handle);
extern void winOrderFront(void *win);
extern void winSetFrame(void *win, int x, int y, int w, int h);
extern void winGetFrame(void *win, int *x, int *y, int *w, int *h);
extern void setViewImage(void *viewPtr, const void *rgba, int w, int h, int dx, int dy, int dw, int dh);
extern void setViewBubble(void *viewPtr, const char *text, int bx, int by, int bw, int bh, int th, int rad, int pad, int lh, int fs);
extern void clearViewBubble(void *viewPtr);
extern void viewRedraw(void *viewPtr);
extern void setViewAlpha(void *viewPtr, double alpha);
extern void *menuCreate(void);
extern void menuAddItem(void *menu, const char *title, int tag, int isSeparator, int isSubmenu, void *submenu, int checked, void *target);
extern void menuPopup(void *menu);
extern void runApp(void);
extern void stopApp(void);
*/
import "C"

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"
)

// 菜单 ID - 与 Windows 版本保持一致
const (
	dID_SIZE_NEG6       = 1010
	dID_SIZE_NEG4       = 1011
	dID_SIZE_NEG2       = 1012
	dID_SIZE_0          = 1013
	dID_SIZE_POS2       = 1014
	dID_SIZE_POS4       = 1015
	dID_SIZE_POS6       = 1016
	dID_EXIT            = 1003
	dID_EXPRESSION_BASE = 2000

	// 字体大小档位（macOS 新增，与 Windows 一致的结构）
	dID_FONT_NEG6 = 1020
	dID_FONT_NEG4 = 1021
	dID_FONT_NEG2 = 1022
	dID_FONT_0    = 1023
	dID_FONT_POS2 = 1024
	dID_FONT_POS4 = 1025
	dID_FONT_POS6 = 1026

	// 透明度档位（0=不透明，数值越大越透明）
	dID_ALPHA_0  = 1030
	dID_ALPHA_20 = 1031
	dID_ALPHA_40 = 1032
	dID_ALPHA_60 = 1033
	dID_ALPHA_80 = 1034
)

type darwinPlatform struct {
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

	running       bool
	currentAction string
	currentFrame  int
	currentBubble string

	// 图片数据
	images       map[string]image.Image // 图片名称 -> 原始解码图
	imageOrder   []string               // 图片切换顺序
	currentIdx   int                    // 当前图片索引
	imageWidth   int                    // 图片绘制区宽度（随缩放）
	imageHeight  int                    // 图片绘制区高度（随缩放）
	bubbleHeight int                    // 气泡区域高度（固定，不随缩放）
	// 缩放
	scale      float64 // 当前缩放比例
	sizeLevel  int     // 当前大小档位
	fontLevel  int     // 字体大小档位（-4..+4，对应 20..28pt）
	alphaLevel int     // 透明度档位（0=不透明，0/20/40/60/80）

	// 原生句柄
	nsWindow     unsafe.Pointer // NSWindow *
	nsView       unsafe.Pointer // PetView *
	menuTarget   unsafe.Pointer // MenuTarget *
	screenHeight int

	mu sync.Mutex
}

func (p *darwinPlatform) Init() error {
	C.setupApp()
	p.screenHeight = int(C.getScreenHeight())
	return nil
}

func (p *darwinPlatform) CreateWindow(title string, x, y, width, height int) error {
	p.X = x
	p.Y = y

	// 尝试加载所有宠物图片（与 Windows 版本一致）
	if err := p.loadAllImages(); err == nil {
		width = p.Width
		height = p.Height
		fmt.Printf("已加载宠物图片，尺寸: %dx%d，共 %d 张\n", width, height, len(p.images))
	} else {
		p.Width = width
		p.Height = height
		fmt.Printf("未找到宠物图片，使用默认尺寸: %dx%d\n", width, height)
	}
	p.X = x
	p.Y = y

	// 屏幕坐标系转换：Windows 用左上角原点，macOS 用左下角原点
	// 这里将逻辑的(左上角)转换为 macOS 的(左下角)
	macY := p.screenHeight - y - height
	if macY < 0 {
		macY = 0
	}

	var viewPtr unsafe.Pointer
	handle := petRegister(p)
	p.nsWindow = unsafe.Pointer(C.createPetWindow(C.intptr_t(handle), C.int(x), C.int(macY), C.int(width), C.int(height), &viewPtr))
	p.nsView = viewPtr
	p.menuTarget = unsafe.Pointer(C.createMenuTarget(C.intptr_t(handle)))

	// 推送初始图片
	p.pushImage()
	return nil
}

// 加载所有图片（与 Windows 版本一致：图片列表顺序、基准尺寸、缩放档位）
func (p *darwinPlatform) loadAllImages() error {
	imageNames := []string{
		"主体", "挥手", "奔跑", "跳跃", "打哈欠",
		"害羞", "生气", "委屈", "震惊", "喝饮料", "睡觉",
	}

	p.images = make(map[string]image.Image)
	p.imageOrder = imageNames
	p.currentIdx = 0
	// 默认缩放：档位0对应的 scale
	p.scale = dScaleForLevel(0)
	p.sizeLevel = 0

	// 基础尺寸（scale=1.0 时的尺寸）
	baseWidth := 200
	baseImageHeight := 300
	bubbleHeight := 100 // 气泡区域高度固定，不随缩放变化（保证字体不变）

	windowWidth := baseWidth
	imageDrawWidth := int(float64(baseWidth) * p.scale)
	imageDrawHeight := int(float64(baseImageHeight) * p.scale)

	for _, name := range imageNames {
		img, err := p.loadImageByName(name)
		if err != nil {
			fmt.Printf("加载图片失败: %s, %v\n", name, err)
			continue
		}
		p.images[name] = img
	}

	if len(p.images) == 0 {
		return fmt.Errorf("未找到任何宠物图片")
	}

	// 设置尺寸和当前图片（窗口高度 = 图片高度 + 气泡高度）
	p.Width = windowWidth
	p.Height = imageDrawHeight + bubbleHeight
	p.imageWidth = imageDrawWidth
	p.imageHeight = imageDrawHeight
	p.bubbleHeight = bubbleHeight
	return nil
}

// 加载单个图片（与 Windows 版本路径查找逻辑一致，额外支持 .app bundle 内的资源路径）
func (p *darwinPlatform) loadImageByName(name string) (image.Image, error) {
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	paths := []string{
		filepath.Join("assets", name+".png"),
		filepath.Join(exeDir, "assets", name+".png"),
		filepath.Join(exeDir, name+".png"),
		// .app bundle 内：可执行文件位于 Contents/MacOS/，资源位于 Contents/Resources/
		filepath.Join(exeDir, "..", "Resources", "assets", name+".png"),
	}

	for _, pth := range paths {
		f, err := os.Open(pth)
		if err != nil {
			continue
		}
		img, _ := png.Decode(f)
		f.Close()
		if img != nil {
			return img, nil
		}
	}
	return nil, fmt.Errorf("未找到图片: %s", name)
}

// 将解码图按保持比例缩放到 drawW×drawH 区域，返回预乘 RGBA 像素及实际绘制尺寸
func scaleToRGBA(img image.Image, drawW, drawH int) ([]byte, int, int) {
	bounds := img.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()
	scaleW := float64(drawW) / float64(origW)
	scaleH := float64(drawH) / float64(origH)
	scale := scaleW
	if scaleH < scaleW {
		scale = scaleH
	}
	w := int(float64(origW) * scale)
	h := int(float64(origH) * scale)
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	rgba := make([]byte, w*h*4)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			sx := int(float64(x) / scale)
			// 上下翻转：让输出缓冲区的第 0 行对应源图底部。
			// CGImage/NSImage 按 CG 的上下原点解释数据，不翻转会导致人物上下颠倒。
			srcY := h - 1 - y
			if srcY < 0 {
				srcY = 0
			}
			sy := int(float64(srcY) / scale)
			if sx >= origW {
				sx = origW - 1
			}
			if sy >= origH {
				sy = origH - 1
			}
			r32, g32, b32, a32 := img.At(sx, sy).RGBA()
			a := uint8(a32 >> 8)
			// 预乘 alpha（与 CGImage 的 kCGImageAlphaPremultipliedLast 对应）
			rgba[(y*w+x)*4+0] = uint8(uint32(r32>>8) * uint32(a) / 255)
			rgba[(y*w+x)*4+1] = uint8(uint32(g32>>8) * uint32(a) / 255)
			rgba[(y*w+x)*4+2] = uint8(uint32(b32>>8) * uint32(a) / 255)
			rgba[(y*w+x)*4+3] = a
		}
	}
	return rgba, w, h
}

// 推送当前图片到视图
func (p *darwinPlatform) pushImage() {
	if p.nsView == nil || len(p.imageOrder) == 0 {
		return
	}
	p.mu.Lock()
	img, ok := p.images[p.imageOrder[p.currentIdx]]
	p.mu.Unlock()
	if !ok || img == nil {
		return
	}
	rgba, w, h := scaleToRGBA(img, p.imageWidth, p.imageHeight)
	// 图片在窗口中居中：水平居中，垂直在图片带内居中（图片带位于气泡区域下方）
	drawX := (p.Width - w) / 2
	drawY := p.bubbleHeight + (p.imageHeight-h)/2
	if drawX < 0 {
		drawX = 0
	}
	C.setViewImage(p.nsView, unsafe.Pointer(&rgba[0]),
		C.int(w), C.int(h), C.int(drawX), C.int(drawY), C.int(w), C.int(h))
}

// 计算气泡布局（尺寸/换行逻辑参考 Windows 版本 drawBubbleOnPixels，
// 并针对 macOS 做了符号断句优化，使换行落在自然位置）
func (p *darwinPlatform) computeBubbleLayout() (text string, bx, by, bw, bh, th, rad, pad, lh, fs int) {
	text = p.currentBubble
	if text == "" {
		return
	}
	// 字体大小随档位变化（基准 24pt，档位 -6..+6 -> 12..48pt）
	fontSize := dFontSizeForLevel(p.fontLevel)
	// 气泡尺寸随字体等比缩放，保证文字增大时气泡自适应
	fontScale := float64(fontSize) / 24.0
	// macOS 24pt 系统字体下，CJK 字符实际宽度约 24px，这里取略大的估算值，
	// 使换行更保守，保证文字左右两侧都有空白边距、不贴边。
	charWidth := int(26 * fontScale)
	lineHeight := int(28 * fontScale) // 行高
	padding := int(20 * fontScale)    // 左右内边距
	minBubbleW := int(100 * fontScale)
	maxBubbleW := int(200 * fontScale)

	// 先按中间符号断句（~、…、!、?、。， 等），让符号后的内容另起一行，
	// 避免纯字符计数硬换行把词组拆开，造成丑陋断行。
	segments := splitBySymbols(text)

	// 以最长分段作为气泡宽度的依据
	longest := 0
	for _, seg := range segments {
		if n := len([]rune(seg)); n > longest {
			longest = n
		}
	}
	bubbleW := maxInt(longest*charWidth+padding*2, minBubbleW)
	if bubbleW > maxBubbleW {
		bubbleW = maxBubbleW
	}

	availTextW := bubbleW - padding*2
	if availTextW < charWidth {
		availTextW = charWidth
	}
	maxCharsPerLine := availTextW / charWidth
	if maxCharsPerLine < 1 {
		maxCharsPerLine = 1
	}

	// 对仍超长的分段按字符计数二次换行
	var lines []string
	for _, seg := range segments {
		rs := []rune(seg)
		if len(rs) <= maxCharsPerLine {
			lines = append(lines, seg)
			continue
		}
		for i := 0; i < len(rs); i += maxCharsPerLine {
			end := minInt(i+maxCharsPerLine, len(rs))
			lines = append(lines, string(rs[i:end]))
		}
	}
	if len(lines) == 0 {
		lines = []string{""}
	}
	bubbleH := len(lines)*lineHeight + padding*2

	bubbleX := (p.Width - bubbleW) / 2
	if bubbleX < 0 {
		bubbleX = 2
	}
	bubbleY := 10
	radius := int(16 * fontScale)
	tailH := int(10 * fontScale)

	return strings.Join(lines, "\n"), bubbleX, bubbleY, bubbleW, bubbleH, tailH, radius, padding, lineHeight, fontSize
}

// 按中间符号断句：在符号串（~、～、!、！、?、？、.、…、。、，、, 等）之后
// 插入换行，使后续内容另起一行展示。
func splitBySymbols(text string) []string {
	runes := []rune(text)
	var segments []string
	var cur strings.Builder
	i := 0
	for i < len(runes) {
		if isBubbleSymbol(runes[i]) {
			// 把整段连续符号并入当前分段
			for i < len(runes) && isBubbleSymbol(runes[i]) {
				cur.WriteRune(runes[i])
				i++
			}
			// 若符号串后还有内容，则在此断句
			if i < len(runes) {
				segments = append(segments, cur.String())
				cur.Reset()
			}
		} else {
			cur.WriteRune(runes[i])
			i++
		}
	}
	if cur.Len() > 0 {
		segments = append(segments, cur.String())
	}
	if len(segments) == 0 {
		segments = []string{text}
	}
	return segments
}

func isBubbleSymbol(r rune) bool {
	switch r {
	case '~', '～', '!', '！', '?', '？', '.', '…', '。', '，', ',':
		return true
	}
	return false
}

// 推送气泡到视图
func (p *darwinPlatform) pushBubble() {
	if p.nsView == nil {
		return
	}
	if p.currentBubble == "" {
		C.clearViewBubble(p.nsView)
		return
	}
	text, bx, by, bw, bh, th, rad, pad, lh, fs := p.computeBubbleLayout()
	// 复制字符串到 C 内存（CFString 会拷贝，但 C.CString 需要保证调用期间有效）
	ctext := C.CString(text)
	C.setViewBubble(p.nsView, ctext,
		C.int(bx), C.int(by), C.int(bw), C.int(bh),
		C.int(th), C.int(rad), C.int(pad), C.int(lh), C.int(fs))
	C.free(unsafe.Pointer(ctext))
}

func (p *darwinPlatform) Show() {
	if p.nsWindow == nil {
		return
	}
	C.winOrderFront(p.nsWindow)
	p.pushImage()
	p.pushBubble()
}

func (p *darwinPlatform) Run() {
	p.running = true
	C.runApp()
}

func (p *darwinPlatform) Close() {
	if p.nsWindow != nil {
		C.stopApp()
	}
}

func (p *darwinPlatform) Navigate(url string) {}

func (p *darwinPlatform) SetSize(width, height int) {
	p.Width = width
	p.Height = height
	if p.nsWindow != nil {
		macY := p.screenHeight - p.Y - height
		if macY < 0 {
			macY = 0
		}
		C.winSetFrame(p.nsWindow, C.int(p.X), C.int(macY), C.int(width), C.int(height))
	}
}

func (p *darwinPlatform) SetPosition(x, y int) {
	p.X = x
	p.Y = y
	if p.nsWindow != nil {
		macY := p.screenHeight - y - p.Height
		if macY < 0 {
			macY = 0
		}
		C.winSetFrame(p.nsWindow, C.int(x), C.int(macY), C.int(p.Width), C.int(p.Height))
	}
}

func (p *darwinPlatform) GetSize() (int, int) {
	return p.Width, p.Height
}

func (p *darwinPlatform) GetPosition() (int, int) {
	return p.X, p.Y
}

func (p *darwinPlatform) ShowNotification(message string) {}

func (p *darwinPlatform) SetOnClick(handler MouseEventHandler)      { p.OnClick = handler }
func (p *darwinPlatform) SetOnRightClick(handler MouseEventHandler) { p.OnRightClick = handler }
func (p *darwinPlatform) SetOnMouseDown(handler MouseEventHandler)  { p.OnMouseDown = handler }
func (p *darwinPlatform) SetOnMouseUp(handler MouseEventHandler)    { p.OnMouseUp = handler }
func (p *darwinPlatform) SetOnMouseMove(handler MouseEventHandler)  { p.OnMouseMove = handler }
func (p *darwinPlatform) SetOnExpressionChange(handler ExpressionChangeHandler) {
	p.OnExpressionChange = handler
}

func (p *darwinPlatform) Eval(code string) {}

func (p *darwinPlatform) UpdatePet(action string, frame int) {
	p.currentAction = action
	p.currentFrame = frame
	if p.nsView != nil {
		C.viewRedraw(p.nsView)
	}
}

func (p *darwinPlatform) ShowBubble(text string) {
	p.currentBubble = text
	p.pushBubble()
}

func (p *darwinPlatform) HideBubble() {
	p.currentBubble = ""
	p.pushBubble()
}

func (p *darwinPlatform) SetDragging(dragging bool) {}

// 切换到下一张图片
func (p *darwinPlatform) SwitchToNextImage() string {
	if len(p.imageOrder) == 0 {
		return ""
	}
	p.mu.Lock()
	p.currentIdx = (p.currentIdx + 1) % len(p.imageOrder)
	p.mu.Unlock()
	p.pushImage()
	return p.GetCurrentImageName()
}

// 获取当前图片名称
func (p *darwinPlatform) GetCurrentImageName() string {
	if len(p.imageOrder) == 0 || p.currentIdx >= len(p.imageOrder) {
		return ""
	}
	return p.imageOrder[p.currentIdx]
}

// 切换到指定图片（索引）
func (p *darwinPlatform) switchToImage(idx int) {
	if idx < 0 || idx >= len(p.imageOrder) {
		return
	}
	p.mu.Lock()
	p.currentIdx = idx
	p.mu.Unlock()
	p.pushImage()
}

// 显示右键菜单（与 Windows 版本的菜单结构一致）
func (p *darwinPlatform) showContextMenu() {
	if p.menuTarget == nil {
		return
	}
	menu := unsafe.Pointer(C.menuCreate())

	// "调整大小"子菜单（档位选择：-6/-4/-2/0/+2/+4/+6）
	sizeMenu := unsafe.Pointer(C.menuCreate())
	sizeLevels := []struct {
		id    int
		level int
		label string
	}{
		{dID_SIZE_NEG6, -6, "  🔍 -6  "},
		{dID_SIZE_NEG4, -4, "  🔍 -4  "},
		{dID_SIZE_NEG2, -2, "  🔍 -2  "},
		{dID_SIZE_0, 0, "  📌 0   "},
		{dID_SIZE_POS2, 2, "  🔍 +2  "},
		{dID_SIZE_POS4, 4, "  🔍 +4  "},
		{dID_SIZE_POS6, 6, "  🔍 +6  "},
	}
	for _, lv := range sizeLevels {
		checked := 0
		if p.sizeLevel == lv.level {
			checked = 1
		}
		clabel := C.CString(lv.label)
		C.menuAddItem(sizeMenu, clabel, C.int(lv.id), 0, 0, nil, C.int(checked), p.menuTarget)
		C.free(unsafe.Pointer(clabel))
	}
	csizeTitle := C.CString("📐 调整大小")
	C.menuAddItem(menu, csizeTitle, 0, 0, 1, sizeMenu, 0, p.menuTarget)
	C.free(unsafe.Pointer(csizeTitle))

	// "字体大小"子菜单（档位选择：-6/-4/-2/0/+2/+4/+6）
	fontMenu := unsafe.Pointer(C.menuCreate())
	fontLevels := []struct {
		id    int
		level int
		label string
	}{
		{dID_FONT_NEG6, -6, "  🔤 -6 (12pt)  "},
		{dID_FONT_NEG4, -4, "  🔤 -4 (15pt)  "},
		{dID_FONT_NEG2, -2, "  🔤 -2 (19pt)  "},
		{dID_FONT_0, 0, "  📌 0 (24pt)   "},
		{dID_FONT_POS2, 2, "  🔤 +2 (30pt)  "},
		{dID_FONT_POS4, 4, "  🔤 +4 (38pt)  "},
		{dID_FONT_POS6, 6, "  🔤 +6 (48pt)  "},
	}
	for _, lv := range fontLevels {
		checked := 0
		if p.fontLevel == lv.level {
			checked = 1
		}
		clabel := C.CString(lv.label)
		C.menuAddItem(fontMenu, clabel, C.int(lv.id), 0, 0, nil, C.int(checked), p.menuTarget)
		C.free(unsafe.Pointer(clabel))
	}
	cfontTitle := C.CString("🔤 字体大小")
	C.menuAddItem(menu, cfontTitle, 0, 0, 1, fontMenu, 0, p.menuTarget)
	C.free(unsafe.Pointer(cfontTitle))

	// "透明度"子菜单（0=不透明，数值越大越透明）
	alphaMenu := unsafe.Pointer(C.menuCreate())
	alphaLevels := []struct {
		id    int
		level int
		label string
	}{
		{dID_ALPHA_0, 0, "  📌 不透明  "},
		{dID_ALPHA_20, 20, "  👻 20%  "},
		{dID_ALPHA_40, 40, "  👻 40%  "},
		{dID_ALPHA_60, 60, "  👻 60%  "},
		{dID_ALPHA_80, 80, "  👻 80%  "},
	}
	for _, lv := range alphaLevels {
		checked := 0
		if p.alphaLevel == lv.level {
			checked = 1
		}
		clabel := C.CString(lv.label)
		C.menuAddItem(alphaMenu, clabel, C.int(lv.id), 0, 0, nil, C.int(checked), p.menuTarget)
		C.free(unsafe.Pointer(clabel))
	}
	calphaTitle := C.CString("👻 透明度")
	C.menuAddItem(menu, calphaTitle, 0, 0, 1, alphaMenu, 0, p.menuTarget)
	C.free(unsafe.Pointer(calphaTitle))

	// 分隔线
	C.menuAddItem(menu, nil, 0, 1, 0, nil, 0, nil)

	// "表情"子菜单
	exprMenu := unsafe.Pointer(C.menuCreate())
	for i, name := range p.imageOrder {
		displayName := "  " + name + "  "
		cname := C.CString(displayName)
		C.menuAddItem(exprMenu, cname, C.int(dID_EXPRESSION_BASE+i), 0, 0, nil, 0, p.menuTarget)
		C.free(unsafe.Pointer(cname))
	}
	cexprTitle := C.CString("😊 表情")
	C.menuAddItem(menu, cexprTitle, 0, 0, 1, exprMenu, 0, p.menuTarget)
	C.free(unsafe.Pointer(cexprTitle))

	// 分隔线
	C.menuAddItem(menu, nil, 0, 1, 0, nil, 0, nil)

	// 退出按钮
	cexitTitle := C.CString("❌ 退出")
	C.menuAddItem(menu, cexitTitle, C.int(dID_EXIT), 0, 0, nil, 0, p.menuTarget)
	C.free(unsafe.Pointer(cexitTitle))

	C.menuPopup(menu)
}

// 处理菜单命令（与 Windows 版本的 handleMenuCommand 逻辑一致）
func (p *darwinPlatform) handleMenuCommand(menuID int) {
	switch {
	case menuID >= dID_SIZE_NEG6 && menuID <= dID_SIZE_POS6:
		var level int
		switch menuID {
		case dID_SIZE_NEG6:
			level = -6
		case dID_SIZE_NEG4:
			level = -4
		case dID_SIZE_NEG2:
			level = -2
		case dID_SIZE_0:
			level = 0
		case dID_SIZE_POS2:
			level = 2
		case dID_SIZE_POS4:
			level = 4
		case dID_SIZE_POS6:
			level = 6
		}
		p.scale = dScaleForLevel(level)
		p.sizeLevel = level
		p.resizeImage()

	case menuID >= dID_FONT_NEG6 && menuID <= dID_FONT_POS6:
		var level int
		switch menuID {
		case dID_FONT_NEG6:
			level = -6
		case dID_FONT_NEG4:
			level = -4
		case dID_FONT_NEG2:
			level = -2
		case dID_FONT_0:
			level = 0
		case dID_FONT_POS2:
			level = 2
		case dID_FONT_POS4:
			level = 4
		case dID_FONT_POS6:
			level = 6
		}
		p.fontLevel = level
		// 字体变化后重新布局：气泡区域高度、气泡尺寸都会随字体自适应
		p.resizeImage()

	case menuID >= dID_ALPHA_0 && menuID <= dID_ALPHA_80:
		var lvl int
		switch menuID {
		case dID_ALPHA_0:
			lvl = 0
		case dID_ALPHA_20:
			lvl = 20
		case dID_ALPHA_40:
			lvl = 40
		case dID_ALPHA_60:
			lvl = 60
		case dID_ALPHA_80:
			lvl = 80
		}
		p.alphaLevel = lvl
		// 透明度作用于整个视图（人物 + 文字一起），0=不透明
		alpha := 1.0 - float64(lvl)/100.0
		if p.nsView != nil {
			C.setViewAlpha(p.nsView, C.double(alpha))
		}

	case menuID == dID_EXIT:
		p.running = false
		C.stopApp()

	case menuID >= dID_EXPRESSION_BASE && menuID < dID_EXPRESSION_BASE+len(p.imageOrder):
		idx := menuID - dID_EXPRESSION_BASE
		name := p.imageOrder[idx]
		p.switchToImage(idx)
		if p.OnExpressionChange != nil {
			p.OnExpressionChange(name)
		}
	}
}

// 根据档位计算缩放比例（与 Windows 版本 scaleForLevel 一致）
func dScaleForLevel(level int) float64 {
	baseScale := 1.0 / (1.2 * 1.2 * 1.2 * 1.2 * 1.2) // 档位0的基准
	factor := 1.0
	abs := level
	if abs < 0 {
		abs = -abs
	}
	for i := 0; i < abs; i++ {
		factor *= 1.2
	}
	if level >= 0 {
		return baseScale * factor
	}
	return baseScale / factor
}

// 根据档位计算字体大小（基准 24pt，每两档约 1.26x 递增，差异明显）
// 档位：-6/-4/-2/0/+2/+4/+6 -> 12/15/19/24/30/38/48pt
func dFontSizeForLevel(level int) int {
	switch level {
	case -6:
		return 12
	case -4:
		return 15
	case -2:
		return 19
	case 0:
		return 24
	case 2:
		return 30
	case 4:
		return 38
	case 6:
		return 48
	}
	if level < -6 {
		return 12
	}
	return 48
}

// 调整图片大小（与 Windows 版本 resizeImage 逻辑一致）
func (p *darwinPlatform) resizeImage() {
	if len(p.imageOrder) == 0 {
		return
	}

	baseWidth := 200
	baseImageHeight := 300
	baseBubbleHeight := 100 // 气泡区域基准高度
	// 气泡区域高度随字体档位等比缩放，字体变大时给气泡留出更多空间
	fontScale := float64(dFontSizeForLevel(p.fontLevel)) / 24.0
	if fontScale < 0.5 {
		fontScale = 0.5
	}

	newWidth := baseWidth
	newImageHeight := int(float64(baseImageHeight) * p.scale)
	newBubbleHeight := int(float64(baseBubbleHeight) * fontScale)
	newTotalHeight := newImageHeight + newBubbleHeight
	imageDrawWidth := int(float64(baseWidth) * p.scale)
	imageDrawHeight := newImageHeight

	p.Width = newWidth
	p.Height = newTotalHeight
	p.imageWidth = imageDrawWidth
	p.imageHeight = imageDrawHeight
	p.bubbleHeight = newBubbleHeight

	// 更新窗口大小与位置（保持左上角位置不变）
	if p.nsWindow != nil {
		macY := p.screenHeight - p.Y - newTotalHeight
		if macY < 0 {
			macY = 0
		}
		C.winSetFrame(p.nsWindow, C.int(p.X), C.int(macY), C.int(newWidth), C.int(newTotalHeight))
	}

	// 重新加载当前尺寸的图片并刷新显示
	p.pushImage()
	p.pushBubble()
}

// ===== 句柄注册表 =====
// C 侧只持有整数句柄，避免把含 Go 指针的结构体传给 C（cgo 指针规则）。
var (
	petRegMu   sync.Mutex
	petReg     = make(map[int]*darwinPlatform)
	petRegNext int
)

func petRegister(p *darwinPlatform) int {
	petRegMu.Lock()
	defer petRegMu.Unlock()
	petRegNext++
	petReg[petRegNext] = p
	return petRegNext
}

func petLookup(handle int) *darwinPlatform {
	petRegMu.Lock()
	defer petRegMu.Unlock()
	return petReg[handle]
}

// ===== 由 Objective-C 回调的导出函数 =====

//export petLeftClick
func petLeftClick(handle C.intptr_t, x, y C.int) {
	p := petLookup(int(handle))
	if p == nil || p.OnClick == nil {
		return
	}
	p.OnClick(MouseEvent{X: int(x), Y: int(y)})
}

//export petRightMouseDown
func petRightMouseDown(handle C.intptr_t, x, y, absX, absY C.int) {
	p := petLookup(int(handle))
	if p == nil || p.OnMouseDown == nil {
		return
	}
	p.OnMouseDown(MouseEvent{X: int(x), Y: int(y)})
}

//export petRightMouseDragged
func petRightMouseDragged(handle C.intptr_t, dx, dy C.int) {
	p := petLookup(int(handle))
	if p == nil || p.OnMouseMove == nil {
		return
	}
	p.OnMouseMove(MouseEvent{X: int(dx), Y: int(dy)})
}

//export petRightMouseUp
func petRightMouseUp(handle C.intptr_t, wasDrag C.int) {
	p := petLookup(int(handle))
	if p == nil {
		return
	}
	if wasDrag == 1 {
		// 拖拽结束：从原生窗口帧同步逻辑坐标(左上角原点)
		if p.nsWindow != nil {
			var mx, my, mw, mh C.int
			C.winGetFrame(p.nsWindow, &mx, &my, &mw, &mh)
			p.X = int(mx)
			p.Y = p.screenHeight - int(my) - int(mh)
			if p.Y < 0 {
				p.Y = 0
			}
		}
		if p.OnMouseUp != nil {
			p.OnMouseUp(MouseEvent{})
		}
	} else {
		// 右键单击（非拖拽）：显示右键菜单
		p.showContextMenu()
		if p.OnRightClick != nil {
			p.OnRightClick(MouseEvent{})
		}
	}
}

//export petMenuSelect
func petMenuSelect(handle C.intptr_t, menuID C.int) {
	p := petLookup(int(handle))
	if p == nil {
		return
	}
	p.handleMenuCommand(int(menuID))
}

// 整数 min/max 辅助函数（与 Windows 版本一致）
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
