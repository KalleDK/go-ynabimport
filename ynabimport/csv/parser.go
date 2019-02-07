package csv

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/kalledk/go-ynabimport/ynabimport"

	yb_transaction "github.com/kalledk/go-ynab/ynab/transaction"
)

type Parser interface {
	UnmarshalCSV(reader *csv.Reader, transactions interface{}) (err error)
}

var parsers map[string]Parser = make(map[string]Parser)

type ParserFunc func(reader *csv.Reader, transactions interface{}) error

func (c ParserFunc) UnmarshalCSV(reader *csv.Reader, transactions interface{}) error {
	return c(reader, transactions)
}

func Register(bank string, reader Parser) {
	parsers[bank] = reader
}

var loadStream ynabimport.StreamLoaderFunc = LoadStream

func init() {
	ynabimport.RegisterExtension(".csv", loadStream)
}

func LoadStream(bank string, reader io.Reader) (transactions []yb_transaction.Transaction, err error) {

	parser, ok := parsers[bank]
	if !ok {
		err = fmt.Errorf("no valid bank parser for %v", bank)
		return
	}

	csvReader := csv.NewReader(reader)
	err = parser.UnmarshalCSV(csvReader, &transactions)

	return

}
