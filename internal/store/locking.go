package store

import (
	"fmt"
	"os"
	"time"
)

type FileLock struct {
	path    string
	timeout time.Duration
	poll    time.Duration
}

func NewFileLock(path string) *FileLock {
	return &FileLock{path: path, timeout: 5 * time.Second, poll: 100 * time.Millisecond}
}

func (l *FileLock) Lock() error {
	deadline := time.Now().Add(l.timeout)
	for {
		file, err := os.OpenFile(l.path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			fmt.Fprintf(file, "%d", time.Now().Unix())
			file.Close()
			return nil
		}
		if !os.IsExist(err) {
			return err
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for lock at %s", l.path)
		}
		time.Sleep(l.poll)
	}
}

func (l *FileLock) Unlock() error {
	return os.Remove(l.path)
}
