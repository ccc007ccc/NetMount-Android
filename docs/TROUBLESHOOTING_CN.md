# 故障排除

## 常见问题

### 安装问题

**Q: KernelSU Manager 无法安装模块**
- 确认下载的是 `.zip` 格式的模块文件
- 检查 KernelSU 版本是否为最新版本
- 尝试重启 KernelSU Manager 应用
- 检查设备存储空间是否充足

**Q: 安装后重启无效果**
- 确认 KernelSU 正常工作（其他模块是否正常）
- 检查 `/data/adb/modules/NetMount-Android` 目录是否存在
- 查看 KernelSU 日志是否有错误信息

### 访问问题

**Q: 无法访问 Web UI (http://localhost:8088)**
- 等待 2-3 分钟让服务完全启动
- 检查设备是否已连接网络
- 尝试使用 `http://127.0.0.1:8088`
- 检查防火墙或安全软件是否拦截

**Q: Web UI 界面空白或加载失败**
- 清除浏览器缓存
- 尝试使用不同浏览器
- 检查设备时间是否正确
- 重启模块：在 KernelSU Manager 中禁用后重新启用

### 挂载问题

**Q: 挂载失败，显示认证错误**
- 检查用户名和密码是否正确
- 确认远程服务器地址格式正确
- 对于 SMB：确认共享名称正确
- 对于 WebDAV：确认 URL 路径完整

**Q: 挂载成功但目录为空**
- 检查网络连接是否稳定
- 确认远程目录确实包含文件
- 等待 5-10 秒让远程内容加载
- 检查远程服务器权限设置

**Q: 挂载到 /sdcard 目录失败**
- 确保设备已完全解锁（锁屏密码已输入）
- 等待设备启动完成后再尝试挂载
- 检查存储权限是否正常

### 性能问题

**Q: 文件访问速度很慢**
- 检查网络连接质量
- 尝试更换挂载参数配置
- 考虑使用有线网络替代 WiFi
- 检查远程服务器性能

**Q: 设备运行卡顿**
- 减少同时挂载的网络存储数量
- 避免在网络存储中运行大文件操作
- 监控设备内存使用情况

## 网络存储特定问题

### SMB/CIFS

**常见错误和解决方案：**

- **"Access denied"**: 检查用户权限，确认账户有访问共享的权限
- **"Network unreachable"**: 检查 IP 地址和网络连接
- **"Protocol negotiation failed"**: 可能是 SMB 版本兼容问题，联系管理员

**SMB 地址格式：**
```
正确: 192.168.1.100/sharename
错误: \\192.168.1.100\sharename
错误: smb://192.168.1.100/sharename
```

### WebDAV

**常见错误和解决方案：**

- **SSL 证书错误**: 
  - 使用 `http://` 替代 `https://` 进行测试
  - 或联系管理员修复证书问题
- **401 Unauthorized**: 检查用户名密码，确认账户状态正常
- **404 Not Found**: 检查 WebDAV 路径是否正确

**WebDAV 地址格式示例：**
```
Nextcloud: https://your-domain.com/remote.php/dav/files/username/
ownCloud: https://your-domain.com/remote.php/webdav/
```

### FTP

**常见错误和解决方案：**

- **连接超时**: 检查防火墙设置，确认 FTP 端口开放
- **被动模式问题**: 大多数情况下被动模式更稳定
- **匿名访问失败**: 确认服务器支持匿名访问

## 日志分析

### 查看实时日志

在 Web UI 的 "日志" 页面可以查看实时运行状态：

- ✅ **绿色信息**: 正常操作
- ⚠️ **黄色警告**: 非致命问题，但需要注意
- ❌ **红色错误**: 严重问题，需要处理

### 常见日志信息

**正常启动：**
```
--- NetMount 守护进程启动 ---
使用配置文件: /data/adb/netmount/config.json
服务器正在监听 :8088...
```

**挂载成功：**
```
✅ 成功启动 'MyCloud' 的挂载进程, PID: 12345
✅ 挂载验证成功: '/sdcard/NetMount/MyCloud' 包含 15 个项目
```

**常见错误信息：**
- `rclone obscure 命令执行失败`: rclone 二进制文件问题
- `创建挂载点目录失败`: 权限或存储空间问题
- `required key not available`: 设备加密未解锁

## 高级故障排除

### 手动检查服务状态

通过 ADB 或终端模拟器：

```bash
# 检查进程是否运行
ps | grep netmountd

# 检查端口监听
netstat -tlnp | grep 8088

# 检查挂载点
mount | grep rclone

# 查看模块文件
ls -la /data/adb/modules/NetMount-Android/
```

### 重置配置

如果配置损坏，可以删除配置文件重新开始：

```bash
# 通过 ADB
adb shell rm /data/adb/netmount/config.json

# 或者在终端模拟器中
su
rm /data/adb/netmount/config.json
```

### 完全重装模块

1. 在 KernelSU Manager 中卸载模块
2. 重启设备
3. 重新安装模块 zip 文件
4. 再次重启设备

## 获取帮助

如果以上方法无法解决问题：

1. **收集信息**：
   - 设备型号和 Android 版本
   - KernelSU 版本
   - 模块版本
   - 详细错误信息和日志

2. **提交 Issue**：
   - GitHub Issues: [项目地址]/issues
   - 包含完整的错误日志
   - 描述复现步骤

3. **社区支持**：
   - 相关 Android 论坛
   - KernelSU 官方群组