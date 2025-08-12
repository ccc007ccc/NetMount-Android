#!/system/bin/sh
# 该脚本将在 Android 系统启动完毕后以服务模式运行

# --- 环境变量 ---
# 确保系统二进制文件在 PATH 中，这对于 fusermount3 至关重要
export PATH=/system/bin:$PATH
# 设置时区为亚洲/上海，避免时间显示错误
export TZ=Asia/Shanghai

MODDIR=${0%/*}

# 添加 fusermount3 到 PATH
export PATH="$MODDIR:$PATH"

# 调试信息：验证 fusermount3 二进制文件路径
FUSE_BIN_PATH="$MODDIR/fusermount3"
precheck_log "fusermount3 二进制文件路径: $FUSE_BIN_PATH"
if [ -x "$FUSE_BIN_PATH" ]; then
    precheck_log "成功: fusermount3 二进制文件存在且可执行"
else
    precheck_log "错误: fusermount3 二进制文件不存在或不可执行"
fi
precheck_log "当前 PATH: $PATH"

# 测试是否能通过 PATH 找到 fusermount3
if which fusermount3 >/dev/null 2>&1; then
    FOUND_FUSE=$(which fusermount3)
    precheck_log "成功: 通过 PATH 找到 fusermount3: $FOUND_FUSE"
else
    precheck_log "错误: 无法通过 PATH 找到 fusermount3"
fi

# 持久化配置目录
CONFIG_DIR="/data/adb/netmount"
CONFIG_FILE="$CONFIG_DIR/config.json"
LOGFILE="$CONFIG_DIR/daemon.log" # 守护进程日志
PRECHECK_LOG="$CONFIG_DIR/precheck.log" # 环境检查日志

# --- 环境准备 ---
mkdir -p "$CONFIG_DIR"
# 清理旧日志
echo "--- NetMount Boot Completed Log ---" > "$LOGFILE"
echo "--- NetMount Pre-check Log ---" > "$PRECHECK_LOG"

log_and_echo() {
    echo "$1"
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$LOGFILE"
}

precheck_log() {
    echo "$1"
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$PRECHECK_LOG"
}

# 等待系统完全就绪
log_and_echo "等待系统完全启动..."
sleep 2

log_and_echo "模块目录: $MODDIR"
log_and_echo "配置文件: $CONFIG_FILE"
log_and_echo "日志文件: $LOGFILE"

# --- 环境检查 (输出到 precheck.log) ---
precheck_log "开始环境检查..."

# 1. 检查网络连接
precheck_log "检查网络连接..."
if ping -c 1 8.8.8.8 >/dev/null 2>&1; then
    precheck_log "成功: 网络连接正常。"
else
    precheck_log "警告: 网络连接可能未就绪。"
fi

# 2. 检查 FUSE 内核模块
if [ ! -c "/dev/fuse" ]; then
    precheck_log "错误: FUSE 设备 /dev/fuse 不存在！挂载将失败。"
    precheck_log "请确保您的内核支持 FUSE。"
else
    precheck_log "成功: FUSE 设备 /dev/fuse 已找到。"
fi

# 3. 检查 /proc/filesystems 是否支持 fuse
if ! grep -q "fuse$" "/proc/filesystems"; then
    precheck_log "警告: /proc/filesystems 中未明确列出 'fuse'。"
else
    precheck_log "成功: /proc/filesystems 中已找到 'fuse'。"
fi

# 4. 检查 SELinux 状态
SELINUX_STATUS=$(getenforce)
precheck_log "信息: 当前 SELinux 状态为: $SELINUX_STATUS"
if [ "$SELINUX_STATUS" = "Enforcing" ]; then
    precheck_log "警告: SELinux 正在强制模式下运行。如果挂载失败，请检查 SELinux 策略 (auditd, logcat)。"
fi

# 5. 检查存储权限
if [ ! -d "/data/media/0" ]; then
    precheck_log "错误: 无法访问 /data/media/0，存储权限可能不足。"
else
    precheck_log "成功: 存储目录 /data/media/0 可访问。"
fi

precheck_log "环境检查完成。"

# --- 授予执行权限 ---
log_and_echo "授予可执行权限..."
chmod 755 "$MODDIR/netmountd"
chmod 755 "$MODDIR/rclone"
chmod 755 "$MODDIR/fusermount3"
log_and_echo "权限授予完成。"

# --- 启动守护进程 ---
log_and_echo "正在启动 NetMount 守护进程..."
# 将守护进程的日志重定向到 daemon.log
# 将配置文件路径作为参数传递给守护进程
exec $MODDIR/netmountd -config "$CONFIG_FILE" >> "$LOGFILE" 2>&1