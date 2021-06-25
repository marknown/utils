package utils

import (
	"bytes"
	"os/exec"
	"runtime"
)

// Exec 执行执行的命令
func Exec(args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	var output bytes.Buffer
	cmd.Stdout = &output
	err := cmd.Run()
	if nil != err {
		return "", err
	}

	return output.String(), nil
}

// ExecInOS 在不同的环境中执行命令
func ExecInOS(cmd string) (string, error) {
	var args = []string{}
	if runtime.GOOS == "windows" {
		args = append(args, "cmd", "/c")
	} else {
		args = append(args, "/bin/sh", "-c")
	}

	args = append(args, cmd)
	return Exec(args...)
}
