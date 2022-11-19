package shared

import "strings"

func CtrlCode(s string) (code string) {
	return map[string]string{
		"disconnect":  "\000DISCONNECT\000",
		"username":    "\000USERNAME\000",
		"acknowledge": "\000ACK\000",
	}[s] //error if property called wrong in client
}

//removes newlines and carriage returns
func Sanitize(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\n", ""), "\r", "")
}
