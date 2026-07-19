package render

import (
	"image"
	"image/draw"
	"os"

	"github.com/lxn/win"
	"golang.org/x/image/bmp"
	"golang.org/x/image/png"
)

type Renderer struct {
	hdc win.HDC
	width, height int
}

func NewRenderer(hdc win.HDC, width, height int) *Renderer {
	return &Renderer{
		hdc:    hdc,
		width:  width,
		height: height,
	}
}

func (r *Renderer) DrawImage(img image.Image, x, y int) {
	bmpData := image.NewRGBA(img.Bounds())
	draw.Draw(bmpData, bmpData.Bounds(), img, img.Bounds().Min, draw.Src)

	hBitmap := win.CreateCompatibleBitmap(r.hdc, int32(bmpData.Bounds().Dx()), int32(bmpData.Bounds().Dy()))
	if hBitmap == 0 {
		return
	}
	defer win.DeleteObject(win.HGDIOBJ(hBitmap))

	memoryDC := win.CreateCompatibleDC(r.hdc)
	if memoryDC == 0 {
		return
	}
	defer win.DeleteDC(memoryDC)

	oldBitmap := win.SelectObject(memoryDC, win.HGDIOBJ(hBitmap))
	defer win.SelectObject(memoryDC, oldBitmap)

	bmpDataPtr := unsafe.Pointer(&bmpData.Pix[0])
	win.SetDIBitsToDevice(
		memoryDC,
		0, 0,
		int32(bmpData.Bounds().Dx()),
		int32(bmpData.Bounds().Dy()),
		0, 0,
		0,
		uint32(bmpData.Bounds().Dy()),
		bmpDataPtr,
		&win.BITMAPINFO{
			BmiHeader: win.BITMAPINFOHEADER{
				BiSize:        uint32(unsafe.Sizeof(win.BITMAPINFOHEADER{})),
				BiWidth:       int32(bmpData.Bounds().Dx()),
				BiHeight:      -int32(bmpData.Bounds().Dy()),
				BiPlanes:      1,
				BiBitCount:    32,
				BiCompression: win.BI_RGB,
			},
		},
		win.DIB_RGB_COLORS,
	)

	win.BitBlt(
		r.hdc,
		int32(x), int32(y),
		int32(bmpData.Bounds().Dx()),
		int32(bmpData.Bounds().Dy()),
		memoryDC,
		0, 0,
		win.SRCCOPY,
	)
}

func (r *Renderer) DrawText(text string, x, y int, color win.COLORREF, font *win.LOGFONT) {
	if font != nil {
		hFont := win.CreateFontIndirect(font)
		if hFont != 0 {
			oldFont := win.SelectObject(r.hdc, win.HGDIOBJ(hFont))
			defer win.SelectObject(r.hdc, oldFont)
			defer win.DeleteObject(win.HGDIOBJ(hFont))
		}
	}

	win.SetTextColor(r.hdc, color)
	win.SetBkMode(r.hdc, win.TRANSPARENT)

	textPtr, _ := syscall.UTF16PtrFromString(text)
	win.TextOut(r.hdc, int32(x), int32(y), textPtr, int32(len(text)))
}

func (r *Renderer) Clear() {
	win.Rectangle(r.hdc, 0, 0, int32(r.width), int32(r.height))
}

func LoadPNG(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return png.Decode(file)
}

func LoadBMP(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return bmp.Decode(file)
}

func CreateCompatibleDC(hdc win.HDC) win.HDC {
	return win.CreateCompatibleDC(hdc)
}

func DeleteDC(hdc win.HDC) {
	win.DeleteDC(hdc)
}
