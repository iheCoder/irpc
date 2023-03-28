package common

import "testing"

func TestFindStringEndInByteArray(t *testing.T) {
	s := "你好啊,hello.;;'',/n /t "
	sb := []byte(s)
	b := make([]byte, len(sb)+2)
	b[0] = StringSep
	copy(b[1:len(sb)-1], sb)

	b[len(b)-1] = StringSep
	x := FindEndForSepInByteArray(b, 0, StringSep)
	if x != len(b)-1 {
		t.Fatalf("wrong")
	}
}
