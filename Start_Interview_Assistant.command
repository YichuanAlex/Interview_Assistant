#!/bin/zsh
# 启动 Interview_Assistant（生产包）
cd /Users/didi/Downloads/Interview_Assistant

# 激活 conda interview 环境（转录依赖所在环境）
CONDA_BASE="/opt/homebrew/Caskroom/miniconda/base"
if [ -f "$CONDA_BASE/etc/profile.d/conda.sh" ]; then
    source "$CONDA_BASE/etc/profile.d/conda.sh"
    conda activate interview
else
    echo "警告：未找到 conda，转录功能可能无法使用"
fi

APP="/Users/didi/Downloads/Interview_Assistant/build/bin/Interview_Assistant.app"
if [ ! -d "$APP" ]; then
    echo "未找到构建产物，请先执行 wails build"
    exit 1
fi

# 清除 macOS 隔离属性，避免每次手动授权
xattr -cr "$APP"

# 直接运行二进制，以便在 Terminal 中看到后端日志
exec "$APP/Contents/MacOS/Interview_Assistant"
