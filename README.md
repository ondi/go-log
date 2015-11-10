```
import "github.com/ondi/go-log"

logger := log.NewLogger(LogLevel)
if len(LogFile) > 0 {
	log_rotate := log.NewRotateLogWriter(LogFile, LogSize, LogBackup, DuplicateOnStderr)
	logger.SetOutput(log_rotate)
}
log.SetLogger(logger)

log.Info("%v", "test")
```

