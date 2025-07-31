package go_walk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListDirStat(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "test-list-dir-stat-*")
	assert.NoError(t, err)
	defer func(path string) {
		err := os.RemoveAll(path)
		assert.NoError(t, err)
	}(tmpDir)

	project1 := filepath.Join(tmpDir, "project1")
	nodeModules1 := filepath.Join(project1, "node_modules")
	project2 := filepath.Join(tmpDir, "project2")
	srcDir := filepath.Join(project2, "src")
	nodeModules2 := filepath.Join(project2, "node_modules")
	nestedNodeModules := filepath.Join(srcDir, "node_modules")

	err = os.MkdirAll(nodeModules1, 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(nodeModules2, 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(nestedNodeModules, 0755)
	assert.NoError(t, err)

	// Add a file to nodeModules1 to test the size attribute
	testFilePath := filepath.Join(nodeModules1, "test.txt")
	err = os.WriteFile(testFilePath, []byte("test content"), 0644)
	assert.NoError(t, err)

	// Call ListDirStat
	directories, err := ListDirStat(tmpDir, "node_modules")
	assert.NoError(t, err)

	// Check the results
	assert.Len(t, directories, 3)

	foundDirs := make(map[string]bool)
	for _, dir := range directories {
		foundDirs[dir.Path] = true
		switch dir.Path {
		case nodeModules1:
			assert.Equal(t, int64(12), dir.Size) // "test content" has 12 bytes
			assert.Equal(t, 1, dir.NumberOfFiles)
			assert.Equal(t, 0, dir.NumberOfSubdirs)
		case nodeModules2:
			assert.Equal(t, int64(0), dir.Size)
			assert.Equal(t, 0, dir.NumberOfFiles)
			assert.Equal(t, 0, dir.NumberOfSubdirs)
		case nestedNodeModules:
			assert.Equal(t, int64(0), dir.Size)
			assert.Equal(t, 0, dir.NumberOfFiles)
			assert.Equal(t, 0, dir.NumberOfSubdirs)
		default:
			t.Fatalf("Unexpected directory path: %s", dir.Path)
		}
	}
	assert.True(t, foundDirs[nodeModules1])
	assert.True(t, foundDirs[nodeModules2])
	assert.True(t, foundDirs[nestedNodeModules])
}

func TestListDirStatNoKeyword(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "test-list-dir-stat-*")
	assert.NoError(t, err)
	defer func(path string) {
		err := os.RemoveAll(path)
		assert.NoError(t, err)
	}(tmpDir)

	project1 := filepath.Join(tmpDir, "project1")
	nodeModules1 := filepath.Join(project1, "node_modules")
	project2 := filepath.Join(tmpDir, "project2")
	srcDir := filepath.Join(project2, "src")
	nodeModules2 := filepath.Join(project2, "node_modules")
	nestedNodeModules := filepath.Join(srcDir, "node_modules")

	err = os.MkdirAll(nodeModules1, 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(nodeModules2, 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(nestedNodeModules, 0755)
	assert.NoError(t, err)

	// Add a file to nodeModules1 to test the size attribute
	testFilePath := filepath.Join(nodeModules1, "test.txt")
	err = os.WriteFile(testFilePath, []byte("test content"), 0644)
	assert.NoError(t, err)

	// Call ListDirStat without a keyword
	directories, err := ListDirStat(tmpDir)
	assert.NoError(t, err)

	// Check the results
	foundDirs := make(map[string]bool)
	for _, dir := range directories {
		foundDirs[dir.Path] = true
		switch dir.Path {
		case tmpDir:
			assert.Equal(t, int64(12), dir.Size)
			assert.Equal(t, 1, dir.NumberOfFiles)
			assert.Equal(t, 6, dir.NumberOfSubdirs)
		case project1:
			assert.Equal(t, int64(12), dir.Size)
			assert.Equal(t, 1, dir.NumberOfFiles)
			assert.Equal(t, 1, dir.NumberOfSubdirs)
		case nodeModules1:
			assert.Equal(t, int64(12), dir.Size)
			assert.Equal(t, 1, dir.NumberOfFiles)
			assert.Equal(t, 0, dir.NumberOfSubdirs)
		case project2:
			assert.Equal(t, int64(0), dir.Size)
			assert.Equal(t, 0, dir.NumberOfFiles)
			assert.Equal(t, 3, dir.NumberOfSubdirs)
		case srcDir:
			assert.Equal(t, int64(0), dir.Size)
			assert.Equal(t, 0, dir.NumberOfFiles)
			assert.Equal(t, 1, dir.NumberOfSubdirs)
		case nodeModules2:
			assert.Equal(t, int64(0), dir.Size)
			assert.Equal(t, 0, dir.NumberOfFiles)
			assert.Equal(t, 0, dir.NumberOfSubdirs)
		case nestedNodeModules:
			assert.Equal(t, int64(0), dir.Size)
			assert.Equal(t, 0, dir.NumberOfFiles)
			assert.Equal(t, 0, dir.NumberOfSubdirs)
		default:
			t.Fatalf("Unexpected directory path: %s", dir.Path)
		}
	}

	assert.Len(t, directories, 7)
	assert.True(t, foundDirs[tmpDir])
	assert.True(t, foundDirs[project1])
	assert.True(t, foundDirs[nodeModules1])
	assert.True(t, foundDirs[project2])
	assert.True(t, foundDirs[srcDir])
	assert.True(t, foundDirs[nodeModules2])
	assert.True(t, foundDirs[nestedNodeModules])
}