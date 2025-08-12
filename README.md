# NetMount-Android 网络挂载

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![KernelSU](https://img.shields.io/badge/KernelSU-Compatible-green.svg)](https://kernelsu.org/)

## 简介

NetMount-Android 是一个为 Android 设备设计的 KernelSU 模块，允许用户通过 Web UI 便捷地挂载各种网络存储服务到本地文件系统。支持 WebDAV、SFTP、SMB 等多种协议。

**其他语言版本**：[English](README_EN.md)

## 主要特性

- 🌐 **Web UI 管理界面** - 简洁直观的网页控制面板
- 📁 **多协议支持** - WebDAV、SMB、FTP 等主流网络存储协议
- 🚀 **自动挂载** - 开机自动挂载配置的网络存储
- 🔒 **智能等待** - 自动检测设备解锁状态，延迟挂载用户数据区域
- 📱 **KernelSU 集成** - 完美集成 KernelSU 模块系统
- ⚡ **实时日志** - 实时查看挂载状态和错误信息

## 系统要求

- Android 设备（ARM64 架构）
- 已安装 KernelSU
- Root 权限

## 快速开始

### 1. 下载模块

从 [Releases 页面](../../releases) 下载最新的 `NetMount-Android.zip` 文件。

### 2. 安装模块

1. 打开 KernelSU Manager
2. 选择 "模块" 页面
3. 点击 "+" 按钮安装下载的 zip 文件
4. 重启设备

### 3. 访问 Web UI

设备重启后，用浏览器访问：`http://localhost:8088`

### 4. 配置网络存储

1. 在 Web UI 中添加网络存储配置
2. 设置挂载点路径（推荐：`/sdcard/NetMount/服务名`）
3. 填写服务器信息和认证信息
4. 点击挂载按钮

## 支持的协议

| 协议 | 描述 | 示例地址 |
|------|------|----------|
| WebDAV | 基于 HTTP 的文件传输协议 | `https://example.com/webdav` |
| SMB | Windows 网络共享协议 | `192.168.1.100/share` |
| FTP | 文件传输协议 | `ftp://192.168.1.100` |

## 配置示例

### WebDAV 配置

- **服务器地址**：`https://cloud.example.com/remote.php/dav/files/username/`
- **认证类型**：密码认证
- **用户名**：`your_username`
- **密码**：`your_password`
- **挂载点**：`/sdcard/NetMount/WebDAV`

### SMB 配置

- **服务器地址**：`192.168.1.100/Documents`
- **认证类型**：密码认证
- **用户名**：`samba_user`
- **密码**：`samba_password`
- **挂载点**：`/sdcard/NetMount/SMB`

## 目录结构

```
NetMount-Android/
├── daemon-src/          # Go 守护进程源码
├── webcode/            # Vue.js Web UI 源码
├── ksu-model/          # KernelSU 模块文件
├── bin/                # 预编译二进制文件
├── build.py            # 构建脚本
└── docs/              # 详细文档
```

## 文档

- [构建说明](docs/BUILD_CN.md) - 详细的构建和开发说明
- [故障排除](docs/TROUBLESHOOTING_CN.md) - 常见问题和解决方案
- [API 文档](docs/API_CN.md) - REST API 接口文档

## 贡献

欢迎提交 Issue 和 Pull Request！


