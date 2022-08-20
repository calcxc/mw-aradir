package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	cp "github.com/otiai10/copy"
)

func copyItem(f fs.DirEntry, dest string) {
	// Create full directory path
	path := filepath.Join(dest, f.Name())
	err := os.MkdirAll(path, os.ModeDir|os.ModePerm)
	checkError(err)

	rc, err := os.Open(path) // For read access.
	checkError(err)
	if !f.IsDir() {
		// Use os.Create() since Zip don't store file permissions.
		fileCopy, err := os.Create(path)
		checkError(err)
		_, err = io.Copy(fileCopy, rc)
		fileCopy.Close()
		checkError(err)
	}
	rc.Close()
}

// unused for now
func CopyFolders(folderpath, dest string) {
	files, err := os.ReadDir(folderpath)
	checkError(err)
	newPath := fmt.Sprint(folderpath, "/")
	err = os.MkdirAll(newPath, 0755)
	checkError(err)
	for _, f := range files {
		err := cp.Copy(fmt.Sprint(folderpath, "/", f.Name()), fmt.Sprint(dest, "/", f.Name()))
		checkError(err)
	}
}

func Copy(folderpath, dest string, filename string) {
	files, err := os.ReadDir(folderpath)
	checkError(err)
	newPath := fmt.Sprint(folderpath, "/")
	err = os.MkdirAll(newPath, 0755)
	checkError(err)
	for _, f := range files {
		if f.Name() == filename {
			err := cp.Copy(fmt.Sprint(folderpath, "/", f.Name()), fmt.Sprint(dest, "/", f.Name()))
			checkError(err)
		}
	}
}
