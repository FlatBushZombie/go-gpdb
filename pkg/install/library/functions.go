package library

import (
	"os"
	"path/filepath"
	"io"
	log "../../core/logger"
	"regexp"
	"archive/zip"
)

import (
	"../../core/arguments"
	"../../core/methods"
)

// Provides all the files in the directory
func ListFilesInDir(dest string) []string {
	log.Println("Listing all the files in the download directory: " + dest)
	files, err := filepath.Glob(dest + "*")
	methods.Fatal_handler(err)
	return files
}


// Find the file that matches the search string
func GetBinaryOfMatchingVersion(files []string) string {

	// Loop through all the files and see if we can find a binaries that matches with the version
	log.Println("Checking if there is binaries that matches the version to install: " + arguments.RequestedInstallVersion)
	for _, f := range files {
		pattern := "greenplum.*" + arguments.RequestedInstallVersion
		matched, _ := regexp.MatchString(pattern, f)

		// If we find a match then return the file name
		if matched {
			return f
		}
	}
	return ""
}

// Unzip the provided file.
func Unzip(src, dest string) error {

	// unzip the file
	log.Println("Unzipping the zip file: " + src)
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}
