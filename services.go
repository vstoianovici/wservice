package wservice

import (
	"bufio"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	// importing for side-effects puposes only (init)
	_ "github.com/lib/pq"
)

// WalletService provides operations on accounts.
type WalletService interface {
	GetTable(string) ([]string, error)
	DoTransfer(string, string, string) (string, error)
}

type sqlDBTx struct {
	sqlDriver      string
	sqlHost        string
	sqlPort        string
	sqlUser        string
	sqlPassword    string
	sqlDbName      string
	sslmode        string
	accountsTable  string
	transfersTable string
}

// CreateService exported to be accessable from outside the package (from main)
func NewService() (WalletService, int, error) {
	var fileName string
	flag.StringVar(&fileName, "file", "./postgresql.cfg", "Path of postgresql config file to be parsed.")
	var portNumber int
	flag.IntVar(&portNumber, "port", 8080, "Port on which the server will listen and serve.")
	flag.Parse()

	file, err := os.Open(fileName)
	if err != nil {
		sError := "There was a problem opening file " + fileName + " "
		var ErrReadFile = errors.New(sError)
		cErr := errors.New(ErrReadFile.Error() + err.Error())
		var d = sqlDBTx{}
		return d, portNumber, cErr
	}
	// make sure we eventually close the CSV file
	defer file.Close()
	var configStruct sqlDBTx
	cSlice := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		item := scanner.Text()
		s := strings.Split(item, " : ")
		v := strings.Split(s[1], ",")
		cSlice = append(cSlice, v[0])
	}
	if err := scanner.Err(); err != nil {
		var d = sqlDBTx{}
		return d, portNumber, err
	}
	configStruct = sqlDBTx{
		sqlDriver:      cSlice[0],
		sqlHost:        cSlice[1],
		sqlPort:        cSlice[2],
		sqlUser:        cSlice[3],
		sqlPassword:    cSlice[4],
		sqlDbName:      cSlice[5],
		sslmode:        cSlice[6],
		accountsTable:  cSlice[7],
		transfersTable: cSlice[8],
	}
	return configStruct, portNumber, nil
}

func (s sqlDBTx) GetTable(t string) ([]string, error) {
	connectionString := "host=" + s.sqlHost + " port=" + s.sqlPort + " user=" + s.sqlUser + " password=" + s.sqlPassword + " dbname=" + s.sqlDbName + " sslmode=" + s.sslmode
	db, err := sqlx.Open(s.sqlDriver, connectionString)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var isCommitted = false
	var results []string
	for ok := true; ok; ok = !isCommitted {
		tx, err := db.Begin()
		if err != nil {
			var ErrStartTx = errors.New("err: error begining transaction in postgres")
			cErr := errors.New(ErrStartTx.Error() + err.Error())
			return nil, cErr
		}
		defer tx.Rollback()

		_, err = tx.Exec(`set transaction isolation level serializable`) // <=== SET ISOLATION LEVEL
		//_, err = tx.Exec(`set transaction isolation level repeatable read`) // <=== SET ISOLATION LEVEL
		if err != nil {
			return nil, err
		}
		var txString string
		if t == s.accountsTable {
			txString = "SELECT * FROM " + s.accountsTable + " ORDER BY AccountID;"
		} else {
			txString = "SELECT * FROM " + s.transfersTable + ";"
		}
		rows, err := tx.Query(txString)
		if err != nil {
			if err == sql.ErrNoRows {
				if t == s.accountsTable {
					var ErrAcc = errors.New("err: there are no defined accounts")
					cErr := errors.New(ErrAcc.Error() + err.Error())
					return nil, cErr
				}
				var ErrAcc = errors.New("err: there are no defined transfers")
				cErr := errors.New(ErrAcc.Error() + err.Error())
				return nil, cErr
			}
			var ErrUnexp = errors.New("err: Unexpected error occured")
			return nil, ErrUnexp
		}
		for rows.Next() {
			if t == s.accountsTable {
				var accountID string
				var balance float64
				var currency string
				var initialBalance float64

				if err := rows.Scan(&accountID, &balance, &currency, &initialBalance); err != nil {
					log.Fatal(err)
				}
				sBalance := fmt.Sprintf("%f", balance)
				sIBalance := fmt.Sprintf("%f", initialBalance)
				rString := "Account: " + accountID + "  Balance = " + sBalance + " " + currency + "  Initial Balance = " + sIBalance
				results = append(results, rString)

			} else {
				var paymentID int
				var fromAccount string
				var toAccount string
				var amount float64
				var currency string
				var time string

				if err := rows.Scan(&paymentID, &fromAccount, &toAccount, &amount, &currency, &time); err != nil {
					log.Fatal(err)
				}
				sPayment := fmt.Sprintf("%d", paymentID)
				sAmount := fmt.Sprintf("%f", amount)
				rString := "Transfer #" + sPayment + "  from: " + fromAccount + "  to:  " + toAccount + " in the amount of " + sAmount + " " + currency + " at " + time
				results = append(results, rString)
			}
		}
		tx.Commit()
		// if t == s.accountsTable {
		// 	log.Println("Accounts listing successful.")
		// } else {
		//log.Println("Transfers listing successful.")
		// }
		isCommitted = true
	}
	return results, nil
}

