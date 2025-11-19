package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRender(t *testing.T) {
	testCases := []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request)
		status  int
	}{
		{
			name: "OK",
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := Response{}
				resp.MakeStringCode()
				resp.Render(w, r)
			},
			status: http.StatusOK,
		},
		{
			name: "bad request error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := Response{}
				resp.MakeStringCode()
				resp.SetError(fmt.Errorf("bad request"), http.StatusBadRequest)
				resp.Render(w, r)
			},
			status: http.StatusBadRequest,
		},
		{
			name: "default http code on error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := Response{}
				resp.SetError(fmt.Errorf("internal server error"))
				resp.Render(w, r)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tc.handler))
			defer server.Close()

			resp, err := http.Get(server.URL)
			require.NoError(t, err)

			require.Equal(t, tc.status, resp.StatusCode)
			require.Contains(t, resp.Header.Get("Content-Type"), "application/json")
		})
	}

}

func TestErrorMsg(t *testing.T) {
	const (
		someErrorMsg = "some error msg"
	)
	testCases := []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request)
		status  int
		err     Error
	}{
		{
			name: "with error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := Response{}
				resp.SetError(fmt.Errorf(someErrorMsg), http.StatusBadRequest)
				resp.Render(w, r)
			},
			status: http.StatusBadRequest,
			err: Error{
				Msg:    someErrorMsg,
				Status: true,
				Code:   400,
			},
		},
		{
			name: "no error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := Response{}
				resp.Render(w, r)
			},
			status: http.StatusOK,
			err:    Error{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tc.handler))
			defer server.Close()

			resp, err := http.Get(server.URL)
			require.NoError(t, err)

			require.Equal(t, tc.status, resp.StatusCode)

			var respBody Response
			err = json.NewDecoder(resp.Body).Decode(&respBody)
			require.NoError(t, err)

			require.Equal(t, tc.err, respBody.Error)
		})
	}

}

func TestRenderSuccessOnly(t *testing.T) {
	testCases := []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request)
		status  int
	}{
		{
			name: "OK",
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := Response{}
				resp.RenderSuccessOnly(w, r)
			},
			status: http.StatusOK,
		},
		{
			name: "bad request error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := Response{}
				resp.SetError(fmt.Errorf("bad request"), http.StatusBadRequest)
				resp.RenderSuccessOnly(w, r)
			},
			status: http.StatusBadRequest,
		},
		{
			name: "default http code on error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := Response{}
				resp.SetError(fmt.Errorf("internal server error"))
				resp.RenderSuccessOnly(w, r)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tc.handler))
			defer server.Close()

			resp, err := http.Get(server.URL)
			require.NoError(t, err)

			require.Equal(t, tc.status, resp.StatusCode)
			require.Contains(t, resp.Header.Get("Content-Type"), "application/json")
		})
	}

}

func TestSetError(t *testing.T) {
	errors := NewError(fmt.Errorf("some error"), http.StatusBadRequest)
	response := Response{}

	response.SetError(errors)
	require.Equal(t, http.StatusBadRequest, response.Code)

}

func TestAddOverrideStatus(t *testing.T) {
	response := Response{}
	response.AddOverrideStatus(200, "test")
	response.Code = 200
	response.MakeStringCode()
	response.RenderStatusCode()
	require.Equal(t, "test", response.CodeRender)
}
