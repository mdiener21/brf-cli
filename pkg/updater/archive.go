package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func extractBinary(archivePath, destination, goos string) error {
	if goos == "windows" {
		return extractZipBinary(archivePath, destination, executableName(goos))
	}
	return extractTarBinary(archivePath, destination, executableName(goos))
}

func extractTarBinary(archivePath, destination, name string) error {
	archive, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer archive.Close()

	gzipReader, err := gzip.NewReader(archive)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	reader := tar.NewReader(gzipReader)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if header.Typeflag != tar.TypeReg || filepath.Base(header.Name) != name {
			continue
		}
		return writeExtractedBinary(destination, reader)
	}
	return fmt.Errorf("archive does not contain %s", name)
}

func extractZipBinary(archivePath, destination, name string) error {
	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, file := range archive.File {
		if !file.Mode().IsRegular() || filepath.Base(file.Name) != name {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			return err
		}
		err = writeExtractedBinary(destination, reader)
		closeErr := reader.Close()
		if err != nil {
			return err
		}
		return closeErr
	}
	return fmt.Errorf("archive does not contain %s", name)
}

func writeExtractedBinary(destination string, source io.Reader) error {
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
	if err != nil {
		return err
	}

	written, copyErr := io.Copy(file, io.LimitReader(source, maxArchiveSize+1))
	closeErr := file.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if written > maxArchiveSize {
		return fmt.Errorf("executable exceeds %d bytes", maxArchiveSize)
	}
	return nil
}
