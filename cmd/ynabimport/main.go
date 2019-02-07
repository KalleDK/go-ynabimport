package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"

	yb_account "github.com/kalledk/go-ynab/ynab/account"
	yb_api "github.com/kalledk/go-ynab/ynab/api"
	yb_budget "github.com/kalledk/go-ynab/ynab/budget"
	yb_transaction "github.com/kalledk/go-ynab/ynab/transaction"

	"github.com/kalledk/go-ynabimport/ynabimport"
	_ "github.com/kalledk/go-ynabimport/ynabimport/jbank"
)

func SprintJSON(model interface{}) string {
	data, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	return string(data)
}

func Upload(ts []yb_transaction.Transaction) ([]string, []yb_transaction.Transaction) {
	accessToken, _ := yb_api.NewAccessToken(os.Getenv("YNAB_TOKEN"))
	budgetID, _ := yb_budget.NewID(os.Getenv("YNAB_BUDGET"))
	accountID, _ := yb_account.NewID(os.Getenv("YNAB_ACCOUNT"))

	doubleID, tsr, err := ynabimport.UploadTransactions(accessToken, budgetID, accountID, ts)
	if err != nil {
		log.Fatal(err)
	}

	return doubleID, tsr
}

func listContains(list []string, val string) bool {
	for _, val := range list {
		if val == val {
			return true
		}
	}
	return false
}

func main() {
	ts, err := ynabimport.LoadFile("jbank", "demo.csv")
	if err != nil {
		log.Fatal(err)
	}

	payeeConverters, err := ynabimport.NewPayeeCollection("payees.json")
	if err != nil {
		log.Fatal(err)
	}

	ynabimport.ListUnknownPayees(payeeConverters, ts)

	keyreader := bufio.NewReader(os.Stdin)
	char, _, err := keyreader.ReadRune()

	if err != nil {
		log.Fatal(err)
	}

	switch char {
	case '\r':
		return
	case 'N':
		return
	case 'n':
		return
	}

	for char != '\r' {
		char, _, err = keyreader.ReadRune()
		if err != nil {
			log.Fatal(err)
		}
	}

	ynabimport.ConvertPayees(payeeConverters, ts)

	fmt.Printf("\nfound %v transactions\n", len(ts))
	for _, t := range ts {
		fmt.Printf(" - | %v | %-30s | %10.2f kr. | %-30s |\n", t.Date, t.PayeeName, float64(t.Amount)/1000, t.Memo)
	}

	fmt.Printf("\nUpload this y/n [n]: ")

	char, _, err = keyreader.ReadRune()

	if err != nil {
		log.Fatal(err)
	}

	switch char {
	case '\r':
		return
	case 'N':
		return
	case 'n':
		return
	}

	for char != '\r' {
		char, _, err = keyreader.ReadRune()
		if err != nil {
			log.Fatal(err)
		}
	}

	dids, rts := Upload(ts)

	for _, t := range rts {
		var dup string
		if listContains(dids, t.ID.String()) {
			dup = "duplicate"
		}
		fmt.Printf(" - | %v | %-30s | %10.2f kr. | %v | %-9s |\n", t.Date, t.PayeeName, float64(t.Amount)/1000, t.ID.String()[:10], dup)
	}

}
