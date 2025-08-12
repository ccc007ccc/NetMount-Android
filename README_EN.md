# NetMount-Android - Network Mount

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![KernelSU](https://img.shields.io/badge/KernelSU-Compatible-green.svg)](https://kernelsu.org/)

## Introduction

NetMount-Android is a KernelSU module designed for Android devices that allows users to conveniently mount various network storage services to the local file system through a Web UI. Supports multiple protocols including WebDAV, SFTP, and SMB.

## Key Features

- 🌐 **Web UI Management** - Clean and intuitive web control panel
- 📁 **Multi-protocol Support** - WebDAV, SMB, FTP and other mainstream network storage protocols
- 🚀 **Auto Mount** - Automatically mount configured network storage on boot
- 🔒 **Smart Waiting** - Automatically detect device unlock status, delay mounting user data areas
- 📱 **KernelSU Integration** - Perfect integration with KernelSU module system
- ⚡ **Real-time Logs** - View mount status and error information in real-time

## System Requirements

- Android device (ARM64 architecture)
- KernelSU installed
- Root permissions

## Quick Start

### 1. Download Module

Download the latest `NetMount-Android.zip` from the [Releases page](../../releases).

### 2. Install Module

1. Open KernelSU Manager
2. Go to "Modules" page
3. Click "+" button to install the downloaded zip file
4. Reboot device

### 3. Access Web UI

After device reboot, open browser and visit: `http://localhost:8088`

### 4. Configure Network Storage

1. Add network storage configuration in Web UI
2. Set mount point path (Recommended: `/sdcard/NetMount/ServiceName`)
3. Fill in server information and authentication details
4. Click mount button

## Supported Protocols

| Protocol | Description | Example Address |
|----------|-------------|-----------------|
| WebDAV | HTTP-based file transfer protocol | `https://example.com/webdav` |
| SMB | Windows network sharing protocol | `192.168.1.100/share` |
| FTP | File Transfer Protocol | `ftp://192.168.1.100` |

## Configuration Examples

### WebDAV Configuration

- **Server Address**: `https://cloud.example.com/remote.php/dav/files/username/`
- **Auth Type**: Password Authentication
- **Username**: `your_username`
- **Password**: `your_password`
- **Mount Point**: `/sdcard/NetMount/WebDAV`

### SMB Configuration

- **Server Address**: `192.168.1.100/Documents`
- **Auth Type**: Password Authentication
- **Username**: `samba_user`
- **Password**: `samba_password`
- **Mount Point**: `/sdcard/NetMount/SMB`

## Directory Structure

```
NetMount-Android/
├── daemon-src/          # Go daemon source code
├── webcode/            # Vue.js Web UI source code
├── ksu-model/          # KernelSU module files
├── bin/                # Pre-compiled binary files
├── build.py            # Build script
└── docs/              # Detailed documentation
```

## Documentation

- [Build Instructions](docs/BUILD_EN.md) - Detailed build and development instructions
- [Troubleshooting](docs/TROUBLESHOOTING_EN.md) - Common issues and solutions
- [API Documentation](docs/API_EN.md) - REST API interface documentation

## Contributing

Issues and Pull Requests are welcome!

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Other Languages**: [中文](README.md)