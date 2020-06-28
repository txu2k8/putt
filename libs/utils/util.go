package utils

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	mathrand "math/rand"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/op/go-logging"
	"github.com/qianlnk/pgbar"
	uuid "github.com/satori/go.uuid"
	"github.com/schollz/progressbar/v3"
)

var logger = logging.MustGetLogger("test")

// SleepProgressBar ...
func SleepProgressBar(d time.Duration) {
	intSeconds := int(d.Seconds())
	bar := progressbar.New(intSeconds)
	bar.Describe(fmt.Sprintf("Sleep %s -", d))
	for i := 0; i < intSeconds; i++ {
		bar.Add(1)
		time.Sleep(1 * time.Second)
	}
	fmt.Println()
}

// PrintWithProgressBar ...
func PrintWithProgressBar(prefix string, total int) {
	var wg sync.WaitGroup
	wg.Add(1)
	bar := pgbar.NewBar(0, prefix, total)
	go func() {
		defer wg.Done()
		for i := 0; i < total; i++ {
			bar.Add()
			time.Sleep(time.Second / 300)
		}
	}()
	wg.Wait()
}

// TimeTrack ...
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	logger.Infof("function %s finishes after %s", name, elapsed)
}

// Prettify returns the string representation of a value.
func Prettify(i interface{}) string {
	var buf bytes.Buffer
	prettify(reflect.ValueOf(i), 0, &buf)
	return buf.String()
}

// prettify will recursively walk value v to build a textual
// representation of the value.
func prettify(v reflect.Value, indent int, buf *bytes.Buffer) {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		strtype := v.Type().String()
		if strtype == "time.Time" {
			fmt.Fprintf(buf, "%s", v.Interface())
			break
		} else if strings.HasPrefix(strtype, "io.") {
			buf.WriteString("<buffer>")
			break
		}

		buf.WriteString("{\n")

		names := []string{}
		for i := 0; i < v.Type().NumField(); i++ {
			name := v.Type().Field(i).Name
			f := v.Field(i)
			if name[0:1] == strings.ToLower(name[0:1]) {
				continue // ignore unexported fields
			}
			if (f.Kind() == reflect.Ptr || f.Kind() == reflect.Slice || f.Kind() == reflect.Map) && f.IsNil() {
				continue // ignore unset fields
			}
			names = append(names, name)
		}

		for i, n := range names {
			val := v.FieldByName(n)
			buf.WriteString(strings.Repeat(" ", indent+2))
			buf.WriteString(n + ": ")
			prettify(val, indent+2, buf)

			if i < len(names)-1 {
				buf.WriteString(",\n")
			}
		}

		buf.WriteString("\n" + strings.Repeat(" ", indent) + "}")
	case reflect.Slice:
		strtype := v.Type().String()
		if strtype == "[]uint8" {
			fmt.Fprintf(buf, "<binary> len %d", v.Len())
			break
		}

		nl, id, id2 := "", "", ""
		if v.Len() > 3 {
			nl, id, id2 = "\n", strings.Repeat(" ", indent), strings.Repeat(" ", indent+2)
		}
		buf.WriteString("[" + nl)
		for i := 0; i < v.Len(); i++ {
			buf.WriteString(id2)
			prettify(v.Index(i), indent+2, buf)

			if i < v.Len()-1 {
				buf.WriteString("," + nl)
			}
		}

		buf.WriteString(nl + id + "]")
	case reflect.Map:
		buf.WriteString("{\n")

		for i, k := range v.MapKeys() {
			buf.WriteString(strings.Repeat(" ", indent+2))
			buf.WriteString(k.String() + ": ")
			prettify(v.MapIndex(k), indent+2, buf)

			if i < v.Len()-1 {
				buf.WriteString(",\n")
			}
		}

		buf.WriteString("\n" + strings.Repeat(" ", indent) + "}")
	default:
		if !v.IsValid() {
			fmt.Fprint(buf, "<invalid value>")
			return
		}
		format := "%v"
		switch v.Interface().(type) {
		case string:
			format = "%q"
		case io.ReadSeeker, io.Reader:
			format = "buffer(%p)"
		}
		fmt.Fprintf(buf, format, v.Interface())
	}
}

