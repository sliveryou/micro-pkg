//go:build illumos

package sysctl

import (
	"os"
)

// FileLock 文件锁
type FileLock struct {
	path string
	f    *os.File
}

// NewFileLock 新建文件锁
func NewFileLock(path string) *FileLock {
	return &FileLock{
		path: path,
	}
}

// Lock 上锁
func (l *FileLock) Lock() error {
	return nil
}

// Unlock 解锁
func (l *FileLock) Unlock() error {
	return nil
}

// Write 写入字节
func (l *FileLock) Write(b []byte) (n int, err error) {
	return 0, nil
}

// WriteString 写入字符串
func (l *FileLock) WriteString(s string) (n int, err error) {
	return 0, nil
}
