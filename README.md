```
import "os"
import "github.com/ondi/go-log"

logger := log.NewLogger(os.Stderr, LogLevel, log.DATETIME1)
if len(LogFile) > 0 {
	log_rotate := log.NewRotateLogWriter(LogFile, LogSize, LogBackup)
	logger.AddOutput(log_rotate, LogLevel, log.DATETIME1)
}
log.SetLogger(logger)

log.Info("%v", "test")
```
