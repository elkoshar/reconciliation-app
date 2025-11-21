package reconciliation

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/elkoshar/reconciliation-app/api"
	"github.com/elkoshar/reconciliation-app/pkg/response"
	"github.com/elkoshar/reconciliation-app/service/reconciliation"
)

var (
	reconService api.ReconciliationService
)

const (
	ErrParseUrlParamMsg = "Parse Url Param Failed. %v"
	ErrCreateDataMsg    = "Create Data Failed. %+v"
	ErrParseValidateMsg = "Failed to Parse and Validate. err=%v"
)

func Init(service api.ReconciliationService) {
	reconService = service
}

// Reconciliation : HTTP Handler for process reconciliation
// @Summary Reconciliation Process
// @Description Reconciliation handles request for reconciliation process
// @Tags Reconciliation
// @Accept multipart/form-data
// @Produce json
// @Param Accept-Language header string true "accept language" default(id)
// @Param start_date formData string true "start date format YYYY-MM-DD" example(2023-01-01)
// @Param end_date formData string true "end date format YYYY-MM-DD" example(2023-01-31)
// @Param system_data formData file true "system data file upload"
// @Param bank_csv formData file false "bank CSV file upload"
// @Success 200 {object} response.Response{data=reconciliation.ReconciliationResult} "Success Response"
// @Failure 400 "Bad Request"
// @Failure 500 "InternalServerError"
// @Router /reconciliation [post]
func Reconciliation(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.Render(w, r)

	var (
		err    error
		result reconciliation.ReconciliationResult
	)

	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		slog.WarnContext(r.Context(), fmt.Sprintf("Parse Multipart Form Failed. err=%v", err))
		resp.SetError(err, http.StatusBadRequest)
		return
	}

	startDate := r.FormValue("start_date")
	endDate := r.FormValue("end_date")

	sysFile, _, err := r.FormFile("system_data")
	if err != nil {
		slog.WarnContext(r.Context(), fmt.Sprintf("Get System Data File Failed. err=%v", err))
		resp.SetError(err, http.StatusBadRequest)
		return
	}
	defer sysFile.Close()

	result, err = reconService.Reconcile(startDate, endDate, sysFile, r.MultipartForm, nil, nil)
	if err != nil {
		slog.WarnContext(r.Context(), fmt.Sprintf("Reconciliation Process Failed. err=%v", err))
		resp.SetError(err, http.StatusInternalServerError)
		return
	}

	resp.Data = result
}
