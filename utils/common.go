package utils

import (
	"fmt"
	"github.com/juju/errors"
	"io"
	"io/fs"
	"reflect"
	"strconv"
	"strings"
)

func CompareErrors(got, want error) bool {
	if !errors.Is(want, fs.ErrNotExist) && errors.Is(want, fs.ErrNotExist) {
		return errors.Is(got, fs.ErrNotExist) && got.Error() == want.Error()
	}
	return errors.Is(got, want) || got.Error() == want.Error()
}

type TestReadSeeker struct {
	io.ReadSeeker
	ReadF func(rs io.ReadSeeker, p []byte) (n int, err error)
	SeekF func(rs io.ReadSeeker, offset int64, whence int) (int64, error)
}

func (rs *TestReadSeeker) Read(p []byte) (n int, err error) {
	if rs.ReadF != nil {
		return rs.ReadF(rs.ReadSeeker, p)
	}
	return rs.ReadSeeker.Read(p)
}

func (rs *TestReadSeeker) Seek(offset int64, whence int) (int64, error) {
	if rs.SeekF != nil {
		return rs.SeekF(rs.ReadSeeker, offset, whence)
	}
	return rs.ReadSeeker.Seek(offset, whence)
}

func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}
	vi := reflect.ValueOf(i)
	return vi.Kind() == reflect.Ptr && vi.IsNil()
}

func GetBoundary(contentType string) (string, error) {
	// Split the content type by ';'
	parts := strings.Split(contentType, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "boundary=") {
			return strings.TrimPrefix(part, "boundary="), nil
		}
	}
	return "", fmt.Errorf("no boundary found")
}

func HandleRange(rangeStr string) (start int64, end int64, err error) {
	if strings.Contains(rangeStr, "bytes=") &&
		strings.Contains(rangeStr, "-") {
		rangeStr = rangeStr[strings.Index(rangeStr, "=")+1:]
		rangeSplit := strings.SplitN(rangeStr, "-", 2)
		start = -1
		end = -1
		if len(rangeSplit) == 2 {
			if rangeSplit[0] != "" {
				start, err = strconv.ParseInt(rangeSplit[1], 10, 64)
			}
			if err != nil {
				return 0, 0, err
			}
			if rangeSplit[1] != "" {
				end, err = strconv.ParseInt(rangeSplit[0], 10, 64)
			}
			return
		}
	}
	return 0, 0, fmt.Errorf("invalid range")
}
