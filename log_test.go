package log

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"gotest.tools/assert"
)

func Test1(t *testing.T) {
	logger := SetLogger(New(LEVELS))

	var buf bytes.Buffer
	ts := time.Now()
	logger.AddOutput("stdout", NewWriterStdany([]Formatter{NewDt("")}, os.Stdout, 0), WhatLevel(LOG_TRACE.Level))
	logger.AddOutput("buf", NewWriterStdany([]Formatter{NewDt("")}, &buf, 0), WhatLevel(LOG_TRACE.Level))
	log_file, _ := NewWriterFileBytes(ts, "/tmp/test.log", []Formatter{NewDt("")}, 1024, 10, 0)
	logger.AddOutput("file", log_file, WhatLevel(LOG_TRACE.Level))
	log_http := NewHttpQueue(
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
	logger.AddOutput("http", log_http, WhatLevel(LOG_TRACE.Level))

	Debug("lalala %s", ByteSize(1024))
	Debug("bububu %s", ByteSize(2048))

	assert.Assert(t, buf.String() == "DEBUG lalala 1.00 KB\nDEBUG bububu 2.00 KB\n", fmt.Sprintf("%q", buf.String()))
}

func Test2(t *testing.T) {
	c := NewLogContext("b0dd37be-0f1e-421d-98c8-222cc57acae0", []Level_t{LOG_ERROR})
	ctx := SetLogContextValue(context.Background(), c)

	logger := SetLogger(New(LEVELS))

	var buf bytes.Buffer
	logger.AddOutput("stdout", NewWriterStdany([]Formatter{NewDt(""), NewSetLogContext()}, os.Stdout, 0), WhatLevel(LOG_TRACE.Level))
	logger.AddOutput("buf", NewWriterStdany([]Formatter{NewDt(""), NewSetLogContext()}, &buf, 0), WhatLevel(LOG_TRACE.Level))

	DebugCtx(ctx, "test")

	assert.Assert(t, buf.String() == "b0dd37be-0f1e-421d-98c8-222cc57acae0 DEBUG test\n", fmt.Sprintf("%q", buf.String()))
}

func Test3(t *testing.T) {
	rps := NewRps(100*time.Millisecond, 10, 100)

	ts := time.Now()
	for i := 0; i < 1000; i++ {
		rps.Add(time.Now())
		time.Sleep(1 * time.Millisecond)
	}
	t.Logf("CYCLE1: %v", time.Since(ts))

	ts = time.Now()
	for i := 0; i < 1000; i++ {
		rps.Add(time.Now())
		time.Sleep(1 * time.Millisecond)
	}
	t.Logf("CYCLE2: %v", time.Since(ts))

	ts = time.Now()
	for i := 0; i < 1000; i++ {
		rps.Add(time.Now())
		time.Sleep(1 * time.Millisecond)
	}
	t.Logf("CYCLE3: %v", time.Since(ts))

	time.Sleep(100 * time.Millisecond)

	s1 := rps.Size(time.Now())
	assert.Assert(t, s1 == 0, s1)
}
