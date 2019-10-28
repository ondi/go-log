```
package main

import "github.com/ondi/go-log"

func main() {
        ExampleLogLevel := 0
        ExampleLogFile := "app.log"
        ExampleLogSize := 10 * 1024 * 1024
        ExampleLogBackup := 5
        logger := log.NewLogger("stderr", ExampleLogLevel, log.NewStderr(log.DATETIME1))
        if len(ExampleLogFile) > 0 {
                if log_file, err := log.NewFile(ExampleLogFile, log.DATETIME1, ExampleLogSize, ExampleLogBackup); err == nil {
                        logger.AddOutput("logfile", ExampleLogLevel, log_file)
                }
        }
        log.SetLogger(logger)

        log.Info("%v", "test")
}
```
