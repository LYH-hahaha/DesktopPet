// go:build darwin

// macOS Cocoa 桥接实现：窗口、视图绘制、鼠标事件、菜单。
// 放在独立的 .m 文件中（单一翻译单元），避免 cgo 在多个目标文件中
// 重复生成 ObjC 类符号（@implementation 只能出现一次）。
//
// 注意：C 侧不持有任何 Go 指针（cgo 指针规则禁止把含 Go 指针的内存传给 C）。
// 这里只持有一个整数句柄（intptr_t），由 Go 侧的注册表映射回 *darwinPlatform。

#import <Cocoa/Cocoa.h>
#import <CoreGraphics/CoreGraphics.h>
#import <stdint.h>

// 由 Go 通过 //export 导出的回调（链接时解析），第一个参数为整数句柄
extern void petLeftClick(intptr_t handle, int x, int y);
extern void petRightMouseDown(intptr_t handle, int x, int y, int absX, int absY);
extern void petRightMouseDragged(intptr_t handle, int dx, int dy);
extern void petRightMouseUp(intptr_t handle, int wasDrag);
extern void petMenuSelect(intptr_t handle, int menuID);

// ===== 由 RGBA 像素数据创建 CGImage（预乘 alpha）=====
static CGImageRef makeCGImage(const void *rgba, int w, int h) {
  if (!rgba || w <= 0 || h <= 0) return NULL;
  CGColorSpaceRef cs = CGColorSpaceCreateDeviceRGB();
  CGContextRef ctx = CGBitmapContextCreate((void *)rgba, (size_t)w, (size_t)h, 8, (size_t)w * 4, cs,
                                           kCGImageAlphaPremultipliedLast);
  CGImageRef img = ctx ? CGBitmapContextCreateImage(ctx) : NULL;
  if (ctx) CGContextRelease(ctx);
  CGColorSpaceRelease(cs);
  return img;  // 调用方负责 release
}

// ===== 自定义视图：绘制宠物图片 + 气泡，并处理鼠标事件 =====
@interface PetView : NSView {
 @public
  NSImage *petImage;
  NSRect imageRect;
  NSString *bubbleText;
  int bubbleX, bubbleY, bubbleW, bubbleH, tailH, radius, padding, lineHeight, fontSize;
  intptr_t goHandle;
  BOOL dragging;
  BOOL isDrag;
  NSPoint dragStartScreen;
  NSPoint winStartOrigin;
}
- (instancetype)initWithFrame:(NSRect)frame handle:(intptr_t)h;
- (void)setPetImage:(CGImageRef)img rect:(NSRect)r;
- (void)setBubble:(NSString *)text
               bx:(int)bx
               by:(int)by
               bw:(int)bw
               bh:(int)bh
               th:(int)th
              rad:(int)rad
              pad:(int)pad
               lh:(int)lh
               fs:(int)fs;
- (void)clearBubble;
- (void)redraw;
- (void)drawBubbleInRect;
@end

@implementation PetView

- (instancetype)initWithFrame:(NSRect)frame handle:(intptr_t)h {
  self = [super initWithFrame:frame];
  if (self) {
    petImage = nil;
    bubbleText = nil;
    goHandle = h;
    dragging = NO;
    isDrag = NO;
  }
  return self;
}

// 使用左上角原点坐标系，与 Windows 版本一致
- (BOOL)isFlipped {
  return YES;
}
- (BOOL)acceptsFirstResponder {
  return YES;
}

- (void)drawRect:(NSRect)dirtyRect {
  (void)dirtyRect;
  // 清空为完全透明
  [[NSColor clearColor] set];
  NSRectFill(self.bounds);

  // 绘制宠物图片（NSImage 绘制会自动适配 isFlipped）
  if (petImage) {
    [petImage drawInRect:imageRect
                fromRect:NSZeroRect
               operation:NSCompositingOperationSourceOver
                fraction:1.0];
  }

  // 绘制气泡
  if (bubbleText && [bubbleText length] > 0) {
    [self drawBubbleInRect];
  }
}

