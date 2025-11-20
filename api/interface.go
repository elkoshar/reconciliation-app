package api

import (
	"io"
	"mime/multipart"

	"github.com/elkoshar/reconciliation-app/service/reconciliation"
)

type ReconciliationService interface {
	Reconcile(startDate string, endDate string, sysData io.Reader, attachement *multipart.Form, systemTransactions []reconciliation.SystemTransaction, bankTransactions []reconciliation.BankTransaction) (reconciliation.ReconciliationResult, error)
}
