package log

import (
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"

	"gotest.tools/assert"
)

func Test1(t *testing.T) {
	logger := NewEmpty()
	SetLogger(logger)

	var buf bytes.Buffer
	logger.AddOutput("stdout", NewStdout(&DT_t{}), LOG_TRACE.Levels)
	logger.AddOutput("buf", NewStdany(&DT_t{}, &buf), LOG_TRACE.Levels)
	log_file, _ := NewFileBytes("/tmp/test.log", &DT_t{}, 1024, 10)
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

	assert.Assert(t, buf.String() == "DEBUG lalala 1.00 KB\nDEBUG bububu 2.00 KB\n", buf.String())
}

func Test2(t *testing.T) {
	c := ContextNew("b0dd37be-0f1e-421d-98c8-222cc57acae0", "ERROR")
	ctx := ContextSet(context.Background(), c)

	logger := NewEmpty()
	SetLogger(logger)

	var buf bytes.Buffer
	logger.AddOutput("stdout", NewStdout(&DT_t{}), LOG_TRACE.Levels)
	logger.AddOutput("buf", NewStdany(&DT_t{}, &buf), LOG_TRACE.Levels)

	DebugCtx(ctx, "test")

	assert.Assert(t, buf.String() == "DEBUG b0dd37be-0f1e-421d-98c8-222cc57acae0 test\n", buf.String())
}
