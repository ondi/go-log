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
	m := NewLogMap()

	var buf bytes.Buffer
	ts := time.Now()
	m.AddOutputs("stdout", NewWriterStdany([]Formatter{NewDt("")}, os.Stdout, 0), WhatLevel(LOG_TRACE.LevelId))
	m.AddOutputs("buf", NewWriterStdany([]Formatter{NewDt("")}, &buf, 0), WhatLevel(LOG_TRACE.LevelId))
	log_file, _ := NewWriterFileBytes(ts, "/tmp/test.log", []Formatter{NewDt("")}, 1024, 10, 0)
	m.AddOutputs("file", log_file, WhatLevel(LOG_TRACE.LevelId))
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
	m.AddOutputs("http", log_http, WhatLevel(LOG_TRACE.LevelId))

	SetLogger(New(&m))

	Debug("lalala %s", ByteSize(1024))
	Debug("bububu %s", ByteSize(2048))

	assert.Assert(t, buf.String() == "DEBUG lalala 1.00 KB\nDEBUG bububu 2.00 KB\n", fmt.Sprintf("%q", buf.String()))
}

func Test2(t *testing.T) {
	c := NewLogContext("b0dd37be-0f1e-421d-98c8-222cc57acae0", 10)
	ctx := SetLogContext(context.Background(), c)

	m := NewLogMap()

	var buf bytes.Buffer
	m.AddOutputs("stdout", NewWriterStdany([]Formatter{NewDt(""), NewGetLogContext()}, os.Stdout, 0), WhatLevel(LOG_TRACE.LevelId))
	m.AddOutputs("buf", NewWriterStdany([]Formatter{NewDt(""), NewGetLogContext()}, &buf, 0), WhatLevel(LOG_TRACE.LevelId))

	SetLogger(New(&m))

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
