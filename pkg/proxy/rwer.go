package proxy

import "io"

var _ io.Reader = (*readerWithTicker)(nil)

// notify each time when next bytes read done
type readerWithTicker struct {
	reader io.Reader
	tick   chan struct{}
}

// return reader and ticker channel
// ticker will update after each read bytes done
func newReaderWithTicker(r io.Reader) (nr io.Reader, t <-chan struct{}) {
	ticker := make(chan struct{}, 1)
	rwt := &readerWithTicker{
		reader: r,
		tick:   ticker,
	}
	return rwt, ticker
}

func (r *readerWithTicker) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	// non blocking send
	select {
	case r.tick <- struct{}{}:
	default:
	}
	return n, err
}
