package cache

import (
	"github.com/Vilsol/yeet/source"
	"io"
	"time"
)

type CachedInstance struct {
	RelativePath string
	AbsolutePath string
	Data         []byte
	ContentType  string
	LoadTime     time.Time
}

type commonInstance struct {
	Instance *CachedInstance
	Get      func(instance *commonInstance, host []byte) (string, io.Reader, int)
}

type KeyValue struct {
	Key   string
	Value *commonInstance
}

type Cache interface {
	Index() (int64, error)
	Get(path []byte, host []byte) (string, io.Reader, int)
	Source() source.Source
	Iter() <-chan KeyValue
	Store(path []byte, host []byte, instance *commonInstance)
}
