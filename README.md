```
import "os"
import "github.com/ondi/go-log"

logger := log.NewLogger("stderr", LogLevel, os.Stderr, log.DATETIME1)
if len(LogFile) > 0 {
	log_rotate := log.NewRotateLogWriter(LogFile, LogSize, LogBackup)
	logger.AddOutput("logfile", LogLevel, log_rotate, log.DATETIME1)
}
log.SetLogger(logger)

log.Info("%v", "test")
```
