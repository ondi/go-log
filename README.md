```
package main

import "github.com/ondi/go-log"

func main() {
	ExampleLogLevel := 0
	ExampleLogFile := "app.log"
	ExampleLogSize := 10 * 1024 * 1024
	ExampleLogBackup := 5

	logger := log.NewLogger("stderr", ExampleLogLevel, log.NewStderr(log.DT))
	log.SetLogger(logger)

	if len(ExampleLogFile) > 0 {
			if log_file, err := log.NewFileBytes(ExampleLogFile, log.DT, ExampleLogSize, ExampleLogBackup); err == nil {
					logger.AddOutput("logfile", ExampleLogLevel, log_file)
			}
	}

	log.Info("%v", "test")
}
```
