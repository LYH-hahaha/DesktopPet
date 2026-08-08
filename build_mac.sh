#!/bin/bash
# 构建 macOS .app 包（点击即可运行）
# 用法: ./build_mac.sh
set -e

APP_NAME="DesktopPet"
BUNDLE_ID="com.deskpet.desktoppet"
APP_DIR="build/${APP_NAME}.app"

echo "==> 编译二进制"
mkdir -p "${APP_DIR}/Contents/MacOS"
mkdir -p "${APP_DIR}/Contents/Resources"
go build -o "${APP_DIR}/Contents/MacOS/${APP_NAME}" .

echo "==> 写入 Info.plist"
cat > "${APP_DIR}/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleName</key>
    <string>${APP_NAME}</string>
    <key>CFBundleDisplayName</key>
    <string>桌面小宠物</string>
    <key>CFBundleIdentifier</key>
    <string>${BUNDLE_ID}</string>
    <key>CFBundleVersion</key>
    <string>1.0</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleExecutable</key>
    <string>${APP_NAME}</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.13</string>
    <key>LSUIElement</key>
    <true/>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
PLIST

echo "==> 复制资源"
rm -rf "${APP_DIR}/Contents/Resources/assets"
cp -R assets "${APP_DIR}/Contents/Resources/assets"

echo "==> 生成 PkgInfo"
echo -n "APPL????" > "${APP_DIR}/Contents/PkgInfo"

echo "==> Ad-hoc 签名"
codesign -s - --force --deep "${APP_DIR}" 2>/dev/null || echo "  (签名跳过：未配置签名身份)"

echo "==> 打包 zip"
ZIP_FILE="build/${APP_NAME}.zip"
rm -f "${ZIP_FILE}"
cd build && zip -r "${APP_NAME}.zip" "${APP_NAME}.app" -x "*.DS_Store" >/dev/null && cd ..
ZIP_PATH="$(pwd)/${ZIP_FILE}"

echo ""
echo "==> 完成"
echo "应用包: $(pwd)/${APP_DIR}"
echo "压缩包: ${ZIP_PATH}"
echo "双击运行: open ${APP_DIR}"
echo ""
echo "注意：分发给其他 Mac 时，由于未做开发者签名/公证，接收方首次打开可能被 Gatekeeper 拦截。"
echo "接收方可：右键 -> 打开；或在终端执行 xattr -dr com.apple.quarantine ${APP_DIR}"
