package reconciliation

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elkoshar/reconciliation-app/pkg/response"
	"github.com/elkoshar/reconciliation-app/service/reconciliation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockReconciliationService is a mock implementation of ReconciliationService
type MockReconciliationService struct {
	mock.Mock
}

func (m *MockReconciliationService) Reconcile(startDate string, endDate string, sysData io.Reader, attachment *multipart.Form, systemTransactions []reconciliation.SystemTransaction, bankTransactions []reconciliation.BankTransaction) (reconciliation.ReconciliationResult, error) {
	args := m.Called(startDate, endDate, sysData, attachment, systemTransactions, bankTransactions)
	return args.Get(0).(reconciliation.ReconciliationResult), args.Error(1)
}

func TestInit(t *testing.T) {
	mockService := new(MockReconciliationService)
	Init(mockService)
	assert.NotNil(t, reconService)
}

func TestReconciliation_Success(t *testing.T) {
	// Setup mock service
	mockService := new(MockReconciliationService)
	Init(mockService)

	expectedResult := reconciliation.ReconciliationResult{
		TotalProcessed:     10,
		TotalMatched:       8,
		TotalUnmatched:     2,
		TotalDiscrepancies: 100,
		UnmatchedSystem:    []reconciliation.SystemTransaction{},
		UnmatchedBank:      make(map[string][]reconciliation.BankTransaction),
	}

	mockService.On("Reconcile",
		"2025-01-01",
		"2025-01-31",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(expectedResult, nil)

	// Create multipart form request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	writer.WriteField("start_date", "2025-01-01")
	writer.WriteField("end_date", "2025-01-31")

	// Add system_data file
	systemPart, err := writer.CreateFormFile("system_data", "system.csv")
	assert.NoError(t, err)
	systemPart.Write([]byte("trx_id,amount,type,timestamp\nTRX001,100.50,CREDIT,2025-01-15 10:30:00"))

	// Add bank_csv file
	bankPart, err := writer.CreateFormFile("bank_csv", "bank.csv")
	assert.NoError(t, err)
	bankPart.Write([]byte("unique_id,amount,date\nBANK001,100.50,2025-01-15"))

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/reconciliation", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	// Execute handler
	Reconciliation(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp.Error.Status)

	mockService.AssertExpectations(t)
}

func TestReconciliation_ParseMultipartFormError(t *testing.T) {
	mockService := new(MockReconciliationService)
	Init(mockService)

	// Create request without multipart form
	req := httptest.NewRequest(http.MethodPost, "/reconciliation", strings.NewReader("invalid data"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	Reconciliation(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Error.Status)
}

func TestReconciliation_MissingSystemDataFile(t *testing.T) {
	mockService := new(MockReconciliationService)
	Init(mockService)

	// Create multipart form without system_data file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("start_date", "2025-01-01")
	writer.WriteField("end_date", "2025-01-31")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/reconciliation", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	Reconciliation(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Error.Status)
	assert.Contains(t, resp.Error.Msg, "no such file")
}

func TestReconciliation_ServiceError(t *testing.T) {
	mockService := new(MockReconciliationService)
	Init(mockService)

	mockService.On("Reconcile",
		"2025-01-01",
		"2025-01-31",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(reconciliation.ReconciliationResult{}, errors.New("service error"))

	// Create valid multipart form request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("start_date", "2025-01-01")
	writer.WriteField("end_date", "2025-01-31")

	systemPart, err := writer.CreateFormFile("system_data", "system.csv")
	assert.NoError(t, err)
	systemPart.Write([]byte("trx_id,amount,type,timestamp\nTRX001,100.50,CREDIT,2025-01-15 10:30:00"))

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/reconciliation", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	Reconciliation(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp response.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Error.Status)
	assert.Equal(t, "service error", resp.Error.Msg)

	mockService.AssertExpectations(t)
}

func TestReconciliation_InvalidDateFormat(t *testing.T) {
	mockService := new(MockReconciliationService)
	Init(mockService)

	mockService.On("Reconcile",
		"invalid-date",
		"2025-01-31",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(reconciliation.ReconciliationResult{}, errors.New("invalid start_date (expected YYYY-MM-DD)"))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("start_date", "invalid-date")
	writer.WriteField("end_date", "2025-01-31")

	systemPart, err := writer.CreateFormFile("system_data", "system.csv")
	assert.NoError(t, err)
	systemPart.Write([]byte("trx_id,amount,type,timestamp"))

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/reconciliation", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	Reconciliation(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp response.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Error.Status)
	assert.Contains(t, resp.Error.Msg, "invalid start_date")

	mockService.AssertExpectations(t)
}

func TestReconciliation_MultipleBankFiles(t *testing.T) {
	mockService := new(MockReconciliationService)
	Init(mockService)

	expectedResult := reconciliation.ReconciliationResult{
		TotalProcessed:     20,
		TotalMatched:       15,
		TotalUnmatched:     5,
		TotalDiscrepancies: 200,
		UnmatchedSystem:    []reconciliation.SystemTransaction{},
		UnmatchedBank:      make(map[string][]reconciliation.BankTransaction),
	}

	mockService.On("Reconcile",
		"2025-01-01",
		"2025-01-31",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(expectedResult, nil)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("start_date", "2025-01-01")
	writer.WriteField("end_date", "2025-01-31")

	systemPart, err := writer.CreateFormFile("system_data", "system.csv")
	assert.NoError(t, err)
	systemPart.Write([]byte("trx_id,amount,type,timestamp\nTRX001,100.50,CREDIT,2025-01-15 10:30:00"))

	// Add multiple bank files
	bank1Part, err := writer.CreateFormFile("bank_csv", "bank1.csv")
	assert.NoError(t, err)
	bank1Part.Write([]byte("unique_id,amount,date\nBANK001,100.50,2025-01-15"))

	bank2Part, err := writer.CreateFormFile("bank_csv", "bank2.csv")
	assert.NoError(t, err)
	bank2Part.Write([]byte("unique_id,amount,date\nBANK002,50.25,2025-01-16"))

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/reconciliation", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	Reconciliation(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp.Error.Status)

	mockService.AssertExpectations(t)
}

func TestReconciliation_EmptyFormValues(t *testing.T) {
	mockService := new(MockReconciliationService)
	Init(mockService)

	mockService.On("Reconcile",
		"",
		"",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(reconciliation.ReconciliationResult{}, errors.New("invalid start_date (expected YYYY-MM-DD)"))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	systemPart, err := writer.CreateFormFile("system_data", "system.csv")
	assert.NoError(t, err)
	systemPart.Write([]byte("trx_id,amount,type,timestamp"))

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/reconciliation", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	Reconciliation(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp response.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Error.Status)

	mockService.AssertExpectations(t)
}

func TestReconciliation_LargeFile(t *testing.T) {
	mockService := new(MockReconciliationService)
	Init(mockService)

	expectedResult := reconciliation.ReconciliationResult{
		TotalProcessed: 1000,
		TotalMatched:   950,
		TotalUnmatched: 50,
	}

	mockService.On("Reconcile",
		"2025-01-01",
		"2025-01-31",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(expectedResult, nil)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("start_date", "2025-01-01")
	writer.WriteField("end_date", "2025-01-31")

	// Create a larger CSV file
	systemPart, err := writer.CreateFormFile("system_data", "system.csv")
	assert.NoError(t, err)

	csvContent := "trx_id,amount,type,timestamp\n"
	for i := 1; i <= 100; i++ {
		csvContent += "TRX001,100.50,CREDIT,2025-01-15 10:30:00\n"
	}
	systemPart.Write([]byte(csvContent))

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/reconciliation", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	Reconciliation(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp.Error.Status)

	mockService.AssertExpectations(t)
}
