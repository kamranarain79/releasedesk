package build

// DownloadStat counts successful downloads of one artifact checksum.
type DownloadStat struct {
	Sha256 string `xorm:"pk varchar(64)"`
	Count  int64  `xorm:"not null default 0"`
}

type DownloadRepository interface {
	Increment(sha256 string) error
	Sum(checksums []string) (int64, error)
}
