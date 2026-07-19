//go:build darwin

package platform

type darwinPlatform struct {
	Width        int
	Height       int
	X            int
	Y            int
	OnClick      MouseEventHandler
	OnRightClick MouseEventHandler
	OnMouseDown  MouseEventHandler
	OnMouseUp    MouseEventHandler
	OnMouseMove  MouseEventHandler
}

func (p *darwinPlatform) Init() error {
	return nil
}

func (p *darwinPlatform) CreateWindow(title string, x, y, width, height int) error {
	p.Width = width
	p.Height = height
	p.X = x
	p.Y = y
	return nil
}

func (p *darwinPlatform) Show() {
}

func (p *darwinPlatform) Run() {
	for {}
}

func (p *darwinPlatform) Close() {
}

func (p *darwinPlatform) Navigate(url string) {
}

func (p *darwinPlatform) SetSize(width, height int) {
	p.Width = width
	p.Height = height
}

func (p *darwinPlatform) SetPosition(x, y int) {
	p.X = x
	p.Y = y
}

func (p *darwinPlatform) GetSize() (int, int) {
	return p.Width, p.Height
}

func (p *darwinPlatform) GetPosition() (int, int) {
	return p.X, p.Y
}

func (p *darwinPlatform) ShowNotification(message string) {
}

func (p *darwinPlatform) SetOnClick(handler MouseEventHandler) {
	p.OnClick = handler
}

func (p *darwinPlatform) SetOnRightClick(handler MouseEventHandler) {
	p.OnRightClick = handler
}

func (p *darwinPlatform) SetOnMouseDown(handler MouseEventHandler) {
	p.OnMouseDown = handler
}

func (p *darwinPlatform) SetOnMouseUp(handler MouseEventHandler) {
	p.OnMouseUp = handler
}

func (p *darwinPlatform) SetOnMouseMove(handler MouseEventHandler) {
	p.OnMouseMove = handler
}

func (p *darwinPlatform) Eval(code string) {
}

func (p *darwinPlatform) UpdatePet(action string, frame int) {
}

func (p *darwinPlatform) ShowBubble(text string) {
}

func (p *darwinPlatform) HideBubble() {
}

func (p *darwinPlatform) SetDragging(dragging bool) {
}