- (void)drawBubbleInRect {
  CGFloat x = bubbleX, y = bubbleY, w = bubbleW, hh = bubbleH, r = radius;
  if (r < 1) r = 1;
  CGFloat tailW = 20.0;
  CGFloat cx = x + w / 2.0;

  // 气泡主路径（圆角矩形 + 底部尾巴）
  NSBezierPath * (^makePath)(CGFloat, CGFloat) = ^(CGFloat ox, CGFloat oy) {
    NSBezierPath *path = [NSBezierPath bezierPath];
    [path appendBezierPathWithRoundedRect:NSMakeRect(x + ox, y + oy, w, hh) xRadius:r yRadius:r];
    [path moveToPoint:NSMakePoint(cx - tailW / 2.0 + ox, y + hh + oy)];
    [path lineToPoint:NSMakePoint(cx + ox, y + hh + tailH + oy)];
    [path lineToPoint:NSMakePoint(cx + tailW / 2.0 + ox, y + hh + oy)];
    [path closePath];
    return path;
  };

  // 柔和投影（粉色调，偏移 2,3）
  NSBezierPath *shadowPath = makePath(2, 3);
  [[NSColor colorWithDeviceRed:130.0 / 255.0 green:90.0 / 255.0 blue:110.0 / 255.0
                         alpha:0.25] setFill];
  [shadowPath fill];

  // 粉色渐变背景（浅粉 -> 略深粉，营造柔和少女感）
  NSBezierPath *bgPath = makePath(0, 0);
  NSGradient *grad =
      [[NSGradient alloc] initWithStartingColor:[NSColor colorWithDeviceRed:1.0
                                                                      green:246.0 / 255.0
                                                                       blue:250.0 / 255.0
                                                                      alpha:1.0]
                                    endingColor:[NSColor colorWithDeviceRed:1.0
                                                                      green:214.0 / 255.0
                                                                       blue:228.0 / 255.0
                                                                      alpha:1.0]];
  [grad drawInBezierPath:bgPath angle:270.0];

  // 外边框（玫瑰粉实线）
  [[NSColor colorWithDeviceRed:1.0 green:143.0 / 255.0 blue:177.0 / 255.0 alpha:1.0] setStroke];
  [bgPath setLineWidth:2.0];
  [bgPath stroke];

  // 内侧花边（虚线，营造蕾丝边效果）
  CGFloat inset = 5.0;
  CGFloat lr = r > inset ? r - inset : 1.0;
  NSBezierPath *lace = [NSBezierPath
      bezierPathWithRoundedRect:NSMakeRect(x + inset, y + inset, w - inset * 2, hh - inset * 2)
                        xRadius:lr
                        yRadius:lr];
  [[NSColor colorWithDeviceRed:1.0 green:170.0 / 255.0 blue:200.0 / 255.0 alpha:1.0] setStroke];
  [lace setLineWidth:1.0];
  CGFloat dash[2] = {3.0, 3.0};
  [lace setLineDash:dash count:2 phase:0];
  [lace stroke];

  // 顶部两侧的小装饰圆点（蕾丝珠饰）
  [[NSColor colorWithDeviceRed:1.0 green:143.0 / 255.0 blue:177.0 / 255.0 alpha:1.0] setFill];
  CGFloat dotR = 2.2;
  [[NSBezierPath bezierPathWithOvalInRect:NSMakeRect(x + 9, y + 9, dotR * 2, dotR * 2)] fill];
  [[NSBezierPath
      bezierPathWithOvalInRect:NSMakeRect(x + w - 9 - dotR * 2, y + 9, dotR * 2, dotR * 2)] fill];

  // 绘制文字（深玫瑰色）
  NSArray *lines = [bubbleText componentsSeparatedByString:@"\n"];
  NSFont *font = [NSFont systemFontOfSize:fontSize > 0 ? fontSize : 24];
  NSDictionary *attrs = @{
    NSFontAttributeName : font,
    NSForegroundColorAttributeName : [NSColor colorWithDeviceRed:150.0 / 255.0
                                                           green:45.0 / 255.0
                                                            blue:90.0 / 255.0
                                                           alpha:1.0]
  };
  int i = 0;
  for (NSString *ln in lines) {
    NSPoint tp = NSMakePoint(bubbleX + padding, bubbleY + padding + i * lineHeight);
    [ln drawAtPoint:tp withAttributes:attrs];
    i++;
  }
}

- (void)setPetImage:(CGImageRef)img rect:(NSRect)r {
  // ARC 下直接重新赋值即可自动释放旧对象
  if (img) {
    petImage = [[NSImage alloc] initWithCGImage:img size:r.size];
  } else {
    petImage = nil;
  }
  imageRect = r;
}

- (void)setBubble:(NSString *)text
               bx:(int)bx
               by:(int)by
               bw:(int)bw
               bh:(int)bh
               th:(int)th
              rad:(int)rad
              pad:(int)pad
               lh:(int)lh
               fs:(int)fs {
  bubbleText = [text copy];  // ARC 自动释放旧对象
  bubbleX = bx;
  bubbleY = by;
  bubbleW = bw;
  bubbleH = bh;
  tailH = th;
  radius = rad;
  padding = pad;
  lineHeight = lh;
  fontSize = fs;
}

