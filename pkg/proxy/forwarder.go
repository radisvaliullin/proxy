package proxy

import (
	"context"
	"io"
	"net"
	"time"
)

// forwards conn stream from one source to another
// sessCancel - cancel session (unlock connections, call close)
// hbDur - heartbeat duration, define heartbeat time interval
// if read/write operation not active longer than heartbeat interval function trigger conn session close
func streamForwarderWithHeartbeat(sessCancel context.CancelFunc, in, out net.Conn, hbDur time.Duration, buffSize int) error {
	// read/write err channel
	rwErrChan := make(chan error)
	// use reader with tick to notify about read/write activity
	inWithTicker, rTicker := newReaderWithTicker(in)

	// read/write goroutine
	go func() {
		buff := make([]byte, buffSize)
		if _, err := io.CopyBuffer(out, inWithTicker, buff); err != nil {
			rwErrChan <- err
		}
		// close signals that goroutine closed
		close(rwErrChan)
	}()

	// track aliveness of read/write
	err := func() error {
		hbTm := time.NewTimer(hbDur)
		defer hbTm.Stop()
		for {
			select {
			case <-rTicker:
				if !hbTm.Stop() {
					<-hbTm.C
				}
				hbTm.Reset(hbDur)
			case <-hbTm.C:
				// unlock connections
				sessCancel()
				return ErrForwardHeartBeat
			case err := <-rwErrChan:
				return err
			}
		}
	}()
	// read r/w error if not yet read
	// blocked until r/w goroutine end
	<-rwErrChan
	return err
}
