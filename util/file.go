package util

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ngageoint/seed-common/constants"
)

//GetFullPath returns the full path of the given file. This expands relative file
// paths and verifes non-relative paths
// Validate path for file existance??
func GetFullPath(rFile, directory string) string {
	if !filepath.IsAbs(rFile) {

		// Test relative to given directory
		if directory != "." && directory != "" {

			var err error
			path := filepath.Join(directory, rFile)
			path, err = filepath.Abs(filepath.ToSlash(path))

			// see if resulting path exists
			if _, err = os.Stat(path); !os.IsNotExist(err) {
				rFile = path

				// see if parent directory exists. If so, assume this directory will be created
			} else if _, err := os.Stat(filepath.Dir(path)); !os.IsNotExist(err) {
				rFile = path
			}

		} else {
			// Test relative to current directory
			curDir, _ := os.Getwd()
			dir := filepath.Join(curDir, rFile)
			if _, err := os.Stat(dir); !os.IsNotExist(err) {
				rFile = filepath.Clean(dir)

				// see if parent directory exists. If so, assume this directory will be created
			} else if _, err := os.Stat(filepath.Dir(dir)); !os.IsNotExist(err) {
				rFile = filepath.Clean(dir)
			}
		}
	}

	return rFile
}

//DockerfileRegistry attempts to find the registry for a dockerfile's base image, if any
func DockerfileBaseRegistry(dir string) (string, error) {
	registry := ""

	// Define the current working directory
	curDirectory, _ := os.Getwd()

	dockerfile := "Dockerfile"
	if dir == "." {
		dockerfile = filepath.Join(curDirectory, dockerfile)
	} else {
		if filepath.IsAbs(dir) {
			dockerfile = filepath.Join(dir, dockerfile)
		} else {
			dockerfile = filepath.Join(curDirectory, dir, dockerfile)
		}
	}

	// Verify dockerfile exists within specified directory.
	_, err := os.Stat(dockerfile)
	if os.IsNotExist(err) {
		PrintUtil("ERROR: %s cannot be found.\n",
			dockerfile)
		PrintUtil("Make sure you have specified the correct directory.\n")

		return "", err
	}

	file, err := os.Open(dockerfile)
	if err == nil {
		// make sure it gets closed
		defer file.Close()

		// create a new scanner and read the file line by line
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			i := strings.Index(line, "/")
			if i > 5 && strings.HasPrefix(line, "FROM ") {
				registry = line[5:i]
				break
			}
		}

		// check for errors
		err = scanner.Err()
	}

	return registry, err
}

//GetSeedFileName Finds and returns the full filepath to the seed.manifest.json
// The second return value indicates whether it exists or not.
func GetSeedFileName(dir string) (string, bool, error) {
	// Define the current working directory
	curDirectory, _ := os.Getwd()

	// set path to seed file -
	// 	Either relative to current directory or located in given directory
	//  	-d directory might be a relative path to current directory
	seedFileName := constants.SeedFileName
	if dir == "." {
		seedFileName = filepath.Join(curDirectory, seedFileName)
	} else {
		if filepath.IsAbs(dir) {
			seedFileName = filepath.Join(dir, seedFileName)
		} else {
			seedFileName = filepath.Join(curDirectory, dir, seedFileName)
		}
	}

	// Check to see if seed.manifest.json exists within specified directory.
	_, err := os.Stat(seedFileName)
	return seedFileName, !os.IsNotExist(err), err
}

//SeedFileName Finds and returns the full filepath to the seed.manifest.json
func SeedFileName(dir string) (string, error) {
	seedFileName, exists, err := GetSeedFileName(dir)
	if !exists {
		PrintUtil("ERROR: %s cannot be found.\n",
			seedFileName)
		PrintUtil("Make sure you have specified the correct directory.\n")
	}

	return seedFileName, err
}

//RemoveAllFiles removes all files in the specified directory
func RemoveAllFiles(v string) {
	err := os.RemoveAll(v)
	if err != nil {
		PrintUtil("Error removing directory: %s\n", err.Error())
	}
}

//ReadLinesFromFile reads all lines from the given file as string array
func ReadLinesFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

//CopyFiles Copies the file specified by copyfrom to copyto
func CopyFiles(copyfrom, copyto string) (bool, error) {
	cfi, err := os.Stat(copyfrom)
	if err != nil {
		PrintUtil("ERROR: Error opening %s for copy\n%s\n", copyfrom, err.Error())
		return false, err
	}

	if !cfi.IsDir() {
		PrintUtil("%s is a file.. copying to %s\n", copyfrom, copyto)
		return CopyFile(copyfrom, copyto)
	}

	// create destination directories
	if _, err = os.Stat(copyto); os.IsNotExist(err) {
		PrintUtil("%s does not exist... creating\n", copyto)
		err = os.MkdirAll(copyto, cfi.Mode())
		if err != nil {
			PrintUtil("ERROR: Error creating destination directory %s for copy\n%s\n", copyto, err.Error())
			return false, err
		}
	}

	files, err := ioutil.ReadDir(copyfrom)
	if err != nil {
		PrintUtil("Error reading %s\n%s\n", copyfrom, err.Error())
	}
	copied := false
	PrintUtil("files: %v\n", files)
	for _, f := range files {
		sfp := filepath.Join(copyfrom, f.Name())
		dfp := filepath.Join(copyto, f.Name())
		if f.IsDir() {
			if empty, _ := IsEmpty(f.Name()); empty {
				err = os.Mkdir(dfp, f.Mode())
			} else {
				copied, err = CopyFiles(sfp, dfp)
			}
			if err != nil {
				PrintUtil("Error copying directory %s\n%s\n", sfp, err.Error())
			}
		} else {
			copied, err = CopyFile(sfp, dfp)
			if err != nil {
				PrintUtil("Error copying file %s\n%s\n", sfp, err.Error())
			}
		}
	}

	PrintUtil("Made it through for loop with args %s, %s; returning %v, %v\n", copyfrom, copyto, copied, err)
	return copied, nil
}

//CopyFile copies the file copyfrom to the filel copyto
func CopyFile(copyfrom, copyto string) (bool, error) {

	var from, to *os.File
	var err error
	if from, err = os.Open(copyfrom); err != nil {
		PrintUtil("ERROR: Error opening %s for copy\n%s\n",
			copyfrom, err.Error())
		return false, err
	}	
	defer from.Close()


	if to, err = os.OpenFile(copyto, os.O_RDWR | os.O_CREATE, 0666); err != nil {
		PrintUtil("ERROR: Error creating %s for copy\n%s\n",
			copyto, err.Error())
		return false, err
	}	
	defer to.Close()

	if _, err  = io.Copy(to, from); err != nil {
		PrintUtil("ERROR: Error copying %s to %s\n%s\n",
			copyfrom, copyto, err.Error())
	}

	return true, nil
}

//IsEmpty returns true if specified directory is empty, false otherwise
func IsEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
        return false, err
    }
    defer f.Close()

    _, err = f.Readdirnames(1) // Or f.Readdir(1)
    if err == io.EOF {
        return true, nil
    }
    return false, err // Either not empty or error, suits both cases
}

//Remove deletes the specified file/directory
func Remove(name string) error {
	return os.Remove(name)
}