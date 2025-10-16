#!/bin/bash

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 配置
APP_NAME="emby-bot"
INSTALL_DIR="/opt/emby-telegram"
SERVICE_NAME="emby-telegram"
SERVICE_USER="emby-bot"

# 打印信息
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否为 root 用户
check_root() {
    if [ "$EUID" -ne 0 ]; then
        print_error "请使用 root 权限运行此脚本"
        exit 1
    fi
}

# 检查系统
check_system() {
    if [ ! -f /etc/os-release ]; then
        print_error "无法检测操作系统"
        exit 1
    fi

    source /etc/os-release
    print_info "检测到系统: $NAME $VERSION"
}

# 检查 systemd
check_systemd() {
    if ! command -v systemctl &> /dev/null; then
        print_error "系统不支持 systemd"
        exit 1
    fi
    print_info "systemd 检查通过"
}

# 创建系统用户
create_user() {
    if id "$SERVICE_USER" &>/dev/null; then
        print_info "用户 $SERVICE_USER 已存在"
    else
        print_info "创建系统用户: $SERVICE_USER"
        useradd -r -s /bin/false -d $INSTALL_DIR $SERVICE_USER
    fi
}

# 创建目录结构
create_directories() {
    print_info "创建目录结构..."
    mkdir -p $INSTALL_DIR/{bin,data,logs}
    mkdir -p /etc/$SERVICE_NAME
}

# 安装二进制文件
install_binary() {
    if [ ! -f "bin/${APP_NAME}-linux-amd64" ]; then
        print_error "找不到编译后的二进制文件: bin/${APP_NAME}-linux-amd64"
        print_info "请先运行: make build-linux"
        exit 1
    fi

    print_info "安装二进制文件..."
    cp bin/${APP_NAME}-linux-amd64 $INSTALL_DIR/bin/$APP_NAME
    chmod +x $INSTALL_DIR/bin/$APP_NAME
}

# 复制配置文件
install_config() {
    if [ -f "configs/config.example.yaml" ]; then
        if [ ! -f "/etc/$SERVICE_NAME/config.yaml" ]; then
            print_info "复制配置文件模板..."
            cp configs/config.example.yaml /etc/$SERVICE_NAME/config.yaml
            print_warn "请编辑 /etc/$SERVICE_NAME/config.yaml 配置文件"
        else
            print_info "配置文件已存在，跳过"
        fi
    fi
}

# 创建 systemd 服务文件
create_systemd_service() {
    print_info "创建 systemd 服务..."

    cat > /etc/systemd/system/${SERVICE_NAME}.service <<EOF
[Unit]
Description=Emby Telegram Bot Service
After=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$INSTALL_DIR/bin
ExecStart=$INSTALL_DIR/bin/$APP_NAME
ExecReload=/bin/kill -HUP \$MAINPID
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
}

# 设置权限
set_permissions() {
    print_info "设置文件权限..."
    chown -R $SERVICE_USER:$SERVICE_USER $INSTALL_DIR
    chown -R $SERVICE_USER:$SERVICE_USER /etc/$SERVICE_NAME
    chmod 600 /etc/$SERVICE_NAME/config.yaml 2>/dev/null || true
}

# 卸载服务
uninstall() {
    echo "=================================="
    echo "Emby Telegram Bot 卸载脚本"
    echo "=================================="
    echo ""

    check_root

    print_warn "即将卸载 $SERVICE_NAME"
    read -p "是否继续? [y/N] " confirm
    if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
        print_info "取消卸载"
        exit 0
    fi

    # 停止并禁用服务
    if systemctl is-active --quiet $SERVICE_NAME; then
        print_info "停止服务..."
        systemctl stop $SERVICE_NAME
    fi

    if systemctl is-enabled --quiet $SERVICE_NAME 2>/dev/null; then
        print_info "禁用服务..."
        systemctl disable $SERVICE_NAME
    fi

    # 删除 systemd 服务文件
    if [ -f "/etc/systemd/system/${SERVICE_NAME}.service" ]; then
        print_info "删除 systemd 服务文件..."
        rm -f /etc/systemd/system/${SERVICE_NAME}.service
        systemctl daemon-reload
    fi

    # 删除安装目录
    if [ -d "$INSTALL_DIR" ]; then
        print_warn "是否删除数据目录 $INSTALL_DIR? [y/N]"
        read -p "(包含所有数据和日志) " confirm_data
        if [ "$confirm_data" = "y" ] || [ "$confirm_data" = "Y" ]; then
            print_info "删除安装目录..."
            rm -rf $INSTALL_DIR
        else
            print_info "保留数据目录: $INSTALL_DIR"
        fi
    fi

    # 删除配置目录
    if [ -d "/etc/$SERVICE_NAME" ]; then
        print_warn "是否删除配置目录 /etc/$SERVICE_NAME? [y/N]"
        read -p "(包含配置文件) " confirm_config
        if [ "$confirm_config" = "y" ] || [ "$confirm_config" = "Y" ]; then
            print_info "删除配置目录..."
            rm -rf /etc/$SERVICE_NAME
        else
            print_info "保留配置目录: /etc/$SERVICE_NAME"
        fi
    fi

    # 删除用户
    if id "$SERVICE_USER" &>/dev/null; then
        print_warn "是否删除系统用户 $SERVICE_USER? [y/N]"
        read -p "" confirm_user
        if [ "$confirm_user" = "y" ] || [ "$confirm_user" = "Y" ]; then
            print_info "删除系统用户..."
            userdel $SERVICE_USER 2>/dev/null || true
        else
            print_info "保留系统用户: $SERVICE_USER"
        fi
    fi

    echo ""
    print_info "卸载完成！"
    echo ""
}

# 显示使用说明
show_usage() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  install     安装服务 (默认)"
    echo "  uninstall   卸载服务"
    echo "  -h, --help  显示此帮助信息"
    echo ""
}

# 主安装流程
install() {
    echo "=================================="
    echo "Emby Telegram Bot 安装脚本"
    echo "=================================="
    echo ""

    check_root
    check_system
    check_systemd

    print_info "开始安装..."

    create_user
    create_directories
    install_binary
    install_config
    create_systemd_service
    set_permissions

    echo ""
    print_info "安装完成！"
    echo ""
    echo "后续步骤:"
    echo "  1. 编辑配置文件: vim /etc/$SERVICE_NAME/config.yaml"
    echo "  2. 启动服务:     systemctl start $SERVICE_NAME"
    echo "  3. 查看状态:     systemctl status $SERVICE_NAME"
    echo "  4. 查看日志:     journalctl -u $SERVICE_NAME -f"
    echo "  5. 开机自启:     systemctl enable $SERVICE_NAME"
    echo ""
}

# 主函数
main() {
    case "${1:-install}" in
        install)
            install
            ;;
        uninstall)
            uninstall
            ;;
        -h|--help)
            show_usage
            ;;
        *)
            print_error "未知选项: $1"
            show_usage
            exit 1
            ;;
    esac
}

main "$@"
