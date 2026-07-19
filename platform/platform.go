package platform

type MouseEvent struct {
	X, Y int
}

type MouseEventHandler func(event MouseEvent)
type ExpressionChangeHandler func(imageName string)

type Platform interface {
	CreateWindow(title string, x, y, width, height int) error
	Show()
	Run()
	Close()
	Navigate(url string)
	SetSize(width, height int)
	SetPosition(x, y int)
	GetSize() (int, int)
	GetPosition() (int, int)
	ShowNotification(message string)
	SetOnClick(handler MouseEventHandler)
	SetOnRightClick(handler MouseEventHandler)
	SetOnMouseDown(handler MouseEventHandler)
	SetOnMouseUp(handler MouseEventHandler)
	SetOnMouseMove(handler MouseEventHandler)
	SetOnExpressionChange(handler ExpressionChangeHandler)
	Eval(code string)
	Init() error
	UpdatePet(action string, frame int)
	ShowBubble(text string)
	HideBubble()
	SetDragging(dragging bool)
	SwitchToNextImage() string
	GetCurrentImageName() string
}
