#!/bin/bash
# Mirror命令完整演示脚本
# 此脚本演示go-version mirror命令的所有功能

echo "======================================"
echo "    Go版本管理器 Mirror命令演示"
echo "======================================"
echo

# 检查工具是否可用
if ! command -v go-version &> /dev/null; then
    echo "❌ go-version 工具未找到，请先安装工具"
    exit 1
fi

echo "✅ go-version 工具已安装"
echo

# 1. 基本镜像源管理
echo "1️⃣  基本镜像源管理"
echo "-----------------------------------"

echo "🔍 查看所有可用镜像源："
go-version mirror list
echo

echo "📋 查看镜像源详细信息："
go-version mirror list --details
echo

# 2. 镜像源测试
echo "2️⃣  镜像源测试功能"
echo "-----------------------------------"

echo "⚡ 测试所有镜像源速度："
echo "（这将需要一些时间，请耐心等待...）"
go-version mirror test
echo

echo "🎯 测试指定镜像源（goproxy-cn）："
go-version mirror test --name goproxy-cn
echo

echo "🔍 验证官方镜像源："
go-version mirror validate --name official
echo

# 3. 自动选择最快镜像
echo "3️⃣  自动选择最快镜像"
echo "-----------------------------------"

echo "🚀 自动选择最快的镜像源："
go-version mirror fastest
echo

echo "📊 显示详细测试过程并选择最快镜像："
go-version mirror fastest --details
echo

# 4. 自定义镜像源管理（演示，不实际执行）
echo "4️⃣  自定义镜像源管理（演示）"
echo "-----------------------------------"

echo "➕ 添加自定义镜像源的命令格式："
echo "go-version mirror add \\"
echo "  --name mycompany \\"
echo "  --url \"https://mirrors.mycompany.com/golang/\" \\"
echo "  --description \"公司内部镜像\" \\"
echo "  --region \"内网\" \\"
echo "  --priority 1"
echo

echo "🗑️  移除自定义镜像源的命令格式："
echo "go-version mirror remove --name mycompany"
echo

# 5. 在安装中使用镜像源（演示）
echo "5️⃣  在安装中使用镜像源（演示）"
echo "-----------------------------------"

echo "📦 使用指定镜像源安装Go版本："
echo "go-version install 1.21.0 --mirror goproxy-cn"
echo

echo "🔄 自动选择最快镜像源安装："
echo "go-version install 1.21.0 --auto-mirror"
echo

echo "⚙️  组合使用其他选项："
echo "go-version install 1.21.0 --mirror aliyun --force --timeout 600"
echo

# 6. 高级用法
echo "6️⃣  高级用法提示"
echo "-----------------------------------"

echo "💡 实用技巧："
echo "• 定期运行 'go-version mirror test' 检查镜像源状态"
echo "• 使用 'go-version mirror fastest' 找到当前最快的镜像源"
echo "• 在网络环境变化时重新测试镜像源速度"
echo "• 为不同的项目或环境添加专用的自定义镜像源"
echo

# 7. 镜像源信息总结
echo "7️⃣  内置镜像源信息"
echo "-----------------------------------"

cat << 'EOF'
镜像源     | 描述              | 地区   | 适用场景
---------|-------------------|--------|------------------
official | Go官方下载源        | 全球   | 海外用户，完整功能
goproxy-cn| 七牛云Go代理镜像   | 中国   | 国内用户，速度快
aliyun   | 阿里云镜像源        | 中国   | 阿里云用户
tencent  | 腾讯云镜像源        | 中国   | 腾讯云用户
huawei   | 华为云镜像源        | 中国   | 华为云用户
EOF

echo
echo "📚 更多帮助信息："
echo "• go-version mirror --help       # 查看mirror命令帮助"
echo "• go-version mirror list --help  # 查看list子命令帮助"
echo "• go-version mirror test --help  # 查看test子命令帮助"
echo "• go-version mirror add --help   # 查看add子命令帮助"

echo
echo "🎉 Mirror命令演示完成！"
echo "====================================="