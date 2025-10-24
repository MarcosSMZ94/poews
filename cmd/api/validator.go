package main

import "os"

func validateFilePath(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return err
	}
	if err != nil {
		return err
	}
	if info.IsDir() {
		return os.ErrInvalid
	}
	return nil
}
