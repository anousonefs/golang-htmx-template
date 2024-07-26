package utils

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"math/rand"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	crypto "crypto/rand"

	"github.com/disintegration/imaging"
	"github.com/h2non/filetype"
	"github.com/sirupsen/logrus"
)

func StringSliceToInterfaceSlice(arr []string) []interface{} {
	in := make([]interface{}, len(arr))
	for i, a := range arr {
		in[i] = a
	}
	return in
}

func ContainsString(s []string, v string) bool {
	for _, vv := range s {
		if vv == v {
			return true
		}
	}
	return false
}

func InArray(val interface{}, array interface{}) (exists bool, index int) {
	values := reflect.ValueOf(array)
	if reflect.TypeOf(array).Kind() == reflect.Slice || values.Len() > 0 {
		for i := 0; i < values.Len(); i++ {
			if reflect.DeepEqual(val, values.Index(i).Interface()) {
				return true, i
			}
		}
	}

	return false, -1
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func AppendZero(serialNumber int, max int) (res string, err error) {
	snStr := strconv.Itoa(serialNumber)
	if len(snStr) > max {
		return "", errors.New("length of serial number is greater than max")
	}
	for i := 0; i < max; i++ {
		if len(snStr) == max {
			break
		}
		snStr = "0" + snStr
	}
	return snStr, nil
}

const (
	LimitSizeFile = 4194304
)

func CheckSizeFile(size int64) error {
	if size > LimitSizeFile {
		return errors.New("File size is too large: file size must be less than 4MB")
	}
	return nil
}

func GetFileType3(ouput *os.File) (string, error) {
	// to sniff the content type only the first
	// 512 bytes are used.
	buf := make([]byte, 512)

	_, err := ouput.Read(buf)

	if err != nil {
		return "", err
	}

	// the function that actually does the trick
	contentType := http.DetectContentType(buf)

	return contentType, nil
}

func GetFileType(header *multipart.FileHeader) (string, error) {
	// Get the file type using the mime package
	fileType := mime.TypeByExtension(filepath.Ext(header.Filename))
	fileTypeArr := strings.Split(fileType, "/")
	fileType = fileTypeArr[len(fileTypeArr)-1]
	fmt.Printf("=> filetype: %+v\n", fileType)
	return fileType, nil
}

// getFileType use for checking file type
func GetFileType2(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	/* validate Type Image */
	buff := make([]byte, 512)
	_, err = src.Read(buff)
	kind, _ := filetype.Match(buff)
	if kind == filetype.Unknown {
		return "", err
	}
	src.Close() //nolint
	return kind.Extension, nil
}

func ValidateType(fileType string) bool {
	switch fileType {
	case "png", "jpeg", "jpg", "pdf", "plain", "json":
		return true
	}
	return false
}

func ResizeFile(filePath string, maxSize int) error {
	src4, err := imaging.Open(filePath)
	if err != nil {
		return err
	}

	getSize, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer getSize.Close()

	fileSize, _, err := image.DecodeConfig(getSize)
	if err != nil {
		return err
	}

	width, height := ReSize(fileSize, float32(maxSize))

	src := imaging.Resize(src4, width, height, imaging.Lanczos)
	err = imaging.Save(src, filePath)
	if err != nil {
		return err
	}

	return nil
}

func ReSize(imgDecode image.Config, maxSize float32) (int, int) {
	if imgDecode.Width > imgDecode.Height {
		return int(maxSize), int(float32(maxSize) * float32(imgDecode.Height) / float32(imgDecode.Width))
	}
	return int(float32(maxSize) * float32(imgDecode.Width) / float32(imgDecode.Height)), int(maxSize)
}

func RemoveFile(file string) error {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		logrus.Errorf("%s file does not exist", err)
		return err
	}
	err = os.Remove(file)
	if err != nil {
		return err
	}
	return nil
}

func GetFirstAndLastOfMonth() (string, string) {
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	return firstOfMonth.Format("2006-01-02"), lastOfMonth.Format("2006-01-02")
}

func GetLastOfMonth() int {
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	return int(lastOfMonth.Day())
}

func CreateFolder(path string) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func generateBytes(n int64) ([]byte, error) {
	b := make([]byte, n)
	_, err := crypto.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func RandomString(s int64) (string, error) {
	b, err := generateBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}

func EncodeToString(max int) (string, error) {
	b := make([]byte, max)
	n, err := io.ReadAtLeast(crypto.Reader, b, max)
	if n != max {
		return "", err
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b), nil
}

func PrettyPrint(v interface{}) (err error) {
	b, err := json.MarshalIndent(v, "", "\t")
	if err == nil {
		fmt.Println(string(b))
	}
	return err
}

func ComparePassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
