package log

import (
	"testing"
	"time"
)

func Example_log1() {
	logger := NewEmpty()
	SetLogger(logger)

	logger.AddOutput("stdout", 0, NewStdout(&DT_t{}))
	log_file, _ := NewFileBytes("/tmp/test.log", &DT_t{}, 1024, 10)
	logger.AddOutput("file", 0, log_file)
	log_http := NewHttp(DefaultTransport(time.Second), 10, 1, "http://localhost", Message_t{}, time.Second, nil)
	logger.AddOutput("http", 0, log_http)

	Debug("lalala")
	Debug("bububu")
	// Output:
	// DEBUG lalala
	// DEBUG bububu
	//
}

func TestLog1(t *testing.T) {

}
