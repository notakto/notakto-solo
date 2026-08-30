package utils

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

const defaultJSONBodyLimit = 2 * 1024 * 1024 // 2Mb max json body size limit

type DecodeStrictJSONParams struct {
	Context  echo.Context
	Dest     any
	MaxBytes int64
}

func DecodeStrictJSON(params DecodeStrictJSONParams) error {
	maxBytes := params.MaxBytes
	if maxBytes == 0 {
		maxBytes = defaultJSONBodyLimit
	}

	req := params.Context.Request()
	req.Body = http.MaxBytesReader(params.Context.Response().Writer, req.Body, maxBytes)

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(params.Dest); err != nil {
		return err
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("request body must contain a single JSON object")
		}
		return err
	}

	return nil
}
