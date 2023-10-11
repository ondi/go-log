package log

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"gotest.tools/assert"
)

func Test1(t *testing.T) {
	logger := NewEmpty()
	SetLogger(logger)

	var buf bytes.Buffer
	ts := time.Now()
	logger.AddOutput("stdout", NewStdout([]Prefixer{&DT_t{}}), LOG_TRACE.Levels)
	logger.AddOutput("buf", NewStdany([]Prefixer{&DT_t{}}, &buf), LOG_TRACE.Levels)
	log_file, _ := NewFileBytes(ts, "/tmp/test.log", []Prefixer{&DT_t{}}, 1024, 10)
	logger.AddOutput("file", log_file, LOG_TRACE.Levels)
	log_http := NewHttp(
		10,
		1,
		NewUrls("http://localhost"),
		MessageKB_t{},
		&http.Client{
			Transport: DefaultTransport(time.Second, 100, 2),
			Timeout:   time.Second,
		},
		RpsLimit(NewRps(time.Second, 100, 1000)),
		PostDelay(time.Millisecond),
	)
	logger.AddOutput("http", log_http, LOG_TRACE.Levels)

	Debug("lalala %s", ByteSize(1024))
	Debug("bububu %s", ByteSize(2048))

	assert.Assert(t, buf.String() == "DEBUG lalala 1.00 KB\nDEBUG bububu 2.00 KB\n", fmt.Sprintf("%q", buf.String()))
}

func Test2(t *testing.T) {
	c := NewErrorsContext("b0dd37be-0f1e-421d-98c8-222cc57acae0", "ERROR")
	ctx := SetErrorsContext(context.Background(), c)

	logger := NewEmpty()
	SetLogger(logger)

	var buf bytes.Buffer
	logger.AddOutput("stdout", NewStdout([]Prefixer{&DT_t{}}), LOG_TRACE.Levels)
	logger.AddOutput("buf", NewStdany([]Prefixer{&DT_t{}}, &buf), LOG_TRACE.Levels)

	DebugCtx(ctx, "test")

	assert.Assert(t, buf.String() == "DEBUG b0dd37be-0f1e-421d-98c8-222cc57acae0 test\n", fmt.Sprintf("%q", buf.String()))
}

func Test3(t *testing.T) {
	rps := NewRps(100*time.Millisecond, 10, 100)
	for i := 0; i < 1000; i++ {
		rps.Add(time.Now())
		time.Sleep(1 * time.Millisecond)
	}
	for i := 0; i < 1000; i++ {
		rps.Add(time.Now())
		time.Sleep(1 * time.Millisecond)
	}
	for i := 0; i < 1000; i++ {
		rps.Add(time.Now())
		time.Sleep(1 * time.Millisecond)
	}

	time.Sleep(100 * time.Millisecond)

	s1 := rps.Size(time.Now())
	assert.Assert(t, s1 == 0, s1)
}
