#!/system/bin/sh
# 该脚本将在 late_start service 模式下执行
# 现在只做基础环境准备，实际启动由 boot_completed.sh 处理

# --- 环境变量 ---
# 确保系统二进制文件在 PATH 中，这对于 fusermount3 至关重要
export PATH=/system/bin:$PATH
# 设置时区为亚洲/上海，避免时间显示错误
export TZ=Asia/Shanghai

MODDIR=${0%/*}

# 添加 fusermount3 到 PATH
export PATH="$MODDIR:$PATH"

# 持久化配置目录
CONFIG_DIR="/data/adb/netmount"
CONFIG_FILE="$CONFIG_DIR/config.json"
LOGFILE="$CONFIG_DIR/service.log" # service 阶段日志

# --- 环境准备 ---
mkdir -p "$CONFIG_DIR"
echo "--- NetMount Service Preparation Log ---" > "$LOGFILE"

log_and_echo() {
    echo "$1"
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$LOGFILE"
}

log_and_echo "模块目录: $MODDIR"
log_and_echo "配置目录: $CONFIG_DIR"

# --- 授予执行权限 ---
log_and_echo "授予可执行权限..."
chmod 755 "$MODDIR/netmountd"
chmod 755 "$MODDIR/rclone"
chmod 755 "$MODDIR/fusermount3"
chmod 755 "$MODDIR/boot_completed.sh"
log_and_echo "权限授予完成。"

log_and_echo "service.sh 执行完成，等待 boot_completed.sh 启动守护进程。"
