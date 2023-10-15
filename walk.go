package go_walk

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DirectoryInfo holds metadata about a directory.
type DirectoryInfo struct {
	Path            string    // Absolute path of the directory.
	Size            int64     // Size of the directory in bytes.
	CreationTime    time.Time // When the directory was created.
	LastModified    time.Time // When the directory was last modified.
	NumberOfFiles   int       // Number of files in the directory.
	NumberOfSubdirs int       // Number of subdirectories within the directory.
}

// ListDirStat lists directories matching the provided keywords in dirPath
// and returns their metadata. If no keywords are provided, all directories
// are matched.
func ListDirStat(dirPath string, keywords ...string) ([]DirectoryInfo, error) {
	pathStat, err := os.Stat(dirPath)
	if err != nil {
		return nil, err
	}

	if !pathStat.IsDir() {
		return nil, errors.New("the path provided is not a directory")
	}

	dirChan := make(chan DirectoryInfo)
	errChan := make(chan error)
	var directories []DirectoryInfo
	var mu sync.Mutex

	keywordSet := make(map[string]struct{})
	for _, keyword := range keywords {
		keywordSet[keyword] = struct{}{}
	}

	wg := &sync.WaitGroup{}

	directoryVisitor := func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			// If no keywords are provided or the directory name matches one of the keywords
			_, exists := keywordSet[entry.Name()]
			if len(keywordSet) == 0 || exists {
				wg.Add(1)

				go func(p string) {
					defer wg.Done()
					dirStat, err := calculateDirStats(p)
					if err != nil {
						errChan <- err
						return
					}
					dirChan <- dirStat
				}(path)
			}
		}
		return nil
	}

	go func() {
		err := filepath.WalkDir(dirPath, directoryVisitor)
		if err != nil {
			errChan <- err
		}
		wg.Wait()
		close(dirChan)
		close(errChan)
	}()

	var dirChanClosed, errChanClosed bool

LOOP:
	for {
		select {
		case dirStat, ok := <-dirChan:
			if !ok {
				dirChanClosed = true
				if dirChanClosed && errChanClosed {
					break LOOP
				}
				continue
			}
			mu.Lock()
			directories = append(directories, dirStat)
			mu.Unlock()
		case e, ok := <-errChan:
			if !ok {
				errChanClosed = true
				if dirChanClosed && errChanClosed {
					break LOOP
				}
				continue
			}
			// Logging errors, can be modified to handle differently.
			fmt.Printf("Error while processing directory: %v\n", e)
		}
	}

	return directories, nil
}

// calculateDirStats computes and returns the statistics for a directory.
func calculateDirStats(path string) (DirectoryInfo, error) {
	var totalSize int64
	var numberOfFiles int
	var numberOfSubdirs int
	var creationTime time.Time
	var lastModified time.Time

	err := filepath.WalkDir(path, func(_ string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
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

		if creationTime.IsZero() || info.ModTime().Before(creationTime) {
			creationTime = info.ModTime()
		}

		if lastModified.IsZero() || info.ModTime().After(lastModified) {
			lastModified = info.ModTime()
		}

		return nil
	})

	if err != nil {
		return DirectoryInfo{}, err
	}

	return DirectoryInfo{
		Path:            path,
		Size:            totalSize,
		CreationTime:    creationTime,
		LastModified:    lastModified,
		NumberOfFiles:   numberOfFiles,
		NumberOfSubdirs: numberOfSubdirs,
	}, nil
}
