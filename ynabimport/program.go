package ynabimport

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	yb_account "github.com/kalledk/go-ynab/ynab/account"
	yb_api "github.com/kalledk/go-ynab/ynab/api"
	yb_budget "github.com/kalledk/go-ynab/ynab/budget"
	yb_client "github.com/kalledk/go-ynab/ynab/client"
	yb_transaction "github.com/kalledk/go-ynab/ynab/transaction"
)

type StreamLoader interface {
	LoadStream(bank string, reader io.Reader) ([]yb_transaction.Transaction, error)
}

var loaders map[string]StreamLoader = make(map[string]StreamLoader)

func RegisterExtension(extension string, loader StreamLoader) {
	loaders[extension] = loader
}

type StreamLoaderFunc func(bank string, reader io.Reader) ([]yb_transaction.Transaction, error)

func (s StreamLoaderFunc) LoadStream(bank string, reader io.Reader) ([]yb_transaction.Transaction, error) {
	return s(bank, reader)
}

func LoadFile(bank string, path string) (transactions []yb_transaction.Transaction, err error) {

	fp, err := os.Open(path)
	if err != nil {
		return
	}
	defer fp.Close()

	ext := filepath.Ext(path)
	loader, ok := loaders[ext]
	if !ok {
		err = fmt.Errorf("no valid filetype loader for %v", ext)
		return
	}

	return loader.LoadStream(bank, fp)
}

type PayeeConverter struct {
	Name   string
	Regexp *regexp.Regexp
}

type PayeeCollection struct {
	converters []*PayeeConverter
}

func (pc *PayeeCollection) Convert(payee string) string {
	for _, converter := range pc.converters {
		if converter.Regexp.MatchString(payee) {
			return converter.Name
		}
	}
	return ""
}

func (pc *PayeeCollection) Contains(payee string) bool {
	for _, converter := range pc.converters {
		if converter.Regexp.MatchString(payee) {
			return true
		}
	}
	return false
}

func UploadTransactions(accessToken yb_api.AccessToken, budgetID yb_budget.ID, accountID yb_account.ID, transactions []yb_transaction.Transaction) (trd []string, trs []yb_transaction.Transaction, err error) {
	client := yb_client.NewClient(accessToken)
	transactionClient := client.Budgets().Budget(budgetID).Transactions()

	for i := range transactions {
		transactions[i].AccountID = accountID
	}

	reply, err := transactionClient.AddList(transactions)
	if err != nil {
		return nil, nil, err
	}

	return reply.DuplicateImportIDs, reply.Transactions, nil
}

func ConvertPayees(converters *PayeeCollection, ts []yb_transaction.Transaction) {
	for i := range ts {
		newName := converters.Convert(ts[i].PayeeName)
		if len(newName) == 0 {
			ts[i].Memo = ts[i].PayeeName
		}
		ts[i].PayeeName = newName
	}
}

func ListUnknownPayees(converters *PayeeCollection, ts []yb_transaction.Transaction) {
	fmt.Println("\nunknown payees")
	for _, t := range ts {
		if !converters.Contains(t.PayeeName) {
			fmt.Printf(" - | %-30s |\n", t.PayeeName)
		}
	}
}

func NewPayeeCollection(path string) (pc *PayeeCollection, err error) {
	var payeemap []struct {
		Name   string
		Regexp string
	}

	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	decoder := json.NewDecoder(fp)
	decoder.Decode(&payeemap)

	pc = &PayeeCollection{}
	pc.converters = make([]*PayeeConverter, len(payeemap))

	for i, p := range payeemap {
		pc.converters[i] = &PayeeConverter{
			Name:   p.Name,
			Regexp: regexp.MustCompile(p.Regexp),
		}

	}

	return pc, nil
}
