//go:build windows
// +build windows

package main

import "syscall"

// setProcessGroupID 在 Windows 上是一个空操作，因为进程管理模型不同。
func setProcessGroupID(attr *syscall.SysProcAttr) {
	// 不需要任何操作
}
