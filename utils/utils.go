package utils

import "fmt"

func ByteCountToHuman(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPEZY"[exp])
}

/*
function humanFileSize(size) {
    if (size < 1024) return size + ' B'
    let i = Math.floor(Math.log(size) / Math.log(1024))
    let num = (size / Math.pow(1024, i))
    let round = Math.round(num)
    num = round < 10 ? num.toFixed(2) : round < 100 ? num.toFixed(1) : round
    return `${num} ${'KMGTPEZY'[i-1]}B`
}
*/
