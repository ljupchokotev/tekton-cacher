package tar

import (
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func TarFiles(paths []string, w io.Writer) error {
	tw := tar.NewWriter(w)

	for _, path := range paths {
		// validate path
		path = filepath.Clean(path)

		walker := func(file string, finfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// fill in header info using func FileInfoHeader
			hdr, err := tar.FileInfoHeader(finfo, finfo.Name())
			if err != nil {
				return err
			}

			relFilePath := file
			if filepath.IsAbs(path) {
				relFilePath, err = filepath.Rel(path, file)
				if err != nil {
					return err
				}
			}
			// ensure header has relative file path
			hdr.Name = relFilePath

			if finfo.Mode()&fs.ModeSymlink != 0 {
				link, err := os.Readlink(relFilePath)
				if err != nil {
					return fmt.Errorf("error reading link: %w", err)
				}

				hdr.Linkname = link
			}

			if err := tw.WriteHeader(hdr); err != nil {
				return fmt.Errorf("error writing header: %w", err)
			}

			// if path is a dir, dont continue
			if finfo.Mode().IsDir() {
				return nil
			}

			if !finfo.Mode().IsRegular() {
				return nil
			}

			// add file to tar
			srcFile, err := os.Open(file)
			if err != nil {
				return fmt.Errorf("can't open %s: %w", file, err)
			}
			defer srcFile.Close()

			_, err = io.Copy(tw, srcFile)
			if err != nil {
				return fmt.Errorf("can't copy %s to tar writer: %w", file, err)
			}
			return nil
		}

		if err := filepath.Walk(path, walker); err != nil {
			fmt.Printf("failed to add %s to tar: %s\n", path, err)
		}
	}

	if err := tw.Close(); err != nil {
		return err
	}

	return nil
}

func Untar(r io.Reader) error {
	absPath, err := filepath.Abs(".")
	if err != nil {
		return err
	}

	tr := tar.NewReader(r)

	// untar each segment
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// determine proper file path info
		finfo := hdr.FileInfo()
		fileName := hdr.Name
		if filepath.IsAbs(fileName) {
			fileName, err = filepath.Rel("/", fileName)
			if err != nil {
				return err
			}
		}
		absFileName := filepath.Join(absPath, fileName)

		if finfo.Mode()&fs.ModeSymlink != 0 {
			symlinkPath := filepath.Join(".", hdr.Linkname)
			if err := os.Symlink(symlinkPath, absFileName); err != nil {
				fmt.Println(fmt.Errorf("symlink warning: %w", err))
			}
			continue
		}

		if finfo.Mode().IsDir() {
			if err := os.MkdirAll(absFileName, 0755); err != nil {
				return err
			}
			continue
		}

		// create new file with original file mode
		file, err := os.OpenFile(absFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, finfo.Mode().Perm())
		if err != nil {
			return err
		}

		n, cpErr := io.Copy(file, tr)
		if closeErr := file.Close(); closeErr != nil { // close file immediately
			return err
		}
		if cpErr != nil {
			return cpErr
		}
		if n != finfo.Size() {
			return fmt.Errorf("unexpected bytes written: wrote %d, want %d", n, finfo.Size())
		}
	}

	return nil
}
