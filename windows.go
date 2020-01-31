//
// +build windows
//

package log

import "os"
import "fmt"

func DupStderr(filename string) (* os.File, error) {
	return nil, fmt.Errorf("NOT IMPLEMENTED")
}

func DupStdout(filename string) (* os.File, error) {
	return nil, fmt.Errorf("NOT IMPLEMENTED")
}
