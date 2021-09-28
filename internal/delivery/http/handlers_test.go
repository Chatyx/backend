package http

import (
	"io/ioutil"
	"net/http"
	"testing"
)

func checkResponseResult(t *testing.T, resp *http.Response, expectedStatusCode int, expectedResponseBody string) {
	t.Helper()

	if resp.StatusCode != expectedStatusCode {
		t.Errorf("Wrong response status code. Expected %d, got %d", expectedStatusCode, resp.StatusCode)
		return
	}

	respBodyPayload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Unexpected error while reading response body: %v", err)
		return
	}

	if string(respBodyPayload) != expectedResponseBody {
		t.Errorf("Wrong response body. Expected %s, got %s", expectedResponseBody, respBodyPayload)
		return
	}
}
