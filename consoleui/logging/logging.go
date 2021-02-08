package logging

import (
	"fmt"
	"godtop/consoleui/config"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"syscall"
)

const LOGFILE = "errors.log"

func New(config config.Config) (io.WriteCloser, error) {
	//create log directory
	cache := config.ConfigDir.QueryCacheFolder()
	err := cache.MkdirAll()
	if err != nil && !os.IsExist(err) {
		return nil, err
	}

	w := &RotateWriter{
		filename:   filepath.Join(cache.Path, LOGFILE),
		maxLogSize: 50000,
	}
	err = w.rotate()
	if err != nil {
		return nil, err
	}

	fmt.Printf("Log file: %v", w.filename)

	log.SetFlags(log.Ltime | log.Lshortfile)
	log.SetOutput(w)

	stderrToLogfile(w.fp)
	return w, nil
}

type RotateWriter struct {
	lock       sync.Mutex
	filename   string
	fp         *os.File
	maxLogSize int64
}

func (w *RotateWriter) Close() error {
	return w.fp.Close()
}

// Write satisfies the io.Writer interface.
func (w *RotateWriter) Write(output []byte) (int, error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	// Rotate if the log hits the size limit
	s, err := os.Stat(w.filename)
	if err == nil {
		if s.Size() > w.maxLogSize {
			w.rotate()
		}
	}
	return w.fp.Write(output)
}

// Perform the actual act of rotating and reopening file.
func (w *RotateWriter) rotate() (err error) {
	// Close existing file if open
	if w.fp != nil {
		err = w.fp.Close()
		w.fp = nil
		if err != nil {
			return
		}
	}
	// This will keep three logs
	for i := 1; i > -1; i-- {
		from := fmt.Sprintf("%s.%d", w.filename, i)
		to := fmt.Sprintf("%s.%d", w.filename, i+1)
		// Rename dest file if it already exists
		_, err = os.Stat(from)
		if err == nil {
			err = os.Rename(from, to)
			if err != nil {
				return
			}
		}
	}
	// Rename dest file if it already exists
	_, err = os.Stat(w.filename)
	if err == nil {
		err = os.Rename(w.filename, fmt.Sprintf("%s.%d", w.filename, 0))
		if err != nil {
			return
		}
	}

	// open the log file
	w.fp, err = os.OpenFile(w.filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0660)
	if err != nil {
		return fmt.Errorf("cannot open log file: %v", err.Error())
	}

	return nil
}

func stderrToLogfile(logfile *os.File) {
	syscall.Dup2(int(logfile.Fd()), 2)
}
