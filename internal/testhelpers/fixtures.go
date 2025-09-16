package testhelpers

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetProjectRoot returns the project root directory
// This helps tests find migrations and static files regardless of where they're run from
func GetProjectRoot() string {
	_, b, _, _ := runtime.Caller(0)
	d := filepath.Dir(b)
	// Go up from internal/testhelpers to project root
	return filepath.Dir(filepath.Dir(d))
}

// GetMigrationsPath returns the correct path to migrations directory
func GetMigrationsPath() string {
	root := GetProjectRoot()
	return "file://" + filepath.Join(root, "migrations")
}

// GetStaticPath returns the correct path to static directory
func GetStaticPath() string {
	root := GetProjectRoot()
	return filepath.Join(root, "static")
}

// TestInProjectRoot changes to project root for test execution
func TestInProjectRoot() func() {
	originalWd, _ := os.Getwd()
	if err := os.Chdir(GetProjectRoot()); err != nil {
		panic("Could not change dir to project root")
	}
	return func() {
		if err := os.Chdir(originalWd); err != nil {
			panic("Could not change back to original dir")
		}
	}
}