- (void)clearBubble {
  bubbleText = nil;  // ARC 自动释放
}

- (void)redraw {
  [self setNeedsDisplay:YES];
}

// 左键单击：切换到下一张图片
- (void)mouseDown:(NSEvent *)event {
  NSPoint p = [self convertPoint:[event locationInWindow] fromView:nil];
  petLeftClick(goHandle, (int)p.x, (int)p.y);
}

// 右键按下：开始拖拽检测
- (void)rightMouseDown:(NSEvent *)event {
  NSPoint p = [self convertPoint:[event locationInWindow] fromView:nil];
  NSPoint sp = [NSEvent mouseLocation];
  dragging = YES;
  isDrag = NO;
  dragStartScreen = sp;
  winStartOrigin = [[self window] frame].origin;
  petRightMouseDown(goHandle, (int)p.x, (int)p.y, (int)sp.x, (int)sp.y);
}

- (void)rightMouseDragged:(NSEvent *)event {
  if (!dragging) return;
  NSPoint sp = [NSEvent mouseLocation];
  CGFloat dx = sp.x - dragStartScreen.x;
  CGFloat dy = sp.y - dragStartScreen.y;
  if (!isDrag && dx * dx + dy * dy > 25) isDrag = YES;
  if (isDrag) {
    NSPoint newOrigin = NSMakePoint(winStartOrigin.x + dx, winStartOrigin.y + dy);
    [[self window] setFrameOrigin:newOrigin];
    petRightMouseDragged(goHandle, (int)dx, (int)dy);
  }
}

- (void)rightMouseUp:(NSEvent *)event {
  BOOL wasDrag = isDrag;
  dragging = NO;
  isDrag = NO;
  petRightMouseUp(goHandle, wasDrag ? 1 : 0);
}

@end

// ===== 菜单 target：菜单项点击时回调 Go =====
@interface MenuTarget : NSObject
@property(nonatomic, assign) intptr_t goHandle;
- (void)menuAction:(NSMenuItem *)sender;
@end

@implementation MenuTarget
- (void)menuAction:(NSMenuItem *)sender {
  petMenuSelect(self.goHandle, (int)[sender tag]);
}
// 强制启用所有以此对象为 target 的菜单项，避免 AppKit 默认校验把项置灰
- (BOOL)validateMenuItem:(NSMenuItem *)menuItem {
  return YES;
}
@end

// ===== 应用与窗口管理 =====
void setupApp() {
  @autoreleasepool {
    NSApplication *app = [NSApplication sharedApplication];
    [app setActivationPolicy:NSApplicationActivationPolicyAccessory];
  }
}

int getScreenHeight() {
  @autoreleasepool {
    return (int)[[NSScreen mainScreen] frame].size.height;
  }
}

// 创建无边框透明置顶窗口。handle 为 Go 侧注册的整数句柄，view 存入 *outView
void *createPetWindow(intptr_t handle, int x, int y, int w, int h, void **outView) {
  @autoreleasepool {
    NSRect frame = NSMakeRect(x, y, w, h);
    PetView *view = [[PetView alloc] initWithFrame:frame handle:handle];
    NSWindow *win = [[NSWindow alloc] initWithContentRect:frame
                                                styleMask:NSWindowStyleMaskBorderless
                                                  backing:NSBackingStoreBuffered
                                                    defer:NO];
    [win setContentView:view];
    [win setOpaque:NO];
    [win setBackgroundColor:[NSColor clearColor]];
    [win setHasShadow:NO];
    [win setLevel:NSStatusWindowLevel];  // 置顶，高于普通窗口
    [win setIgnoresMouseEvents:NO];
    [win setAcceptsMouseMovedEvents:YES];
    [win setMovableByWindowBackground:NO];
    [win setReleasedWhenClosed:NO];
    if (outView) *outView = (__bridge void *)view;
    return (__bridge_retained void *)win;
  }
}

void *createMenuTarget(intptr_t handle) {
  @autoreleasepool {
    MenuTarget *t = [[MenuTarget alloc] init];
    t.goHandle = handle;
    return (__bridge_retained void *)t;
  }
}

void winOrderFront(void *win) {
  @autoreleasepool {
    NSWindow *w = (__bridge NSWindow *)win;
    [w orderFrontRegardless];
    [[NSApplication sharedApplication] activateIgnoringOtherApps:YES];
  }
}

