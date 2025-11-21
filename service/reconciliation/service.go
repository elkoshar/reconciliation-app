package reconciliation

import (
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"strconv"
	"strings"
	"time"
)

const (
	SystemTimeFormat = "2006-01-02 15:04:05"
	BankTimeFormat   = "2006-01-02"
)

type reconciliationService struct {
}

type ReconciliationService interface {
	Reconcile(startDate string, endDate string, sysData io.Reader, attachement *multipart.Form, systemTransactions []SystemTransaction, bankTransactions []BankTransaction) (ReconciliationResult, error)
}

func NewReconciliationService() ReconciliationService {
	return &reconciliationService{}
}

func (s *reconciliationService) Reconcile(startDate string, endDate string, sysData io.Reader, attachement *multipart.Form, systemTransactions []SystemTransaction, bankTransactions []BankTransaction) (res ReconciliationResult, err error) {

	startTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return ReconciliationResult{}, fmt.Errorf("invalid start_date (expected YYYY-MM-DD)")
	}
	endTime, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return ReconciliationResult{}, fmt.Errorf("invalid end_date (expected YYYY-MM-DD)")
	}

	endTime = endTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	sysTrx, err := LoadSystemTransactions(sysData, startTime, endTime)
	if err != nil {
		return ReconciliationResult{}, fmt.Errorf("failed to load system transactions: %v", err)
	}

	var allBankTrx []BankTransaction

	for _, fileHeader := range attachement.File["bank_csv"] {
		f, err := fileHeader.Open()
		if err != nil {
			continue
		}
		defer f.Close()

		bankName := fmt.Sprintf("Stmt-%s", fileHeader.Filename)
		bTrx, err := LoadBankStatement(f, bankName, startTime, endTime)
		if err == nil {
			allBankTrx = append(allBankTrx, bTrx...)
		}
	}

	return reconcileProcess(sysTrx, allBankTrx), nil
}

func reconcileProcess(systemTransactions []SystemTransaction, bankTransactions []BankTransaction) (result ReconciliationResult) {
	result = ReconciliationResult{
		UnmatchedBank:  make(map[string][]BankTransaction),
		TotalProcessed: len(systemTransactions) + len(bankTransactions),
	}
	bankMap := make(map[string][]int)
	matchedBanks := make(map[int]bool)

	for i, b := range bankTransactions {
		key := generateKey(b.Date, b.Amount)
		bankMap[key] = append(bankMap[key], i)
	}
	var stillUnmatchedSystem []SystemTransaction

	for _, sys := range systemTransactions {
		sysDate := time.Date(sys.TransactionTime.Year(), sys.TransactionTime.Month(), sys.TransactionTime.Day(), 0, 0, 0, 0, time.UTC)
		finalAmount := getSignedAmount(sys)

		key := generateKey(sysDate, finalAmount)

		indices, exists := bankMap[key]
		matched := false

		if exists {
			for _, idx := range indices {
				if !matchedBanks[idx] {
					matchedBanks[idx] = true
					matched = true
					result.TotalMatched++

					break
				}
			}
		}

		if !matched {
			stillUnmatchedSystem = append(stillUnmatchedSystem, sys)
		}
	}

	bankMapByDate := make(map[string][]int)
	for i, b := range bankTransactions {
		if !matchedBanks[i] {
			dKey := b.Date.Format("2006-01-02")
			bankMapByDate[dKey] = append(bankMapByDate[dKey], i)
		}
	}

	for _, sys := range stillUnmatchedSystem {
		sysDateKey := sys.TransactionTime.Format("2006-01-02")
		foundDiscrepancy := false

		if indices, exists := bankMapByDate[sysDateKey]; exists {
			for _, idx := range indices {
				if !matchedBanks[idx] {
					bankTrx := bankTransactions[idx]
					sysSignedAmount := getSignedAmount(sys)

					diff := sysSignedAmount - bankTrx.Amount
					if diff < 0 {
						diff = -diff
					}

					fmt.Printf("[Discrepancy] Date: %s | SysID: %s | Diff: %.2f",
						sysDateKey, sys.TransactionID, diff.ToFloat())

					result.TotalDiscrepancies += diff
					result.TotalMatched++
					matchedBanks[idx] = true
					foundDiscrepancy = true
					break
				}
			}
		}

		if !foundDiscrepancy {
			result.UnmatchedSystem = append(result.UnmatchedSystem, sys)
		}
	}

	for i, b := range bankTransactions {
		if !matchedBanks[i] {
			result.UnmatchedBank[b.BankName] = append(result.UnmatchedBank[b.BankName], b)
		}
	}

	result.TotalUnmatched = len(result.UnmatchedSystem)
	for _, v := range result.UnmatchedBank {
		result.TotalUnmatched += len(v)
	}

	return
}

func LoadSystemTransactions(r io.Reader, start, end time.Time) ([]SystemTransaction, error) {
	csvReader := csv.NewReader(r)

	csvReader.TrimLeadingSpace = true

	// Skip header
	if _, err := csvReader.Read(); err != nil {
		return nil, err
	}

	var trxs []SystemTransaction

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		tTime, _ := time.Parse(SystemTimeFormat, record[3])

		if tTime.Before(start) || tTime.After(end) {
			continue
		}

		amountFloat, _ := strconv.ParseFloat(record[1], 64)

		trxs = append(trxs, SystemTransaction{
			TransactionID:   record[0],
			Amount:          ToMoney(amountFloat),
			Type:            TransactionType(strings.ToUpper(record[2])),
			TransactionTime: tTime,
		})
	}
	return trxs, nil
}

func LoadBankStatement(r io.Reader, bankName string, start, end time.Time) ([]BankTransaction, error) {
	csvReader := csv.NewReader(r)

	csvReader.TrimLeadingSpace = true

	if _, err := csvReader.Read(); err != nil {
		return nil, err
	}

	var trxs []BankTransaction

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}

		dTime, _ := time.Parse(BankTimeFormat, record[2])

		if dTime.Before(start) || dTime.After(end) {
			continue
		}

		amountFloat, _ := strconv.ParseFloat(record[1], 64)

		trxs = append(trxs, BankTransaction{
			BankName: bankName,
			UniqueID: record[0],
			Amount:   ToMoney(amountFloat),
			Date:     dTime,
		})
	}
	return trxs, nil
}

func generateKey(date time.Time, amount Money) string {
	// Format: YYYY-MM-DD hh:mm-amount
	return fmt.Sprintf("%s-%d", date.Format("2006-01-02 15:04"), amount)
}

func getSignedAmount(sys SystemTransaction) Money {
	rawAmount := sys.Amount
	if rawAmount < 0 {
		rawAmount = -rawAmount
	}

	cleanType := strings.ToUpper(strings.TrimSpace(string(sys.Type)))

	if cleanType == "DEBIT" {
		return -rawAmount
	}
	return rawAmount
}
