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
	m := NewLevelMap()

	var buf bytes.Buffer
	ts := time.Now()
	m.AddOutputs("stdout", NewWriterStdany([]Formatter{NewPartDateTime(""), NewPartLevelName("", ""), NewPartTextMessage(), NewPartNewLine()}, os.Stdout, 0), WhatLevel(0))
	m.AddOutputs("buf", NewWriterStdany([]Formatter{NewPartDateTime(""), NewPartLevelName("", ""), NewPartTextMessage(), NewPartNewLine()}, &buf, 0), WhatLevel(0))
	log_file, _ := NewWriterFileBytes(ts, "/tmp/test.log", []Formatter{NewPartDateTime(""), NewPartLevelName("", ""), NewPartTextMessage(), NewPartNewLine()}, 1024, 10, 0)
	m.AddOutputs("file", log_file, WhatLevel(0))
	log_http := NewWriterHttp(
		NewUrls("http://localhost"),
		MessageKB_t{},
		&http.Client{
			Transport: DefaultTransport(time.Second, 100, 2),
			Timeout:   time.Second,
		},
		RpsLimit(NewRps(time.Second, 100, 1000)),
		PostDelay(time.Millisecond),
	)
	m.AddOutputs("http", log_http, WhatLevel(0))

	SetLogger(New(m))

	Debug("lalala %s", ByteSize(1024))
	Debug("bububu %s", ByteSize(2048))

	assert.Assert(t, buf.String() == "DEBUG lalala 1.00 KB\nDEBUG bububu 2.00 KB\n", fmt.Sprintf("%q", buf.String()))
}

func Test2(t *testing.T) {
	c := NewLogBuffer("b0dd37be-0f1e-421d-98c8-222cc57acae0", 10)
	ctx := SetLogBuffer(context.Background(), c)

	m := NewLevelMap()

	var buf bytes.Buffer
	m.AddOutputs("stdout", NewWriterStdany([]Formatter{NewPartDateTime(""), NewPartBufferId(), NewPartLevelName("_", "_"), NewPartTextMessage(), NewPartNewLine()}, os.Stdout, 0), WhatLevel(0))
	m.AddOutputs("buf", NewWriterStdany([]Formatter{NewPartDateTime(""), NewPartBufferId(), NewPartLevelName("_", "_"), NewPartTextMessage(), NewPartNewLine()}, &buf, 0), WhatLevel(0))

	SetLogger(New(m))

	DebugCtx(ctx, "test")

	assert.Assert(t, buf.String() == "b0dd37be-0f1e-421d-98c8-222cc57acae0 _DEBUG_ test\n", fmt.Sprintf("%q", buf.String()))
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
