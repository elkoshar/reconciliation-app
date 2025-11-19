package response

import (
	"net/http"
	"time"

	"github.com/go-chi/render"
)

// Response defines http response for the client
type Response struct {
	Code               int            `json:"-"`
	Data               interface{}    `json:"data,omitempty"`
	Error              Error          `json:"error"`
	Message            string         `json:"message"`
	ServerTime         int64          `json:"serverTime"`
	Pagination         interface{}    `json:"pagination,omitempty"`
	StartTime          time.Time      `json:"-"`
	IsStringCode       bool           `json:"-"`
	CodeRender         interface{}    `json:"code"`
	OverrideStatusText map[int]string `json:"-"`
}

func NewResponse() Response {
	return Response{
		StartTime: time.Now(),
		Code:      200,
	}
}

type responseCustom struct {
	Success bool `json:"success"`
}

// Error defines the error
type Error struct {
	Status bool   `json:"status" example:"false"` // true if we have error
	Msg    string `json:"msg" example:" "`        // error message
	Code   int    `json:"code" example:"0"`       // application error code for tracing
}

func NewError(err error, code int) *Error {
	return &Error{
		Status: true,
		Msg:    err.Error(),
		Code:   code,
	}
}

func (e Error) Error() string {
	return e.Msg
}

// SetError set the response to return the given error.
// code is http status code, http.StatusInternalServerError is the default value
func (res *Response) SetError(err error, code ...int) {

	cerr, ok := err.(*Error)
	if ok {
		code = []int{cerr.Code}
	}

	if len(code) > 0 {
		res.Code = code[0]
	} else {
		res.Code = http.StatusInternalServerError
	}

	if err != nil {
		res.Error = Error{
			Msg:    err.Error(),
			Status: true,
			Code:   res.Code,
		}
	}

}

// Render writes the http response to the client.
// statusCode override http status code sent but not the actual response code from Response
func (res *Response) Render(w http.ResponseWriter, r *http.Request, statusCode ...int) {
	if res.Code == 0 {
		res.Code = http.StatusOK
	}

	res.ServerTime = time.Now().Unix()
	render.Status(r, res.Code)

	if len(statusCode) > 0 {
		render.Status(r, statusCode[0])
	} else {
		render.Status(r, res.Code)
	}

	res.RenderStatusCode()
	render.JSON(w, r, res)
}

// only render success
func (res *Response) RenderSuccessOnly(w http.ResponseWriter, r *http.Request) {
	var resp responseCustom
	if res.Code == 0 {
		res.Code = http.StatusOK
	}

	if res.Code == http.StatusOK {
		resp.Success = true
	}

	render.Status(r, res.Code)
	render.JSON(w, r, resp)
}

func (res *Response) MakeStringCode() {
	res.IsStringCode = true
}

func (res *Response) RenderStatusCode() {
	res.CodeRender = res.Code
	if !res.IsStringCode {
		return
	}

	statusText := res.OverrideStatusText[res.Code]
	if statusText == "" {
		statusText = http.StatusText(res.Code)
	}
	res.CodeRender = statusText
}

func (res *Response) AddOverrideStatus(code int, text string) {
	if res.OverrideStatusText == nil {
		res.OverrideStatusText = make(map[int]string)
	}

	res.OverrideStatusText[code] = text
}
