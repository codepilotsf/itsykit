package setup

import (
	"fmt"
	"io"
	"sync"
	"time"
)

type commandProgress struct {
	once          sync.Once
	done, stopped chan struct{}
}

func newCommandProgress(out io.Writer, terminal bool, label string) *commandProgress {
	p := &commandProgress{done: make(chan struct{}), stopped: make(chan struct{})}
	if !terminal {
		fmt.Fprintln(out, label)
		close(p.stopped)
		return p
	}
	fmt.Fprint(out, label+" |")
	go func() {
		defer close(p.stopped)
		ticker := time.NewTicker(120 * time.Millisecond)
		defer ticker.Stop()
		frames := []byte{'/', '-', '\\', '|'}
		i := 0
		for {
			select {
			case <-p.done:
				fmt.Fprint(out, "\r\x1b[2K")
				return
			case <-ticker.C:
				fmt.Fprintf(out, "\r%s %c", label, frames[i%len(frames)])
				i++
			}
		}
	}()
	return p
}
func (p *commandProgress) stop() { p.once.Do(func() { close(p.done); <-p.stopped }) }

type progressOutput struct {
	io.Writer
	stop func()
}

func (w progressOutput) Write(b []byte) (int, error) { w.stop(); return w.Writer.Write(b) }
