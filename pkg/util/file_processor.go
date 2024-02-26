package util

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

type fileMetadata struct {
	height   int
	width    int
	fileType string
}

func Resize(ctx context.Context, file multipart.File, opts fileMetadata) ([]byte, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("end of file error %w", err)
		}
		return nil, err
	}

	content := strings.Split(http.DetectContentType(data), "/")
	contentType := content[0]
	if contentType != opts.fileType {
		return nil, fmt.Errorf("invalid file expected %+v", opts.fileType)
	}

	return nil, nil
}
