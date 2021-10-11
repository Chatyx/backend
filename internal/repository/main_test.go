// +build unit

package repository

import "errors"

type Row []interface{}

type RowResult struct {
	Row Row
	Err error
}

var errUnexpected = errors.New("unexpected error")
