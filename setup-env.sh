#!/bin/bash
# Go版本管理器环境变量设置脚本

# 获取用户主目录
HOME_DIR="$HOME"

# 设置环境变量
GOROOT="$HOME_DIR/.go-version/current"
GOPATH="$HOME_DIR/go"
GOBIN="$GOROOT/bin"

# 检查shell类型
SHELL_TYPE=$(basename "$SHELL")

# 根据shell类型选择配置文件
case $SHELL_TYPE in
    bash)
        CONFIG_FILE="$HOME/.bashrc"
        ;;
    zsh)
        CONFIG_FILE="$HOME/.zshrc"
        ;;
    fish)
        CONFIG_FILE="$HOME/.config/fish/config.fish"
        ;;
    *)
        echo "不支持的shell类型: $SHELL_TYPE"
        echo "请手动将以下环境变量添加到您的shell配置文件中："
        echo ""
        echo "export GOROOT=$GOROOT"
        echo "export GOPATH=$GOPATH"
        echo "export PATH=\$PATH:\$GOROOT/bin"
        exit 1
        ;;
esac

# 检查是否已经存在这些环境变量
if grep -q "GOROOT=" "$CONFIG_FILE"; then
    echo "环境变量已存在于 $CONFIG_FILE 中"
else
    echo "" >> "$CONFIG_FILE"
    echo "# Go版本管理器环境变量" >> "$CONFIG_FILE"
    echo "export GOROOT=$GOROOT" >> "$CONFIG_FILE"
    echo "export GOPATH=$GOPATH" >> "$CONFIG_FILE"
    echo "export PATH=\$PATH:\$GOROOT/bin" >> "$CONFIG_FILE"
    echo "环境变量已添加到 $CONFIG_FILE"
fi

echo ""
echo "请运行以下命令使环境变量生效："
echo "source $CONFIG_FILE"
echo ""
echo "或者重新打开终端"
