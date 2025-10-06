package testhelpers

import (
	"path/filepath"
	"runtime"
)

func GetProjectRoot() string {
	_, b, _, _ := runtime.Caller(0)
	d := filepath.Dir(b)
	return filepath.Dir(filepath.Dir(d))
}

func GetMigrationsPath() string {
	root := GetProjectRoot()
	return "file://" + filepath.Join(root, "migrations")
}
