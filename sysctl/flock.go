//go:build !windows && !illumos

package sysctl

import (
	"os"
	"syscall"

	"github.com/pkg/errors"
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
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_RDWR, 0o666)
	if err != nil {
		return errors.WithMessage(err, "os open err")
	}

	l.f = f
	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		return errors.WithMessagef(err, "flock file: %s err", l.path)
	}

	return nil
}

// Unlock 解锁
func (l *FileLock) Unlock() error {
	defer l.f.Close()

	return syscall.Flock(int(l.f.Fd()), syscall.LOCK_UN)
}

// Write 写入字节
func (l *FileLock) Write(b []byte) (n int, err error) {
	return l.f.Write(b)
}

// WriteString 写入字符串
func (l *FileLock) WriteString(s string) (n int, err error) {
	return l.f.WriteString(s)
}
