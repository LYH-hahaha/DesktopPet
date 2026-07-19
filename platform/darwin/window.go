//go:build darwin

package darwin

import (
	"desktopPet/platform"
)

type Window struct {
	Width        int
	Height       int
	X            int
	Y            int
	OnClick      platform.MouseEventHandler
	OnRightClick platform.MouseEventHandler
	OnMouseDown  platform.MouseEventHandler
	OnMouseUp    platform.MouseEventHandler
	OnMouseMove  platform.MouseEventHandler
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
	return nil
}

func (w *Window) Show() {
}

func (w *Window) Run() {
	for {}
}

func (w *Window) Close() {
}

func (w *Window) Navigate(url string) {
}

func (w *Window) SetSize(width, height int) {
	w.Width = width
	w.Height = height
}

func (w *Window) SetPosition(x, y int) {
	w.X = x
	w.Y = y
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
}

func (w *Window) ShowBubble(text string) {
}

func (w *Window) HideBubble() {
}

func (w *Window) SetDragging(dragging bool) {
}
