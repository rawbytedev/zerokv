package zerokv

type ScanOption func(*scanConfig)

type scanConfig struct {
	Reverse  bool
	Prefetch bool
	Limit    int
}

func NewScanConfig() *scanConfig {
	return &scanConfig{}
}
func WithReverse() ScanOption {
	return func(sc *scanConfig) {
		sc.Reverse = true
	}
}

func WithPrefetch() ScanOption {
	return func(sc *scanConfig) {
		sc.Prefetch = true
	}
}
