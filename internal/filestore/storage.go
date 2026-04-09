package filestore



import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type FileStorage interface {
	Save(filename string, data io.Reader) (string, error)

	Get(path string) (io.ReadCloser, error)

	Delete(path string) error

	Exists(path string) (bool, error)
}

type LocalFileStorage struct {
	basePath string
}

func NewLocalFileStorage(basePath string) (*LocalFileStorage, error) {

	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalFileStorage{
		basePath: basePath,
	}, nil
}

func (l *LocalFileStorage) Save(filename string, data io.Reader) (string, error) {

	ext := filepath.Ext(filename)

	fileUUID := uuid.New().String()
	newFilename := fileUUID + ext

	filePath := filepath.Join(l.basePath, newFilename)

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, data); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filePath, nil
}

func (l *LocalFileStorage) Get(path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(l.basePath, path)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

func (l *LocalFileStorage) Delete(path string) error {
	fullPath := filepath.Join(l.basePath, path)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", path)
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (l *LocalFileStorage) Exists(path string) (bool, error) {
	fullPath := filepath.Join(l.basePath, path)

	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("failed to check file existence: %w", err)
}

