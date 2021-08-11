```
package main

import "github.com/ondi/go-log"

func main() {
	LogLevel := 0
	LogFile := "app.log"
	LogDate := "2006-01-02 15:04:05"
	LogSize := 10 * 1024 * 1024
	LogBackup := 5

	logger := log.NewLogger("stderr", log.WhatLevel(LogLevel), log.NewStderr(&log.DTFL_t{Format: LogDate, Depth: 5}))
	log.SetLogger(logger)

	if len(LogFile) > 0 {
		if log_file, err := log.NewFileBytes(LogFile, &log.DTFL_t{Format: LogDate, Depth: 5}, LogSize, LogBackup); err == nil {
			logger.AddOutput("logfile", log.WhatLevel(LogLevel), log_file)
		}
	}

	log.Info("%v", "test")
}
```
