package video

import "time"

type Source struct {
	ID          string
	URL         string
	Title       string
	Description string
	Uploader    string
	Duration    time.Duration
	Thumbnail   string
	Extractor   string
}

type Media struct {
	Source  Source
	Path    string
	WorkDir string
}

func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}

	return *p
}
