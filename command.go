package utils

import (
	"bytes"
	"os/exec"
	"runtime"
	"strings"
)

// Exec Run 执行执行的命令（阻塞直到完成）
func ExecRun(args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	var output bytes.Buffer
	cmd.Stdout = &output
	err := cmd.Run()

	if nil != err {
		return "", err
	}

	return output.String(), nil
}

// Exec Start 执行执行的命令（不会等待该命令完成即返回）
func ExecStart(args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	err := cmd.Start()

	if nil != err {
		return "", err
	}

	return "异步执行无输出", nil
}

// ExecInOS 在不同的环境中执行命令
func ExecInOS(cmd string) (string, error) {
	var args = []string{}
	if runtime.GOOS == "windows" {
		args = append(args, "cmd", "/c")
	} else {
		args = append(args, "/bin/sh", "-c")
	}

	cmd = strings.TrimSpace(cmd)
	args = append(args, cmd)

	// 如果命令以 & 调用异步执行命令
	if strings.HasSuffix(cmd, "&") {
		return ExecStart(args...)
	}

	return ExecRun(args...)
}
