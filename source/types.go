package source

import (
	"github.com/Vilsol/yeet/utils"
)

type IndexFunc = func(absolutePath string, cleanedPath string) int64

type Source interface {
	Get(path string, host []byte) *utils.StreamHijacker
	IndexPath(dir string, f IndexFunc) (int64, int64, error)
	Watch() (<-chan WatchEvent, error)
}

type WatchOp int

const (
	WatchCreate WatchOp = iota
	WatchModify
	WatchDelete
	WatchRename
)

type WatchEvent struct {
	CleanPath string
	AbsPath   string
	Op        WatchOp
}
