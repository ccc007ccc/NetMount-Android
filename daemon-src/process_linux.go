//go:build linux
// +build linux

package main

import "syscall"

// setProcessGroupID 设置进程组 ID，以便我们可以独立于父进程管理它。
// 这在 Linux 上是必需的，可以防止守护进程在父脚本退出时被杀死。
func setProcessGroupID(attr *syscall.SysProcAttr) {
	attr.Setpgid = true
}
