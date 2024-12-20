package helpers

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

func SaveUploadedFile(file multipart.File, header *multipart.FileHeader, uploadDir string, userId uint) (string, error) {
	filename := fmt.Sprintf("%d_%s", userId, filepath.Base(header.Filename))

	err := os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	// Define the file path
	filePath := filepath.Join(uploadDir, filename)

	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Copy the uploaded file's content to the destination file
	_, err = io.Copy(out, file)
	if err != nil {
		return "", err
	}

	return filePath, nil
}
