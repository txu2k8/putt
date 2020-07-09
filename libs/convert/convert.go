package convert

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unsafe"

	"github.com/op/go-logging"
)

// Some converters such as ip->int / int->ip

var logger = logging.MustGetLogger("test")

// IP2Int convert ipaddress to int number
func IP2Int(ip string) int {
	ipArr := []int{}
	for _, x := range strings.Split(ip, ".") {
		intX, _ := strconv.Atoi(x)
		ipArr = append(ipArr, intX)
	}
	return ipArr[0]<<24 | ipArr[1]<<16 | ipArr[2]<<8 | ipArr[3]
}

// Int2IP convert int number to ipaddress
func Int2IP(n int) string {
	ipArr := make([]string, 4)
	ipArr[3] = strconv.Itoa((n & 0xff))
	ipArr[2] = strconv.Itoa((n & 0xff00) >> 8)
	ipArr[1] = strconv.Itoa((n & 0xff0000) >> 16)
	ipArr[0] = strconv.Itoa((n & 0xff000000) >> 24)
	return strings.Join(ipArr, ".")
}

// StrNumToIntArr .
func StrNumToIntArr(strN, sep string, lenArr int) []int {
	intArr := []int{}
	patten := fmt.Sprintf("[^0-9%s|\\-1\\-\\-9]", sep)
	reg := regexp.MustCompile(patten)
	matched := reg.FindStringSubmatch(strN)
	if matched != nil {
		panic(fmt.Sprintf("non-integer in strN:%s", strN))
	}
	strSplit := strings.Split(strN, sep)
	for _, s := range strSplit {
		si, _ := strconv.Atoi(s)
		intArr = append(intArr, si)
	}
	n := len(intArr)
	if lenArr > n {
		end := intArr[n-1]
		for i := n; i < lenArr; i++ {
			intArr = append(intArr, end)
		}
	}
	return intArr
}

// Base64Encode ...
func Base64Encode(input []byte) string {
	encodeString := base64.StdEncoding.EncodeToString(input)
	return encodeString
}

// Base64Decode ...
func Base64Decode(encodeString string) []byte {
	decodeBytes, err := base64.StdEncoding.DecodeString(encodeString)
	if err != nil {
		panic(err)
	}
	return decodeBytes
}

// Byte2String returns a human-readable string for the given size bytes
// precision: decimal, 12 * 1024 * 1024 * 1000 --> 11.7GB
func Byte2String(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

// String2Byte Parse a string size to int64 byte
func String2Byte(s string) int64 {
	const unit = 1024
	const u = "kMGTPE"
	div := float64(unit)

	reg := regexp.MustCompile(`(^[+-]?(0|([1-9]\d*))(\.\d+)?)\s?(\S+)`)
	matched := reg.FindStringSubmatch(s)
	exp := strings.Index(u, strings.ToUpper(matched[len(matched)-1][:1]))
	for x := 0; x < exp; x++ {
		div *= unit
	}
	n, _ := strconv.ParseFloat(matched[1], 64)
	b := int64(n * div)
	return b
}

// BytesToStringFast .
func BytesToStringFast(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// StringToBytes .
func StringToBytes(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{Data: sh.Data, Len: sh.Len, Cap: 0}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

// EscapeString convert string with "\\" to "\\\\"
func EscapeString(s string) string {
	var signEscapeMap = map[string]string{
		"\\": "\\\\",
	}

	for source := range signEscapeMap {
		s = strings.Replace(s, source, signEscapeMap[source], -1)
	}

	return s
}

// To12Hour Convert a 24-hour clock value to a 12-hour one.
func To12Hour(h24 int) (int, string) {
	switch {
	case h24 == 0:
		return 12, "am"
	case h24 < 12:
		return h24, "am"
	case h24 == 12:
		return 12, "pm"
	default:
		return h24 - 12, "pm"
	}
}

// To24Hour Convert a 13-hour clock value to a 24-hour one.
func To24Hour(h12 int, ampm string) int {
	switch ampm {
	case "am":
		if h12 == 12 {
			return 0
		}
		return h12
	case "pm":
		if h12 == 12 {
			return 12
		}
		return h12 + 12
	default:
		panic("Please input ampm as am or pm")
	}
}

// ReverseArr ...
func ReverseArr(arr []interface{}) (reverseArr []interface{}) {
	length := len(arr)
	for i := 0; i < length; i++ {
		reverseArr = append(reverseArr, arr[length-1-i])
	}
	return
}

// ReverseStringArr ...
func ReverseStringArr(arr []string) []string {
	sArr := make([]interface{}, len(arr))
	for i, v := range arr {
		sArr[i] = v
	}

	rArr := ReverseArr(sArr)
	reverseArr := make([]string, len(sArr))
	for i, v := range rArr {
		reverseArr[i] = v.(string)
	}
	return reverseArr
}

// StrFirstToUpper upper the first char in string
func StrFirstToUpper(str string) string {
	if len(str) < 1 {
		return ""
	}
	strArry := []rune(str)
	if strArry[0] >= 97 && strArry[0] <= 122 {
		strArry[0] -= 32
	}
	return string(strArry)
}
