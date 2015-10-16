//
// Log rotator
//

package log_go

import "os"
import "io"
import "fmt"
import "sync"

type NoStderr_t struct {}
type Stderr_t struct {}

func (* NoStderr_t) Write(m []byte) (int, error) {
	return 0, nil
}

func (* Stderr_t) Write(m []byte) (int, error) {
	return os.Stderr.Write(m)
}

type RotateLogWriter struct {
	mx sync.Mutex
	fp * os.File
	filename string
	max_bytes int
	curr_bytes int
	backup_count int
	stderr io.Writer
}

func (self * RotateLogWriter) Write(m []byte) (n int, err error) {
	self.mx.Lock()
	defer self.mx.Unlock()
	if self.curr_bytes + len(m) >= self.max_bytes {
		self.LogCycle()
	}
	n, err = self.fp.Write(m)
	self.stderr.Write(m)
	self.curr_bytes += n
	return
}

func (self * RotateLogWriter) LogCycle() {
	if self.fp != nil {
		self.fp.Close()
	}
	os.Remove(fmt.Sprintf("%s.%d", self.filename, self.backup_count))
	for i := self.backup_count; i > 1; i-- {
		os.Rename(fmt.Sprintf("%s.%d", self.filename, i - 1), fmt.Sprintf("%s.%d", self.filename, i))
	}
	os.Rename(self.filename, fmt.Sprintf("%s.%d", self.filename, 1))
	self.curr_bytes = 0
	var err error
	self.fp, err = os.OpenFile(self.filename, os.O_WRONLY | os.O_CREATE /*| os.O_APPEND*/, 0644)
	if err != nil {
		self.stderr = &Stderr_t{}
		fmt.Fprintln(self.stderr, err.Error())
	}
}

func NewRotateLogWriter(filename string, max_bytes int, backup_count int, stderr bool) (log * RotateLogWriter) {
	var duplicate io.Writer
	if stderr {
		duplicate = &Stderr_t{}
	} else {
		duplicate = &NoStderr_t{}
	}
	log = &RotateLogWriter{filename: filename, max_bytes: max_bytes, backup_count: backup_count, stderr: duplicate}
	log.LogCycle()
	return
}
