package utils

import (
	"io"
)

var _ io.ReadCloser = (*StreamHijacker)(nil)

type StreamHijacker struct {
	Buffer   []byte
	Offset   int
	Upstream io.ReadCloser
	fileType string
	Size     int
	OnClose  func(*StreamHijacker)
}

func NewStreamHijacker(size int, fileType string, upstream io.ReadCloser) *StreamHijacker {
	return &StreamHijacker{
		Buffer:   make([]byte, size),
		Offset:   0,
		Upstream: upstream,
		fileType: fileType,
		Size:     size,
	}
}

func (s *StreamHijacker) Read(p []byte) (n int, err error) {
	read, err := s.Upstream.Read(p)
	copy(s.Buffer[s.Offset:], p[:read])
	s.Offset += read
	return read, err
}

func (s *StreamHijacker) Close() error {
	if s.OnClose != nil && s.Offset == s.Size {
		s.OnClose(s)
	}
	return s.Upstream.Close()
}

func (s *StreamHijacker) FileType() string {
	if s.fileType == "" {
		// TODO Read the first few bytes and guess the file type from upstream
		// if s.Offset >= 512 {
		// 	s.fileType = http.DetectContentType(s.Buffer[:512])
		// }

		// Fallback
		s.fileType = "application/octet-stream"
	}

	return s.fileType
}
