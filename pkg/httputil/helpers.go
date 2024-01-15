package httputil

import (
	"encoding/json"
	"io"
)

func DecodeBody(r io.ReadCloser, v any) error {
	defer r.Close()

	if err := json.NewDecoder(r).Decode(v); err != nil {
		return ErrDecodeBodyFailed.Wrap(err)
	}
	return nil
}