void winSetFrame(void *win, int x, int y, int w, int h) {
  @autoreleasepool {
    NSWindow *wn = (__bridge NSWindow *)win;
    [wn setFrame:NSMakeRect(x, y, w, h) display:YES];
  }
}

void winGetFrame(void *win, int *x, int *y, int *w, int *h) {
  @autoreleasepool {
    NSWindow *wn = (__bridge NSWindow *)win;
    NSRect f = [wn frame];
    if (x) *x = (int)f.origin.x;
    if (y) *y = (int)f.origin.y;
    if (w) *w = (int)f.size.width;
    if (h) *h = (int)f.size.height;
  }
}

void setViewImage(void *viewPtr, const void *rgba, int w, int h, int dx, int dy, int dw, int dh) {
  @autoreleasepool {
    PetView *v = (__bridge PetView *)viewPtr;
    CGImageRef img = makeCGImage(rgba, w, h);
    if (img) {
      [v setPetImage:img rect:NSMakeRect(dx, dy, dw, dh)];
      CGImageRelease(img);  // view 内部已拷贝为 NSImage
    }
    [v redraw];
  }
}

void setViewBubble(void *viewPtr, const char *text, int bx, int by, int bw, int bh, int th, int rad,
                   int pad, int lh, int fs) {
  @autoreleasepool {
    PetView *v = (__bridge PetView *)viewPtr;
    NSString *s = text ? [NSString stringWithUTF8String:text] : @"";
    [v setBubble:s bx:bx by:by bw:bw bh:bh th:th rad:rad pad:pad lh:lh fs:fs];
    [v redraw];
  }
}

void clearViewBubble(void *viewPtr) {
  @autoreleasepool {
    PetView *v = (__bridge PetView *)viewPtr;
    [v clearBubble];
    [v redraw];
  }
}

void viewRedraw(void *viewPtr) {
  @autoreleasepool {
    PetView *v = (__bridge PetView *)viewPtr;
    [v redraw];
  }
}

// 设置整个视图的透明度（人物 + 文字一起），1.0=不透明，0.0=完全透明
void setViewAlpha(void *viewPtr, double alpha) {
  @autoreleasepool {
    NSView *v = (__bridge NSView *)viewPtr;
    [v setAlphaValue:alpha];
  }
}

// ===== 菜单 =====
void *menuCreate() {
  @autoreleasepool {
    NSMenu *m = [[NSMenu alloc] init];
    return (__bridge_retained void *)m;
  }
}

void menuAddItem(void *menu, const char *title, int tag, int isSeparator, int isSubmenu,
                 void *submenu, int checked, void *target) {
  @autoreleasepool {
    NSMenu *m = (__bridge NSMenu *)menu;
    if (isSeparator) {
      [m addItem:[NSMenuItem separatorItem]];
      return;
    }
    NSString *t = [NSString stringWithUTF8String:title];
    NSMenuItem *item = [[NSMenuItem alloc] initWithTitle:t
                                                  action:@selector(menuAction:)
                                           keyEquivalent:@""];
    [item setTag:tag];
    if (checked) [item setState:NSControlStateValueOn];
    if (isSubmenu && submenu) {
      NSMenu *sub = (__bridge NSMenu *)submenu;
      [item setSubmenu:sub];
    }
    if (target) {
      // 子菜单父项也设置 target，使 validateMenuItem: 能启用它（避免 AppKit
      // 因 target=nil 走响应链校验失败而把父项置灰，连带子菜单无法展开）
      [item setTarget:(__bridge id)target];
    }
    [m addItem:item];
  }
}

void menuPopup(void *menu) {
  @autoreleasepool {
    NSMenu *m = (__bridge NSMenu *)menu;
    NSPoint p = [NSEvent mouseLocation];
    [m popUpMenuPositioningItem:nil atLocation:p inView:nil];
  }
}

// ===== 运行循环 =====
void runApp() {
  @autoreleasepool {
    NSApplication *app = [NSApplication sharedApplication];
    [app run];
  }
}

void stopApp() {
  @autoreleasepool {
    NSApplication *app = [NSApplication sharedApplication];
    [app stop:nil];
    // 投递一个空事件以解除 [app run] 的阻塞
    NSEvent *e = [NSEvent otherEventWithType:NSEventTypeApplicationDefined
                                    location:NSMakePoint(0, 0)
                               modifierFlags:0
                                   timestamp:0
                                windowNumber:0
                                     context:nil
                                     subtype:0
                                       data1:0
                                       data2:0];
    [app postEvent:e atStart:YES];
  }
}
