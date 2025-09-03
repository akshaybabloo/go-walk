package go_walk

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ErrorList holds a list of errors.
type ErrorList []error

// Error returns a string representation of the error list.
func (e ErrorList) Error() string {
	if len(e) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d errors occurred:\n", len(e)))
	for _, err := range e {
		sb.WriteString(fmt.Sprintf("\t- %s\n", err.Error()))
	}
	return sb.String()
}

// DirectoryInfo holds metadata about a directory.
type DirectoryInfo struct {
	Path            string    // Absolute path of the directory.
	Size            int64     // Size of the directory in bytes.
	LastModified    time.Time // When the directory was last modified.
	NumberOfFiles   int       // Number of files in the directory.
	NumberOfSubdirs int       // Number of subdirectories within the directory.
}

// ListDirStat lists directories matching the provided keywords in dirPath
// and returns their metadata. If no keywords are provided, all directories
// are matched. Returns aggregated errors if they occur.
func ListDirStat(dirPath string, keywords ...string) ([]DirectoryInfo, error) {
	pathStat, err := os.Stat(dirPath)
	if err != nil {
		return nil, err
	}

	if !pathStat.IsDir() {
		return nil, errors.New("the path provided is not a directory")
	}

	const numWorkers = 8
	workChan := make(chan string)
	dirChan := make(chan DirectoryInfo)
	errChan := make(chan error)
	var directories []DirectoryInfo
	var errs ErrorList
	var mu sync.Mutex

	keywordSet := make(map[string]struct{})
	for _, keyword := range keywords {
		keywordSet[keyword] = struct{}{}
	}

	wg := &sync.WaitGroup{}

	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range workChan {
				dirStat, err := calculateDirStats(path)
				if err != nil {
					errChan <- err
					continue
				}
				dirChan <- dirStat
			}
		}()
	}

	directoryVisitor := func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			_, exists := keywordSet[entry.Name()]
			if len(keywordSet) == 0 || exists {
				workChan <- path
			}
		}
		return nil
	}

	go func() {
		err := filepath.WalkDir(dirPath, directoryVisitor)
		if err != nil {
			errChan <- err
		}
		close(workChan)
		wg.Wait()
		close(dirChan)
		close(errChan)
	}()

	for dirStat := range dirChan {
		mu.Lock()
		directories = append(directories, dirStat)
		mu.Unlock()
	}

	for e := range errChan {
		mu.Lock()
		errs = append(errs, e)
		mu.Unlock()
	}

	if len(errs) > 0 {
		return directories, errs
	}

	return directories, nil
}

// calculateDirStats computes and returns the statistics for a directory.
func calculateDirStats(path string) (DirectoryInfo, error) {
	var totalSize int64
	var numberOfFiles int
	var numberOfSubdirs int

	info, err := os.Stat(path)
	if err != nil {
		return DirectoryInfo{}, err
	}
	lastModified := info.ModTime()

	err = filepath.WalkDir(path, func(subPath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself from being counted as a subdirectory
		if path == subPath {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		if entry.IsDir() {
			numberOfSubdirs++
		} else {
			totalSize += info.Size()
			numberOfFiles++
		}

		return nil
	})

	if err != nil {
		return DirectoryInfo{}, err
	}

	return DirectoryInfo{
		Path:            path,
		Size:            totalSize,
		LastModified:    lastModified,
		NumberOfFiles:   numberOfFiles,
		NumberOfSubdirs: numberOfSubdirs,
	}, nil
}
