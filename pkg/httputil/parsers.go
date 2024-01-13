package httputil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/Chatyx/backend/pkg/log"
	"github.com/Chatyx/backend/pkg/validator"

	"github.com/julienschmidt/httprouter"
)

type BodyParser struct {
	validate validator.Validate
}

func NewBodyParser(v validator.Validate) BodyParser {
	return BodyParser{validate: v}
}

func (p BodyParser) Parse(ctx context.Context, req *http.Request, dst any) error {
	logger := log.FromContext(ctx)
	defer req.Body.Close()

	if err := json.NewDecoder(req.Body).Decode(dst); err != nil {
		logger.WithError(err).Debug("Failed to decode body")
		return ErrDecodeBodyFailed.Wrap(err)
	}

	if err := p.validate.Struct(dst); err != nil {
		ve := validator.Error{}
		if errors.As(err, &ve) {
			logger.WithError(err).Debug("Validation failed")

			data := make(map[string]any, len(ve.Fields))
			for field, reason := range ve.Fields {
				data[field] = reason
			}

			return ErrValidationFailed.WithData(data)
		}

		return ErrInternalServer.Wrap(err)
	}

	return nil
}

type singleValueParser struct {
	key      string
	tag      string
	validate validator.Validate
}

func (p singleValueParser) parse(ctx context.Context, rawVal string, dst any) error {
	logger := log.FromContext(ctx)

	var (
		val any
		err error
	)

	switch dst.(type) {
	case *int:
		var i int

		i, err = strconv.Atoi(rawVal)
		val = i
	case *bool:
		var b bool

		b, err = strconv.ParseBool(rawVal)
		val = b
	case *float64:
		var f float64

		f, err = strconv.ParseFloat(rawVal, 64)
		val = f
	case *string:
		val = rawVal
	default:
		return fmt.Errorf("unsupport type %T to parse", dst)
	}

	if rawVal != "" && err != nil {
		logger.WithError(err).Debug("Failed to parse param `%s`", p.key)
		return ErrDecodeParamsFailed.WithData(map[string]any{
			p.key: "failed to parse",
		})
	}

	if err = p.validate.Var(val, p.tag); err != nil {
		ve := validator.Error{}
		if errors.As(err, &ve) {
			logger.WithError(err).Debug("Validation failed")

			data := make(map[string]any, len(ve.Fields))
			for field, reason := range ve.Fields {
				data[field] = reason
			}

			return ErrValidationFailed.WithData(data)
		}

		return ErrInternalServer.Wrap(err)
	}

	reflect.ValueOf(dst).Elem().Set(reflect.ValueOf(val))
	return nil
}

type HeaderParser struct {
	key        string
	baseParser singleValueParser
}

func NewHeaderParser(v validator.Validate, key, tag string) HeaderParser {
	return HeaderParser{
		key: key,
		baseParser: singleValueParser{
			key:      key,
			tag:      tag,
			validate: v,
		},
	}
}

func (p HeaderParser) Parse(ctx context.Context, req *http.Request, dst any) error {
	val := req.Header.Get(p.key)
	return p.baseParser.parse(ctx, val, dst)
}

type QueryParser struct {
	key        string
	baseParser singleValueParser
}

func NewQueryParser(v validator.Validate, key, tag string) QueryParser {
	return QueryParser{
		key: key,
		baseParser: singleValueParser{
			key:      key,
			tag:      tag,
			validate: v,
		},
	}
}

func (p QueryParser) Parse(ctx context.Context, req *http.Request, dst any) error {
	val := req.URL.Query().Get(p.key)
	return p.baseParser.parse(ctx, val, dst)
}

type PathParser struct {
	key        string
	baseParser singleValueParser
}

func NewPathParser(v validator.Validate, key, tag string) PathParser {
	return PathParser{
		key: key,
		baseParser: singleValueParser{
			key:      key,
			tag:      tag,
			validate: v,
		},
	}
}

func (p PathParser) Parse(ctx context.Context, req *http.Request, dst any) error {
	val := httprouter.ParamsFromContext(req.Context()).ByName(p.key)
	return p.baseParser.parse(ctx, val, dst)
}
