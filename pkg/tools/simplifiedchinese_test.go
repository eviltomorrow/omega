package tools

import (
	"testing"

	"golang.org/x/text/encoding/simplifiedchinese"
)

func TestIsUTF8(t *testing.T) {
	var s1 = "你好, Hello"
	flag := IsUTF8([]byte(s1))
	t.Log(flag)
}

func TestIsGBK(t *testing.T) {
	var s1 = "你好, Hello"
	tmp, _ := simplifiedchinese.GBK.NewEncoder().Bytes([]byte(s1))
	flag := IsGBK([]byte(tmp))
	t.Log(flag)

	flag = IsUTF8([]byte(tmp))
	t.Log(flag)
}
