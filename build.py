# -*- coding: utf-8 -*-
import os
import shutil
import zipfile
import subprocess

# --- 配置 ---
GO_EXECUTABLE = "go"
GO_ARCH = "arm64"
GO_OS = "linux"

# --- 路径定义 ---
ROOT_DIR = os.path.dirname(os.path.abspath(__file__))
DAEMON_SRC_DIR = os.path.join(ROOT_DIR, "daemon-src")
WEBCODE_DIR = os.path.join(ROOT_DIR, "webcode")
BUILD_DIR = os.path.join(ROOT_DIR, "build")
KSU_MODEL_DIR = os.path.join(ROOT_DIR, "ksu-model")
WEBROOT_DIR = os.path.join(KSU_MODEL_DIR, "webroot")
DAEMON_OUTPUT_PATH = os.path.join(KSU_MODEL_DIR, "netmountd")
RCLONE_SOURCE_PATH = os.path.join(ROOT_DIR, "bin", "arm64-rclone")
RCLONE_OUTPUT_PATH = os.path.join(KSU_MODEL_DIR, "rclone")
FUSERMOUNT3_SOURCE_PATH = os.path.join(ROOT_DIR, "bin", "fusermount3")
FUSERMOUNT3_OUTPUT_PATH = os.path.join(KSU_MODEL_DIR, "fusermount3")

def run_command(cmd, cwd, env=None):
    """通用函数，用于运行命令并检查错误"""
    print(f"在 '{cwd}' 中运行命令: {' '.join(cmd)}")
    # 在 Windows 上，npm 命令通常是 .cmd 文件，需要 shell=True
    process = subprocess.run(cmd, cwd=cwd, capture_output=True, text=True, shell=True, encoding='utf-8', env=env)
    if process.returncode != 0:
        print(f"命令执行失败: {' '.join(cmd)}")
        print("--- STDOUT ---")
        print(process.stdout)
        print("--- STDERR ---")
        print(process.stderr)
        exit(1)
    # print("命令成功执行。") # 减少不必要的输出

def clean():
    """清理旧的构建产物"""
    print("--- 正在清理构建目录 ---")
    if os.path.exists(BUILD_DIR):
        shutil.rmtree(BUILD_DIR)
    if os.path.exists(WEBROOT_DIR):
        shutil.rmtree(WEBROOT_DIR)
    if os.path.exists(DAEMON_OUTPUT_PATH):
        os.remove(DAEMON_OUTPUT_PATH)
    if os.path.exists(RCLONE_OUTPUT_PATH):
        os.remove(RCLONE_OUTPUT_PATH)
    # 清理 fusermount3 文件
    if os.path.exists(FUSERMOUNT3_OUTPUT_PATH):
        os.remove(FUSERMOUNT3_OUTPUT_PATH)
    # 清理模块压缩包
    zip_files = [f for f in os.listdir(ROOT_DIR) if f.endswith('.zip')]
    for zf in zip_files:
        os.remove(os.path.join(ROOT_DIR, zf))
    print("清理完成。")

def build_frontend():
    """构建 Vue.js 前端"""
    print("--- 正在构建前端 ---")
    run_command(["npm", "install"], WEBCODE_DIR)
    run_command(["npm", "run", "build"], WEBCODE_DIR)
    print("前端构建完成。")

def build_daemon():
    """编译 Go 守护进程"""
    print("--- 正在编译 Go 守护进程 ---")
    env = os.environ.copy()
    env["GOOS"] = GO_OS
    env["GOARCH"] = GO_ARCH
    env["CGO_ENABLED"] = "0"
    cmd = [GO_EXECUTABLE, "build", "-o", DAEMON_OUTPUT_PATH, "-ldflags=-s -w", "."]
    run_command(cmd, DAEMON_SRC_DIR, env=env)
    print(f"守护进程已编译到: {DAEMON_OUTPUT_PATH}")

def copy_web_files():
    """将构建好的前端文件复制到 webroot"""
    print("--- 正在复制前端文件 ---")
    frontend_dist_dir = os.path.join(WEBCODE_DIR, "dist")
    if not os.path.exists(frontend_dist_dir):
        print(f"错误: 前端构建目录 '{frontend_dist_dir}' 不存在。")
        exit(1)
    shutil.copytree(frontend_dist_dir, WEBROOT_DIR)
    print(f"前端文件已复制到: {WEBROOT_DIR}")

def copy_fusermount3_binary():
    """复制 fusermount3 二进制文件"""
    print("--- 正在复制 fusermount3 ---")
    if not os.path.exists(FUSERMOUNT3_SOURCE_PATH):
        print(f"错误: 未找到 fusermount3 源文件: {FUSERMOUNT3_SOURCE_PATH}")
        exit(1)
    shutil.copy(FUSERMOUNT3_SOURCE_PATH, FUSERMOUNT3_OUTPUT_PATH)
    print(f"fusermount3 已复制到: {FUSERMOUNT3_OUTPUT_PATH}")

def copy_rclone_binary():
    """复制 rclone 可执行文件"""
    print("--- 正在复制 rclone ---")
    if not os.path.exists(RCLONE_SOURCE_PATH):
        print(f"错误: 未找到 rclone 源文件: {RCLONE_SOURCE_PATH}")
        exit(1)
    shutil.copy(RCLONE_SOURCE_PATH, RCLONE_OUTPUT_PATH)
    print(f"rclone 已复制到: {RCLONE_OUTPUT_PATH}")

def get_module_id():
    """从 module.prop 文件中读取模块 ID"""
    prop_path = os.path.join(KSU_MODEL_DIR, "module.prop")
    with open(prop_path, 'r', encoding='utf-8') as f:
        for line in f:
            if line.strip().startswith('id='):
                return line.strip().split('=', 1)[1]
    print("错误: 在 module.prop 中未找到模块 ID。")
    exit(1)

def create_zip():
    """创建 KernelSU 模块的 zip 压缩包"""
    print("--- 正在创建模块 zip 包 ---")
    if not os.path.exists(BUILD_DIR):
        os.makedirs(BUILD_DIR)
    
    module_id = get_module_id()
    zip_name = f"{module_id}.zip"
    zip_path = os.path.join(BUILD_DIR, zip_name)

    with zipfile.ZipFile(zip_path, 'w', zipfile.ZIP_DEFLATED) as zf:
        for root, _, files in os.walk(KSU_MODEL_DIR):
            for file in files:
                file_path = os.path.join(root, file)
                archive_path = os.path.relpath(file_path, KSU_MODEL_DIR)
                
                # 为关键文件设置可执行权限
                if file in ["netmountd", "rclone", "service.sh", "fusermount3"]:
                    # 0o755 for rwxr-xr-x -> (S_IRWXU | S_IRGRP | S_IXGRP | S_IROTH | S_IXOTH)
                    attr = (0o755 << 16)
                    info = zipfile.ZipInfo(archive_path)
                    info.external_attr = attr
                    with open(file_path, 'rb') as f:
                        zf.writestr(info, f.read(), zipfile.ZIP_DEFLATED)
                else:
                    zf.write(file_path, archive_path)
    
    print(f"模块已成功打包到: {zip_path}")

def main():
    """主构建函数"""
    clean()
    build_frontend()
    build_daemon()
    copy_web_files()
    copy_rclone_binary()
    copy_fusermount3_binary()
    create_zip()
    print("\n构建成功！")

if __name__ == "__main__":
    main()