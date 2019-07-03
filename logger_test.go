package log

import "os"
import "testing"

func ExampleLog1() {
	logger := NewLogger("stderr", 0, os.Stdout, "")
	SetLogger(logger)
	Debug("lalala")
	Debug("bububu")
/* Output:
DEBUG lalala
DEBUG bububu
*/
}

func Test1(t * testing.T) {

}
