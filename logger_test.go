package log

import "testing"

func ExampleLog1() {
	SetLogger(NewLogger("stdout", 0, NewStdout("")))
	Debug("lalala")
	Debug("bububu")
/* Output:
DEBUG lalala
DEBUG bububu
*/
}

func TestLog1(t * testing.T) {

}
