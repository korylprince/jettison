package main

import (
	"syscall"
)

func hide(path string) error {

	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	attrs, err := syscall.GetFileAttributes(p)
	if err != nil {
		return err
	}
	attrs = attrs | syscall.FILE_ATTRIBUTE_HIDDEN
	err = syscall.SetFileAttributes(p, attrs)
	if err != nil {
		return err
	}
	return nil
}
