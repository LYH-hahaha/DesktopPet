package main

import (
	"path/filepath"

	"desktopPet/platform"
)

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

func main() {
	// 设置 DLL 目录（确保从正确的路径加载）
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
