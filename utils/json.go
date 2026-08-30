package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

const defaultJSONBodyLimit = 2 * 1024 * 1024 // 2Mb max json body size limit

type DecodeStrictJSONParams struct {
	Context        echo.Context
	Dest           any
	MaxBytes       int64
	AllowEmptyBody bool
}

func DecodeStrictJSON(params DecodeStrictJSONParams) error {
	maxBytes := params.MaxBytes
	if maxBytes == 0 {
		maxBytes = defaultJSONBodyLimit
	}

	req := params.Context.Request()
	req.Body = http.MaxBytesReader(params.Context.Response().Writer, req.Body, maxBytes)

	var raw json.RawMessage
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&raw); err != nil {
		if params.AllowEmptyBody && errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("request body must contain a single JSON object")
		}
		return err
	}

	tokenDecoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := tokenDecoder.Token()
	if err != nil {
		return err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '{' {
		return errors.New("request body must contain a JSON object")
	}

	strictDecoder := json.NewDecoder(bytes.NewReader(raw))
	strictDecoder.DisallowUnknownFields()
	return strictDecoder.Decode(params.Dest)
}