func (s sqlDBTx) DoTransfer(fromAccount string, toAccount string, transferAmount string) (string, error) {
	connectionString := "host=" + s.sqlHost + " port=" + s.sqlPort + " user=" + s.sqlUser + " password=" + s.sqlPassword + " dbname=" + s.sqlDbName + " sslmode=" + s.sslmode
	db, err := sqlx.Open(s.sqlDriver, connectionString)
	if err != nil {
		log.Println("err", err)
		return "error", err
	}

	if fromAccount == toAccount {
		var ErrSameAcc = errors.New("the source account is the same as the destination account. ")
		//log.Println("err", ErrSameAcc)
		return "error", ErrSameAcc
	}

	defer db.Close()
	var isCommitted = false

	for ok := true; ok; ok = !isCommitted {
		tx, err := db.Begin()
		if err != nil {
			var ErrStartTx = errors.New("err: error begining transaction in postgres")
			cErr := errors.New(ErrStartTx.Error() + err.Error())
			return "error", cErr
		}

		defer tx.Rollback()
		_, err = tx.Exec(`set transaction isolation level serializable`) // <=== SET ISOLATION LEVEL
		//_, err = tx.Exec(`set transaction isolation level repeatable read`) // <=== SET ISOLATION LEVEL
		if err != nil {
			return "error", err
		}

		_, err = tx.Exec("LOCK TABLE Accounts IN SHARE ROW EXCLUSIVE MODE;") // <=== Lock table
		if err != nil {
			log.Println(err, "...continuing...")
			continue
		}

		//var id int
		var sBalance string
		var sCurrency string
		txString := "SELECT Balance , Currency FROM " + s.accountsTable + " WHERE AccountID ='" + fromAccount + "';"
		err = tx.QueryRow(txString).Scan(&sBalance, &sCurrency)
		if err != nil {
			if err == sql.ErrNoRows {
				var ErrNoSource = errors.New("The source account does not exist")
				return "error", ErrNoSource
			}
			var ErrUnexpect = errors.New("err: unexpected error")
			cErr := errors.New(ErrUnexpect.Error() + err.Error())
			return "error", cErr
		}
		fBalance, err := strconv.ParseFloat(sBalance, 64)
		if err != nil {
			var ErrParse = errors.New("Error parsing blance")
			return "error", ErrParse
		}

		fAmount, err := strconv.ParseFloat(transferAmount, 64)
		if err != nil {
			var ErrParse = errors.New("err: error begining transaction in postgres")
			cErr := errors.New(ErrParse.Error() + err.Error())
			return "error", cErr
		}

		if fBalance < fAmount {
			var ErrBalance = errors.New("Balance insuficient for transaction")
			return "error", ErrBalance
		}

		//var id int
		var dCurrency string
		txString = "SELECT Currency FROM " + s.accountsTable + " WHERE AccountID ='" + toAccount + "';"
		err = tx.QueryRow(txString).Scan(&dCurrency)
		if err != nil {
			if err == sql.ErrNoRows {
				var ErrNoSource = errors.New("The destination account does not exist")
				return "error", ErrNoSource
			}
			return "error", err
		}

		if dCurrency != sCurrency {
			var ErrMissmatch = errors.New("Not same currency in transaction source and destination")
			return "error", ErrMissmatch
		}

		txString = "UPDATE " + s.accountsTable + " SET balance = balance - " + transferAmount + " WHERE accountid = '" + fromAccount + "';"
		_, err = tx.Exec(txString)
		if err != nil {
			if strings.Contains(err.Error(), "could not serialize access due to") {
				log.Println(err, "...continuing...")
				continue
			} else {
				if strings.Contains(err.Error(), "new row for relation \"accounts\" violates check constraint") {
					var ErrParse = errors.New("err: Please check available balance before making transactions. ")
					return "error", ErrParse
				}
				return "error", err
			}
		}

		txString = "UPDATE " + s.accountsTable + " SET balance = balance + " + transferAmount + " WHERE accountid= '" + toAccount + "';"
		_, err = tx.Exec(txString)
		if err != nil {
			if strings.Contains(err.Error(), "could not serialize access due to") {
				continue
			} else {
				return "error", err
			}
		}
		t0 := time.Now().Format(time.RFC3339)
		txString = "INSERT INTO " + s.transfersTable + " (transid, From_Account, To_Account, Amount, Currency, TTime) VALUES( nextval('Payment_counter'), '" + fromAccount + "', '" + toAccount + "', '" + transferAmount + "', '" + sCurrency + "', '" + t0 + "' );"
		_, err = tx.Exec(txString)
		if err != nil {
			if strings.Contains(err.Error(), "could not serialize access due to") {
				continue
			} else {
				return "error", err
			}
		}
		tx.Commit()
		//log.Printf("Transfer from %s to %s of %s %s successful.", fromAccount, toAccount, transferAmount, sCurrency)
		isCommitted = true
	}
	return "success", nil
}
