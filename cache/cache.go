package cache

import "time"

type CachedInstance struct {
	RelativePath string
	AbsolutePath string
	Data         []byte
	ContentType  string
	LoadTime     time.Time
}

type Cache interface {
	Index() (int64, error)
	Get(path []byte) (string, []byte)
}
