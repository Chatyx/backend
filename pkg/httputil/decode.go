package httputil

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

const (
	decodeQuerySource  = "query"
	decodePathSource   = "path"
	decodeHeaderSource = "header"
)

func DecodeBody(r io.ReadCloser, v any) error {
	defer r.Close()

	if err := json.NewDecoder(r).Decode(v); err != nil {
		return ErrDecodeBodyFailed.Wrap(err)
	}
	return nil
}

type RequestDecoder struct {
	req *http.Request
}

func NewRequestDecoder(req *http.Request) RequestDecoder {
	return RequestDecoder{req: req}
}

func (d RequestDecoder) MergeResults(errs ...error) error {
	for _, err := range errs {
		if err == nil {
			continue
		}

		return err
	}

	return nil
}

func (d RequestDecoder) Body(v any) error {
	defer d.req.Body.Close()

	if err := json.NewDecoder(d.req.Body).Decode(v); err != nil {
		return ErrDecodeBodyFailed.Wrap(err)
	}
	return nil
}

func (d RequestDecoder) Query(key string, val any, defaultVal any) error {
	return d.decodeSingleParam(decodeQuerySource, key, val, defaultVal)
}

func (d RequestDecoder) Path(key string, val any, defaultVal any) error {
	return d.decodeSingleParam(decodePathSource, key, val, defaultVal)
}

func (d RequestDecoder) Header(key string, val any, defaultVal any) error {
	return d.decodeSingleParam(decodeHeaderSource, key, val, defaultVal)
}

func (d RequestDecoder) decodeSingleParam(src, key string, val any, defaultVal any) error {
	var (
		rawVal   string
		preError Error
	)

	switch src {
	case decodeQuerySource:
		rawVal = d.req.URL.Query().Get(key)
		preError = ErrDecodeQueryParamsFailed
	case decodePathSource:
		rawVal = httprouter.ParamsFromContext(d.req.Context()).ByName(key)
		preError = ErrDecodePathParamsFailed
	case decodeHeaderSource:
		rawVal = d.req.Header.Get(key)
		preError = ErrDecodeHeaderParamsFailed
	default:
		return fmt.Errorf("unsupport source type %s to decode", src)
	}

	if rawVal == "" && defaultVal != nil {
		reflect.ValueOf(val).Elem().Set(reflect.ValueOf(defaultVal))
		return nil
	}

	var (
		err error
		v   any
	)

	switch val.(type) {
	case *string:
		v = rawVal
	case *int:
		v, err = strconv.Atoi(rawVal)
	case *float64:
		v, err = strconv.ParseFloat(rawVal, 64)
	case *bool:
		v, err = strconv.ParseBool(rawVal)
	}

	if err != nil {
		err = fmt.Errorf("parse %s: %v", key, err)
		return preError.Wrap(err).WithData(map[string]string{
			key: fmt.Sprintf("failed to parse %s", reflect.TypeOf(val).Elem().Name()),
		})
	}

	reflect.ValueOf(val).Elem().Set(reflect.ValueOf(v))
	return nil
}
