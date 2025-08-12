# Build Instructions

## Build Environment Requirements

- **Python 3.6+** - For running build scripts
- **Go 1.18+** - For compiling the daemon
- **Node.js 16+** - For building the frontend
- **npm** - Node.js package manager

## Build Steps

### 1. Clone Project

```bash
git clone <repository-url>
cd NetMount-Android
```

### 2. Prepare Binary Dependencies

Ensure the `bin/` directory contains:
- `arm64-rclone` - rclone ARM64 executable
- `fusermount3` - FUSE mount tool

Download from:
- rclone: https://rclone.org/downloads/
- fusermount3: Copy from Linux system or compile

### 3. Run Build Script

```bash
python build.py
```

The build script will automatically:
- Clean old build files
- Install frontend dependencies and build Vue.js app
- Compile Go daemon (target: Linux ARM64)
- Copy web files to webroot
- Copy binary dependencies
- Create KernelSU module zip package

### 4. Get Build Results

After building, find in `build/` directory:
```
build/
└── NetMount-Android.zip    # KernelSU module package
```

## Manual Build Steps

For manual building:

### 1. Build Frontend

```bash
cd webcode
npm install
npm run build
cd ..
```

### 2. Build Daemon

```bash
cd daemon-src
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o ../ksu-model/netmountd -ldflags="-s -w" .
cd ..
```

### 3. Copy Files

```bash
# Copy frontend files
cp -r webcode/dist ksu-model/webroot

# Copy binary files
cp bin/arm64-rclone ksu-model/rclone
cp bin/fusermount3 ksu-model/fusermount3
```

### 4. Package Module

```bash
cd ksu-model
zip -r ../NetMount-Android.zip .
cd ..
```

## Development Environment

### Frontend Development

```bash
cd webcode
npm install
npm run dev  # Start development server
```

Frontend dev server starts at `http://localhost:5173`.

### Backend Development

```bash
cd daemon-src
go run . -config ./config.json
```

Daemon starts HTTP service on `:8088`.

## File Structure Explanation

```
NetMount-Android/
├── daemon-src/              # Go daemon source code
│   ├── main.go             # Main program
│   ├── process_linux.go    # Linux-specific process handling
│   ├── process_windows.go  # Windows-specific process handling
│   └── go.mod              # Go module dependencies
├── webcode/                # Vue.js frontend source
│   ├── src/                # Source directory
│   ├── package.json        # Dependencies configuration
│   └── vite.config.js      # Vite build configuration
├── ksu-model/              # KernelSU module files
│   ├── module.prop         # Module properties
│   ├── service.sh          # Startup script
│   ├── boot-completed.sh   # Boot completion script
│   └── webroot/            # Web files (generated during build)
├── bin/                    # Pre-compiled binary files
└── build.py               # Build script
```

## Custom Builds

### Modify Target Architecture

Edit configuration in `build.py`:

```python
GO_ARCH = "arm64"  # Can change to arm, amd64, etc.
GO_OS = "linux"    # Target operating system
```

### Add Compile Options

Modify Go compile parameters in `build_daemon()`:

```python
cmd = [GO_EXECUTABLE, "build", "-o", DAEMON_OUTPUT_PATH, "-ldflags=-s -w", "."]
```

## Troubleshooting

### Build Failures

#### 1. Go Compilation Failed
- Check Go version is 1.18+
- Ensure network connectivity (downloading dependencies)
- Verify CGO_ENABLED=0 setting

#### 2. Frontend Build Failed
- Check Node.js version is 16+
- Delete `node_modules` and reinstall dependencies
- Check network connectivity (npm registry access)

#### 3. Missing Binary Files
- Ensure `bin/arm64-rclone` exists and is executable
- Ensure `bin/fusermount3` exists and is executable

### Runtime Issues

#### 1. Permission Issues
- Ensure module runs with root privileges
- Check SELinux settings

#### 2. Mount Failures
- Check rclone version compatibility
- Ensure FUSE kernel module is loaded