// GetCurDir ...
func GetCurDir() string {
	dir, _ := os.Executable()
	exPath := filepath.Dir(dir)
	return exPath
}

// UniqueID returns a unique UUID-like identifier.
func UniqueID() string {
	u := make([]byte, 16)
	io.ReadFull(rand.Reader, u)
	return fmt.Sprintf("%s", u)
}

// GetUUID ...
func GetUUID() string {
	return uuid.NewV4().String()
}

// GetCurrentTimeUnix ...
func GetCurrentTimeUnix() int64 {
	return time.Now().Unix()
}

// GetRandomInt ...
func GetRandomInt(min int, max int) int {
	r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	time.Sleep(1 * time.Nanosecond)
	p := r.Perm(max - min + 1)
	return p[min]
}

// GetRandomInt64 return rand int64 i range [-m, n]
func GetRandomInt64(min, max int64) int64 {
	if min > max {
		logger.Panic("The min is greater than max!")
	}

	if min < 0 {
		f64Min := math.Abs(float64(min))
		i64Min := int64(f64Min)
		result, _ := rand.Int(rand.Reader, big.NewInt(max+1+i64Min))
		return result.Int64() - i64Min
	}
	result, _ := rand.Int(rand.Reader, big.NewInt(max-min+1))
	return min + result.Int64()
}

// GetRandString return a random string -- ERROR
func GetRandString(strSize int64) string {
	const letterBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)
	b := make([]byte, strSize)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := strSize-1, mathrand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = mathrand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// GetRandomString return a random string
func GetRandomString(strSize int64) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	time.Sleep(1 * time.Nanosecond)
	for i := int64(0); i < strSize; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

// GetRandomDigit return a random Digit
func GetRandomDigit(strSize int64) string {
	str := "0123456789"
	bytes := []byte(str)
	result := []byte{}
	r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	time.Sleep(1 * time.Nanosecond)
	for i := int64(0); i < strSize; i++ {
		if i == 0 {
			bytes9 := bytes[1:]
			result = append(result, bytes9[r.Intn(len(bytes9))])
		} else {
			result = append(result, bytes[r.Intn(len(bytes))])
		}
	}
	return string(result)
}

// GetFileMd5sum ...
func GetFileMd5sum(f *os.File) string {
	logger.Debugf("Get file MD5: %s", f.Name())
	md5 := md5.New()
	io.Copy(md5, f)
	fileMd5 := hex.EncodeToString(md5.Sum(nil))
	logger.Debug(fileMd5, f.Name())
	return fileMd5
}

// GetFileMd5sumWithPath ...
func GetFileMd5sumWithPath(filePath string) string {
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		logger.Panic(err)
	}
	defer file.Close()
	return GetFileMd5sum(file)
}

// PathExists use "os.Stat" judge if the file or dir exist
// nil: exist
// os.IsNotExist(err) == true: not exist
// other error type: not sure
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// CreateFile create original file, each line with line_number, and specified line size
// mode: w 只能写 覆盖整个文件 不存在则创建; a 只能写 从文件底部添加内容 不存在则创建
func CreateFile(filePath string, fileSize int64, lineSize int64, mode string) string {
	logger.Debugf(">> Create/Write file: %s", filePath)
	var flag int
	fileDir := path.Dir(filePath)
	err := os.MkdirAll(fileDir, os.ModePerm)
	if err != nil {
		logger.Panic(err)
	}

	lineCount := fileSize / lineSize
	unalignedSize := fileSize % lineSize

	switch mode {
	case "r": // 只能读
		flag = os.O_RDONLY
	case "r+": // 可读可写 不会创建不存在的文件 从顶部开始写 会覆盖之前此位置的内容
		flag = os.O_RDONLY | os.O_TRUNC
	case "w": // 只能写 覆盖整个文件 不存在则创建
		flag = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	case "w+": // 可读可写 如果文件存在 则覆盖整个文件不存在则创建
		flag = os.O_CREATE | os.O_RDWR | os.O_TRUNC
	case "a": // 只能写 从文件底部添加内容 不存在则创建
		flag = os.O_CREATE | os.O_WRONLY | os.O_APPEND
	case "a+": // 可读可写 从文件顶部读取内容 从文件底部添加内容 不存在则创建
		flag = os.O_CREATE | os.O_RDWR | os.O_APPEND
	default: // "a"
		flag = os.O_CREATE | os.O_WRONLY | os.O_APPEND

	}
	file, err := os.OpenFile(filePath, flag, 0666)
	if err != nil {
		logger.Panic(err)
	}
	defer file.Close()

	var lineNum int64
	for lineNum = 0; lineNum < lineCount; lineNum++ {
		randString := GetRandomString(lineSize-int64(2)-int64(len(string(lineNum)))) + "\n"
		randLineString := fmt.Sprintf("%s:%s", strconv.FormatInt(lineNum, 10), randString)
		file.WriteString(randLineString)
	}
	if unalignedSize > 0 {
		file.WriteString(GetRandomString(unalignedSize))
	}
	fileMd5 := GetFileMd5sumWithPath(filePath)
	return fileMd5
}

