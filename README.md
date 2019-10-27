```
package main

import "github.com/ondi/go-log"

func main() {
	ExampleLogLevel := 0
	ExampleLogFile := "app.log"
	ExampleLogSize := 10 * 1024 * 1024
	ExampleLogBackup := 5
	logger := log.NewLogger("stderr", ExampleLogLevel, log.NewLogStderr(log.DATETIME1))
	if len(ExampleLogFile) > 0 {
		log_rotate := log.NewLogFile(ExampleLogFile, log.DATETIME1, ExampleLogSize, ExampleLogBackup)
		logger.AddOutput("logfile", ExampleLogLevel, log_rotate)
	}
	log.SetLogger(logger)

	log.Info("%v", "test")
}
```
