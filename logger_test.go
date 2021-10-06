package log

import (
	"bytes"
	"context"
	"testing"
	"time"

	"gotest.tools/assert"
)

func Test1(t *testing.T) {
	logger := NewEmpty()
	SetLogger(logger)

	var buf bytes.Buffer
	logger.AddOutput("stdout", LOG_TRACE, NewStdout(&DT_t{}))
	logger.AddOutput("buf", LOG_TRACE, NewStdany(&DT_t{}, &buf))
	log_file, _ := NewFileBytes("/tmp/test.log", &DT_t{}, 1024, 10)
	logger.AddOutput("file", LOG_TRACE, log_file)
	log_http := NewHttp(
		10,
		1,
		NewUrls("http://localhost"),
		MessageKB_t{},
		DefaultClient(DefaultTransport(time.Second, 100, 2), time.Second),
		RpsLimit(NewRps(time.Second, 100, 1000)),
		PostDelay(time.Millisecond),
	)
	logger.AddOutput("http", LOG_TRACE, log_http)

	Debug("lalala %s", ByteSize(1024))
	Debug("bububu %s", ByteSize(2048))

	assert.Assert(t, buf.String() == "DEBUG lalala 1.00 KB\nDEBUG bububu 2.00 KB\n", buf.String())
}

func Test2(t *testing.T) {
	c := NewContext("b0dd37be-0f1e-421d-98c8-222cc57acae0")
	ctx := CtxSet(context.Background(), c)

	logger := NewEmpty()
	SetLogger(logger)

	var buf bytes.Buffer
	logger.AddOutput("stdout", LOG_TRACE, NewStdout(&DT_t{}))
	logger.AddOutput("buf", LOG_TRACE, NewStdany(&DT_t{}, &buf))

	DebugCtx(ctx, "test")

	assert.Assert(t, buf.String() == "DEBUG b0dd37be-0f1e-421d-98c8-222cc57acae0 test\n", buf.String())
}
