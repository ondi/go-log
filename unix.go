//
// +build linux
//

package log

import (
	"os"
	"syscall"
)

func DupStderr(filename string) (fp *os.File, err error) {
	if fp, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE /*| os.O_APPEND*/, 0644); err != nil {
		return
	}
	if err = syscall.Dup2(int(fp.Fd()), syscall.Stderr); err != nil {
		fp.Close()
		return
	}
	return
}

func DupStdout(filename string) (fp *os.File, err error) {
	if fp, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE /*| os.O_APPEND*/, 0644); err != nil {
		return
	}
	if err = syscall.Dup2(int(fp.Fd()), syscall.Stdout); err != nil {
		fp.Close()
		return
	}
	return
}
