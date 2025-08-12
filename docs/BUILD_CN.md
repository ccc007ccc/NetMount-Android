# 构建说明

## 构建环境要求

- **Python 3.6+** - 用于运行构建脚本
- **Go 1.18+** - 用于编译守护进程
- **Node.js 16+** - 用于构建前端界面
- **npm** - Node.js 包管理器

## 构建步骤

### 1. 克隆项目

```bash
git clone <repository-url>
cd NetMount-Android
```

### 2. 准备二进制依赖

确保 `bin/` 目录包含以下文件：
- `arm64-rclone` - rclone ARM64 版本可执行文件
- `fusermount3` - FUSE 挂载工具

可以从以下地址下载：
- rclone: https://rclone.org/downloads/
- fusermount3: 从 Linux 系统复制或编译

### 3. 运行构建脚本

```bash
python build.py
```

构建脚本将自动执行以下步骤：
- 清理旧的构建文件
- 安装前端依赖并构建 Vue.js 应用
- 编译 Go 守护进程（目标：Linux ARM64）
- 复制 Web 文件到 webroot
- 复制二进制依赖文件
- 创建 KernelSU 模块 zip 包

### 4. 获取构建结果

构建完成后，在 `build/` 目录下会生成：
```
build/
└── NetMount-Android.zip    # KernelSU 模块安装包
```

## 手动构建步骤

如果需要手动构建，可以按以下步骤：

### 1. 构建前端

```bash
cd webcode
npm install
npm run build
cd ..
```

### 2. 构建守护进程

```bash
cd daemon-src
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o ../ksu-model/netmountd -ldflags="-s -w" .
cd ..
```

### 3. 复制文件

```bash
# 复制前端文件
cp -r webcode/dist ksu-model/webroot

# 复制二进制文件
cp bin/arm64-rclone ksu-model/rclone
cp bin/fusermount3 ksu-model/fusermount3
```

### 4. 打包模块

```bash
cd ksu-model
zip -r ../NetMount-Android.zip .
cd ..
```

## 开发环境

### 前端开发

```bash
cd webcode
npm install
npm run dev  # 启动开发服务器
```

前端开发服务器会在 `http://localhost:5173` 启动。

### 后端开发

```bash
cd daemon-src
go run . -config ./config.json
```

守护进程会在 `:8088` 端口启动 HTTP 服务。

## 文件结构说明

```
NetMount-Android/
├── daemon-src/              # Go 守护进程源码
│   ├── main.go             # 主程序
│   ├── process_linux.go    # Linux 特定进程处理
│   ├── process_windows.go  # Windows 特定进程处理
│   └── go.mod              # Go 模块依赖
├── webcode/                # Vue.js 前端源码
│   ├── src/                # 源码目录
│   ├── package.json        # 依赖配置
│   └── vite.config.js      # Vite 构建配置
├── ksu-model/              # KernelSU 模块文件
│   ├── module.prop         # 模块属性
│   ├── service.sh          # 启动脚本
│   ├── boot-completed.sh   # 启动完成脚本
│   └── webroot/            # Web 文件（构建时生成）
├── bin/                    # 预编译二进制文件
└── build.py               # 构建脚本
```

## 自定义构建

### 修改目标架构

编辑 `build.py` 中的配置：

```python
GO_ARCH = "arm64"  # 可改为 arm, amd64 等
GO_OS = "linux"    # 目标操作系统
```

### 添加编译选项

在 `build_daemon()` 函数中修改 Go 编译参数：

```python
cmd = [GO_EXECUTABLE, "build", "-o", DAEMON_OUTPUT_PATH, "-ldflags=-s -w", "."]
```

## 故障排除

### 构建失败

#### 1. Go 编译失败
- 检查 Go 版本是否为 1.18+
- 确认网络连接正常（下载依赖）
- 检查 CGO_ENABLED=0 设置

#### 2. 前端构建失败
- 检查 Node.js 版本是否为 16+
- 删除 `node_modules` 重新安装依赖
- 检查网络连接（npm 仓库访问）

#### 3. 缺少二进制文件
- 确认 `bin/arm64-rclone` 存在且可执行
- 确认 `bin/fusermount3` 存在且可执行

### 运行时问题

#### 1. 权限问题
- 确认模块以 root 权限运行
- 检查 SELinux 设置

#### 2. 挂载失败
- 检查 rclone 版本兼容性
- 确认 FUSE 内核模块加载