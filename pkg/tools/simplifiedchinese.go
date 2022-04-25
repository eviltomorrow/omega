package tools

import (
	"golang.org/x/text/encoding/simplifiedchinese"
)

type Charset string

const (
	UTF8     Charset = "UTF-8"
	GB18030  Charset = "GB18030"
	GBK      Charset = "GBK"
	HZGB2312 Charset = "HZGB2312"
)

func BytesToString(charset Charset, buf []byte) string {
	var str string
	switch charset {
	case GB18030:
		tmp, _ := simplifiedchinese.GB18030.NewDecoder().Bytes(buf)
		str = string(tmp)
	case GBK:
		tmp, _ := simplifiedchinese.GBK.NewDecoder().Bytes(buf)
		str = string(tmp)
	case HZGB2312:
		tmp, _ := simplifiedchinese.HZGB2312.NewDecoder().Bytes(buf)
		str = string(tmp)
	case UTF8:
		fallthrough
	default:
		str = string(buf)
	}
	return str
}

func BytesToStringSlow(buf []byte) string {
	var (
		charset = GetStrCoding(buf)
		str     string
	)
	switch charset {
	case GBK:
		tmp, _ := simplifiedchinese.GBK.NewDecoder().Bytes(buf)
		str = string(tmp)
	case UTF8:
		fallthrough
	default:
		str = string(buf)
	}
	return str
}

func GetStrCoding(data []byte) Charset {
	if IsUTF8(data) {
		return UTF8
	}
	if IsGBK(data) {
		return GBK
	}
	return UTF8
}

func IsGBK(data []byte) bool {
	length := len(data)
	var i int = 0
	for i < length {
		if data[i] <= 0x7f {
			i++
			continue
		} else {
			if data[i] >= 0x81 &&
				data[i] <= 0xfe &&
				data[i+1] >= 0x40 &&
				data[i+1] <= 0xfe &&
				data[i+1] != 0xf7 {
				i += 2
				continue
			} else {
				return false
			}
		}
	}
	return true
}

func preNum(data byte) int {
	var mask byte = 0x80
	var num int = 0
	for i := 0; i < 8; i++ {
		if (data & mask) == mask {
			num++
			mask = mask >> 1
		} else {
			break
		}
	}
	return num
}

func IsUTF8(data []byte) bool {
	i := 0
	for i < len(data) {
		if (data[i] & 0x80) == 0x00 {
			i++
			continue
		} else if num := preNum(data[i]); num > 2 {
			// i++
			for j := 0; j < num-1 && i < len(data); j++ {
				if (data[i] & 0xc0) != 0x80 {
					return false
				}
				i++
			}
		} else {
			return false
		}
	}
	return true
}
