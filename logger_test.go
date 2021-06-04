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
	log_http := NewHttp(
		10,
		1,
		NewUrls("http://localhost"),
		MessageKB_t{},
		DefaultClient(DefaultTransport(time.Second), time.Second),
	)
	logger.AddOutput("http", 0, log_http)

	Debug("lalala %s", ByteSize(1024))
	Debug("bububu %s", ByteSize(2048))
	// Output:
	// DEBUG lalala 1.00 KB
	// DEBUG bububu 2.00 KB
	//
}

func TestLog1(t *testing.T) {

}
