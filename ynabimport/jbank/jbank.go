package jbank

import (
	"encoding/csv"
	"fmt"
	"hash/crc32"
	"math"
	"strings"

	jb_transaction "github.com/kalledk/go-jbank/jbank/transaction"
	yb_transaction "github.com/kalledk/go-ynab/ynab/transaction"
	yb_csv "github.com/kalledk/go-ynabimport/ynabimport/csv"
)

func init() {
	yb_csv.Register("jbank", csvParser)
}

var csvParser yb_csv.ParserFunc = UnmarshalCSV

func TrimValueLength(val int64, n uint) int64 {
	max := int64(math.Pow10(int(n - 1)))
	min := -1 * max
	for val < min || max < val {
		val = val / 10
	}
	return val
}

func TrimStringLength(val string, n int) string {
	if len(val) > n {
		return val[:n]
	}
	return val
}

func MakeChecksum(balance int64, text string) string {
	base := []byte(fmt.Sprintf("%v%v", balance, text))
	return fmt.Sprintf("%08x\n", crc32.Checksum(base, crc32.IEEETable))
}

func MakeImportID(src jb_transaction.Transaction) string {

	date := src.Date.Format("20060102")
	amount := TrimValueLength(src.Amount, 6)
	checksum := MakeChecksum(src.Balance, src.Text)

	id := fmt.Sprintf("jb:%s:%+d:%s", date, amount, checksum)

	result := strings.Replace(id, " ", "_", -1)
	if len(result) > 26 {
		result = result[:26]
	}
	return result
}

func Convert(src jb_transaction.Transaction) (dst yb_transaction.Transaction) {
	return yb_transaction.Transaction{
		Date:      src.Date.Format("2006-01-02"),
		Amount:    src.Amount * 10,
		PayeeName: src.Text,
		ImportID:  MakeImportID(src),
	}
}

func UnmarshalCSV(reader *csv.Reader, transactions interface{}) (err error) {
	yts := transactions.(*[]yb_transaction.Transaction)

	jts, err := jb_transaction.FromCSV(reader)
	if err != nil {
		return err
	}

	for _, jt := range jts {
		*yts = append(*yts, Convert(jt))
	}

	return err
}
