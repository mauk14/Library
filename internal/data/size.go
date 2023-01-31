package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidSizeFormat = errors.New("invalid size format")

type Size int32

func (r Size) MarshalJSON() ([]byte, error) {

	jsonValue := fmt.Sprintf("%d pages", r)

	quotedJSONValue := strconv.Quote(jsonValue)

	return []byte(quotedJSONValue), nil
}

func (r *Size) UnmarshalJSON(jsonValue []byte) error {

	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidSizeFormat
	}

	parts := strings.Split(unquotedJSONValue, " ")

	if len(parts) != 2 || parts[1] != "pages" {
		return ErrInvalidSizeFormat
	}

	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidSizeFormat
	}

	*r = Size(i)
	return nil
}
