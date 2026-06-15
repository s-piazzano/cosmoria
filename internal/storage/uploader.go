package storage

import (
	"fmt"
	"io"
	"net/http"
)

const defaultMaxSize = 50 << 20

type UploadInfo struct {
	Filename   string
	MimeType   string
	Reader     io.Reader
	Size       int64
}

func ParseUpload(r *http.Request, maxSize int64) (*UploadInfo, error) {
	if maxSize <= 0 {
		maxSize = defaultMaxSize
	}

	if err := r.ParseMultipartForm(maxSize); err != nil {
		return nil, fmt.Errorf("upload: parse form: %w", err)
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("upload: get file: %w", err)
	}
	defer file.Close()

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return &UploadInfo{
		Filename: header.Filename,
		MimeType: mimeType,
		Reader:   file,
		Size:     header.Size,
	}, nil
}
