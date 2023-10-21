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
	logger := New()
	SetLogger(logger)

	var buf bytes.Buffer
	ts := time.Now()
	logger.AddOutput("stdout", NewStdout([]Formatter{NewDt("")}), WhatLevel(LOG_TRACE.Level))
	logger.AddOutput("buf", NewStdany([]Formatter{NewDt("")}, &buf), WhatLevel(LOG_TRACE.Level))
	log_file, _ := NewFileBytes(ts, "/tmp/test.log", []Formatter{NewDt("")}, 1024, 10)
	logger.AddOutput("file", log_file, WhatLevel(LOG_TRACE.Level))
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
	logger.AddOutput("http", log_http, WhatLevel(LOG_TRACE.Level))

	Debug("lalala %s", ByteSize(1024))
	Debug("bububu %s", ByteSize(2048))

	assert.Assert(t, buf.String() == "DEBUG lalala 1.00 KB\nDEBUG bububu 2.00 KB\n", fmt.Sprintf("%q", buf.String()))
}

func Test2(t *testing.T) {
	c := NewErrorsContext("b0dd37be-0f1e-421d-98c8-222cc57acae0", "ERROR")
	ctx := SetErrorsContext(context.Background(), c)

	logger := New()
	SetLogger(logger)

	var buf bytes.Buffer
	logger.AddOutput("stdout", NewStdout([]Formatter{NewDt(""), NewCx()}), WhatLevel(LOG_TRACE.Level))
	logger.AddOutput("buf", NewStdany([]Formatter{NewDt(""), NewCx()}, &buf), WhatLevel(LOG_TRACE.Level))

	DebugCtx(ctx, "test")

	assert.Assert(t, buf.String() == "b0dd37be-0f1e-421d-98c8-222cc57acae0 DEBUG test\n", fmt.Sprintf("%q", buf.String()))
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
