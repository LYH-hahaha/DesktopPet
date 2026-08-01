package main

import (
	"path/filepath"
	"runtime"

	"desktopPet/platform"
)

// main 是程序入口。
// 关键：必须在最开始 runtime.LockOSThread()，将 main goroutine 绑定到一个 OS 线程。
// Windows 的消息队列和窗口都绑定到创建它们的线程：CreateWindow、PeekMessage、DispatchMessage、
// UpdateLayeredWindow、wndProc 回调必须在同一个 OS 线程上执行。
// 若 Go runtime 在某个 syscall 边界把 main goroutine 调度到另一个 OS 线程，
// PeekMessage 就会读不到原线程消息队列里的消息（表现为：循环在跑但 processed 0 msgs），
// 系统随后判定窗口 hung 并创建 Ghost 窗口接管输入，导致"卡死"。
func main() {
	runtime.LockOSThread()

	exePath, _ := filepath.Abs(".")
	_ = exePath

	p, err := platform.NewPlatform()
	if err != nil {
		panic(err)
	}

	err = p.Init()
	if err != nil {
		panic(err)
	}

	petWidth := 200
	petHeight := 300

	x, y := 100, 100

	err = p.CreateWindow("桌面宠物", x, y, petWidth, petHeight)
	if err != nil {
		panic(err)
	}

	p.Show()

	dp := NewDesktopPet(p)

	dp.Start()
}

// 图片对应的气泡文字
var imageTexts = map[string]string{
	"主体":  "你好呀~我是你的桌面小宠物!",
	"挥手":  "再见啦~下次见!",
	"奔跑":  "跑跑跑~好开心!",
	"跳跃":  "蹦蹦跳跳~",
	"打哈欠": "困了...想睡觉...",
	"害羞":  "害羞~不要看我!",
	"生气":  "哼!不理你了!",
	"委屈":  "呜呜...好委屈...",
	"震惊":  "什么?!不会吧!",
	"喝饮料": "咕噜咕噜~好喝!",
	"睡觉":  "Zzz...好困...",
}

type DesktopPet struct {
	platform platform.Platform
	running  bool
}

func NewDesktopPet(p platform.Platform) *DesktopPet {
	return &DesktopPet{
		platform: p,
		running:  true,
	}
}

func (dp *DesktopPet) HandleClick(event platform.MouseEvent) {
	// 左键单击：切换到下一张图片
	name := dp.platform.SwitchToNextImage()
	if text, ok := imageTexts[name]; ok {
		dp.platform.ShowBubble(text)
	} else {
		dp.platform.ShowBubble(name)
	}
}

func (dp *DesktopPet) HandleExpressionChange(imageName string) {
	// 表情切换：更新气泡文字
	if text, ok := imageTexts[imageName]; ok {
		dp.platform.ShowBubble(text)
	} else {
		dp.platform.ShowBubble(imageName)
	}
}

func (dp *DesktopPet) Start() {
	dp.platform.SetOnClick(dp.HandleClick)
	dp.platform.SetOnExpressionChange(dp.HandleExpressionChange)

	// 显示初始气泡
	dp.platform.ShowBubble("你好呀~我是你的桌面小宠物!")

	dp.platform.Run()
}

func (dp *DesktopPet) Stop() {
	dp.running = false
	dp.platform.Close()
}