// CreateFileOfSize will return an *os.File that is of size bytes
func CreateFileOfSize(dir string, fileNamePrefix string, size int64) (*os.File, error) {
	file, err := ioutil.TempFile(dir, fileNamePrefix)
	if err != nil {
		return nil, err
	}

	err = file.Truncate(size)
	if err != nil {
		file.Close()
		os.Remove(file.Name())
		return nil, err
	}

	return file, nil
}

// RunCmd run command at local
func RunCmd(cmdSpc string) (rc int, output string, err error) {
	logger.Infof("Run cmd: %s", cmdSpc)
	var stdOut, stdErr bytes.Buffer
	cmdSpcSplit := strings.Split(cmdSpc, " ")
	name := cmdSpcSplit[0]
	args := []string{}
	if len(cmdSpcSplit) > 1 {
		args = cmdSpcSplit[1:]
	}
	cmd := exec.Command(name, args...)
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	err = cmd.Run()
	output = stdOut.String()
	rc = 0
	if stdErr.String() != "" {
		rc = -1
		output += stdErr.String()
	}
	if err != nil {
		logger.Infof("cmd exec failed: %s", fmt.Sprint(err))
		rc = -1
	}

	// rc, err = strconv.Atoi(strings.Replace(stdOut.String(), "\n", "", -1))
	// if err != nil {
	// 	logger.Info(err)
	// 	return -1, output, err
	// }
	return rc, output, err
}

// DeepCopy ...
func DeepCopy(src, dst interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

// DedupStringArr ...
func DedupStringArr(arr []string) (output []string) {
	tempMap := make(map[string]bool)
	for _, value := range arr {
		if _, ok := tempMap[value]; !ok {
			tempMap[value] = true
			output = append(output, value)
		}
	}
	return
}

// UniqArr Remove Repeated Element in Array
func UniqArr(arr []string) (newArr []string) {
	newArr = make([]string, 0)
	for i := 0; i < len(arr); i++ {
		repeat := false
		for j := i + 1; j < len(arr); j++ {
			if arr[i] == arr[j] {
				repeat = true
				break
			}
		}
		if !repeat {
			newArr = append(newArr, arr[i])
		}
	}
	return
}

// GetLocalIP .
func GetLocalIP() (ip string) {
	conn, err := net.Dial("udp", "google.com:80")
	if err != nil {
		logger.Error(err.Error())
		return
	}
	defer conn.Close()
	ip = strings.Split(conn.LocalAddr().String(), ":")[0]
	return
}

// IsPingOK .
func IsPingOK(ip string) error {
	var cmdSpc string
	sysstr := "Windows"
	switch sysstr {
	case "Windows":
		cmdSpc = fmt.Sprintf("C:\\Windows\\System32\\ping %s", ip)
	case "Linux":
		cmdSpc = fmt.Sprintf("ping -c1 %s", ip)
	default:
		cmdSpc = fmt.Sprintf("ping %s", ip)
	}
	rc, out, err := RunCmd(cmdSpc)
	logger.Info(out)
	if err != nil {
		return err
	}
	if rc != 0 {
		return fmt.Errorf(out)
	}
	return nil
}

// Prompt .
func Prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}

// MinInt .
func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// MaxInt .
func MaxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}
