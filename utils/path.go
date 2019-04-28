package utils

import (
	"os"
	"path/filepath"
)

/*

 */
func GetMainDirectory() string {
	exePath := os.Args[0]
	exePath, _ = filepath.Abs(exePath)

	return filepath.Dir(exePath) + "/"
}

/*

 */
func GetExeFileName() string {
	exePath := os.Args[0]
	exePath, _ = filepath.Abs(exePath)

	return filepath.Base(exePath)
}