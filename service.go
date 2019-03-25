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
	// importing as blank for side-effects puposes only (init)
	_ "github.com/lib/pq"
)

// This is where the core bussiness logic resides (the service layer of the gokit onion) and on top of it we will be layering other functionalities that go-kit helps with

// WalletService is the inteface to be used from outside the package that provides
// operations on accounts.
type WalletService interface {
	GetTable(string) ([]string, error)
	DoTransfer(string, string, string) (string, error)
}

// sqlDBTx is a type that defines the necessary information to establish a Postgres
// database connection and what tables to access (structure of the DB)
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

// NewService exported to be accessable from outside the package (from main)
// NewService is necessary because we need the ability to create a sqlDBTx stuct from outside the package (like from main)
func NewService() (WalletService, int, error) {
	var fileName string
	// Parse the postrgres configuration file name and path. if not deifned the default is "postgresql.cfg" from /cmd
	flag.StringVar(&fileName, "file", "./postgresql.cfg", "Path of postgresql config file to be parsed.")
	var portNumber int
	// Parse the port number that the server uses to listen and serve. If none is defined the default is 8080
	flag.IntVar(&portNumber, "port", 8080, "Port on which the server will listen and serve.")
	flag.Parse()

	// Open Postgres configuraiton file
	file, err := os.Open(fileName)
	// If there is an error return a suggestive error message
	if err != nil {
		sError := "There was a problem opening file " + fileName + " "
		var ErrReadFile = errors.New(sError)
		cErr := errors.New(ErrReadFile.Error() + err.Error())
		var d = sqlDBTx{}
		return d, portNumber, cErr
	}
	// Defering the file closure to make sure it will eventually be closed
	defer file.Close()
	// declaring a sqlDBTx stuct that will hold the Postgres connection info
	var configStruct sqlDBTx

	// Read file and split each line on the " : " separator and then split the string to the right of
	// the separtor by another spearator (",") and keep the string to the left of separator
	cSlice := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		item := scanner.Text()
		s := strings.Split(item, " : ")
		v := strings.Split(s[1], ",")
		cSlice = append(cSlice, v[0])
	}
	// In case of error return the message to the outer function
	if err := scanner.Err(); err != nil {
		var d = sqlDBTx{}
		return d, portNumber, err
	}
	// Assign the gathered values to the configStruct struct of type sqlDBTx
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
	// Return the sqlDBTx struct that holds the Postgres db configuration parameters and the Listen and Serve port number
	return configStruct, portNumber, nil
}

// GetTable is a sqlDBTx type method and its purpose is to fetch the information contained in one of the 2 tables
// of the DB (one that keeps track of transfers and one that keeps track of the information in the wallet accounts)
// GetTable is also one of core functionalities of the Wallet service and has its own go-kit endpoint
func (s sqlDBTx) GetTable(t string) ([]string, error) {
	// Based on the information contained on a sqlDBTx struct created with the "NewService" function a DB connection string is defined and a connection is opened
	connectionString := "host=" + s.sqlHost + " port=" + s.sqlPort + " user=" + s.sqlUser + " password=" + s.sqlPassword + " dbname=" + s.sqlDbName + " sslmode=" + s.sslmode
	db, err := sqlx.Open(s.sqlDriver, connectionString)
	// If any error, return it to parent function
	if err != nil {
		return nil, err
	}
	// Make sure we actually close the connction once we're done
	defer db.Close()

	// While the transaction we are about to execute is not commited we will retry it until succesful
	var isCommitted = false
	var results []string
	for ok := true; ok; ok = !isCommitted {
		// Start a transaction against the Postgres db
		// If at anypoint between the "begin" and "commit" there is any kind of issue all changes to the db will be reverted
		tx, err := db.Begin()
		// If we get an error return a descriptive message and roll back the transaction in the "defer" section
		if err != nil {
			var ErrStartTx = errors.New("err: error begining transaction in postgres")
			cErr := errors.New(ErrStartTx.Error() + err.Error())
			return nil, cErr
		}
		defer tx.Rollback()

		// Set the transaction ISOLATION LEVEL to "Serializable" to allow for multiple instances of the server to run transactions against the same Postgres db
		_, err = tx.Exec(`set transaction isolation level serializable`)
		//_, err = tx.Exec(`set transaction isolation level repeatable read`) // <=== SET ISOLATION LEVEL
		// If an error occurs return it to the parent function
		if err != nil {
			return nil, err
		}

		// Set a table lock so we exclude any type of conflicts that could generate data corruption
		_, err = tx.Exec("LOCK TABLE Accounts IN SHARE ROW EXCLUSIVE MODE;") // <=== Lock table
		// If an error occurs we retry the transaction
		if err != nil {
			log.Println(err, "...continuing...")
			continue
		}

		var txString string
		if t == s.accountsTable {
			// If we are trying to access the table that keeps information about accounts run the following query
			txString = "SELECT * FROM " + s.accountsTable + " ORDER BY AccountID;"
		} else {
			// If, instead we are trying to access the table that keeps information about fund transfers run the following query
			txString = "SELECT * FROM " + s.transfersTable + ";"
		}
		// Get the query result
		rows, err := tx.Query(txString)
		// If there is an error when sending the query
		if err != nil {
			// If the error message is indicative of a db collision retry the transaction in a new iteration
			if strings.Contains(err.Error(), "could not serialize access due to") {
				continue
			}
			// If the error is that the table has now rows
			if err == sql.ErrNoRows {
				// And if the table has information about accounts
				if t == s.accountsTable {
					// Return an apropriate error to the outer function
					var ErrAcc = errors.New("err: there are no defined accounts")
					cErr := errors.New(ErrAcc.Error() + err.Error())
					return nil, cErr
				}
				// Or, if the table has information about fund transfers return an apropriate error to the outer function
				var ErrAcc = errors.New("err: there are no defined transfers")
				cErr := errors.New(ErrAcc.Error() + err.Error())
				return nil, cErr
			}
			// If we got a different error return
			var ErrUnexp = errors.New("err: Unexpected error occured")
			cErr := errors.New(ErrUnexp.Error() + err.Error())
			return nil, cErr
		}
		// For each row returned in the query results
		for rows.Next() {
			if t == s.accountsTable {
				// If the table has information about accounts get the account ID, balance, currency and the initial balance in a slice of strings
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
				// If the table has information about func transfers get the payment ID, source account, destination account, currency and timestamp in a slice of strings
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
		// If we've gotten this far without any errors we can commit our transaction, break out of the transaction loop as the transaction was successful
		// and return the slice with all the information parsed from the query's result and a nil error to the parent function
		tx.Commit()
		isCommitted = true
	}
	return results, nil
}

// DoTransfer is a sqlDBTx type method that is responsible for the actual fund transfer transaction from one account to another
// DoTransfer takes in 3 arguments: the source account, the destination account and the transfered amount and returns a confirmation string and an empty error
// GetTable is also one of core functionalities of the Wallet service and has its own go-kit endpoint
func (s sqlDBTx) DoTransfer(fromAccount string, toAccount string, transferAmount string) (string, error) {
	// Based on the information contained on a sqlDBTx struct created with the "NewService" function a DB connection string is defined and a connection is opened
	connectionString := "host=" + s.sqlHost + " port=" + s.sqlPort + " user=" + s.sqlUser + " password=" + s.sqlPassword + " dbname=" + s.sqlDbName + " sslmode=" + s.sslmode
	db, err := sqlx.Open(s.sqlDriver, connectionString)
	// If any error, return it to parent function
	if err != nil {
		log.Println("err", err)
		return "error", err
	}

	// check if the source account and destination account are the same and return an error before any transactions happen as we do not support transactions of this type
	if fromAccount == toAccount {
		var ErrSameAcc = errors.New("the source account is the same as the destination account. ")
		//log.Println("err", ErrSameAcc)
		return "error", ErrSameAcc
	}
	// Make sure we actually close the connction once we're done
	defer db.Close()

	// While the transaction we are about to execute is not commited we will retry it until succesful
	var isCommitted = false
	for ok := true; ok; ok = !isCommitted {
		// Start a transaction against the Postgres db
		tx, err := db.Begin()
		// If we get an error return a descriptive message and roll back the transaction in the "defer" section, then restart transaction
		// If at anypoint between the "begin" and "commit" there is an issue all changes to the db will be reverted
		if err != nil {
			var ErrStartTx = errors.New("err: error begining transaction in postgres")
			cErr := errors.New(ErrStartTx.Error() + err.Error())
			return "error", cErr
		}
		defer tx.Rollback()

		// Set the transaction ISOLATION LEVEL to "Serializable" to allow for multiple instances of the server to run transactions against the same Postgres db
		_, err = tx.Exec(`set transaction isolation level serializable`)
		//_, err = tx.Exec(`set transaction isolation level repeatable read`) // <=== SET ISOLATION LEVEL
		// If an error occurs return it to the parent function and restart the transaction
		if err != nil {
			return "error", err
		}

		// Set a table lock so we exclude any type of conflicts that could generate data corruption
		_, err = tx.Exec("LOCK TABLE Accounts IN SHARE ROW EXCLUSIVE MODE;") // <=== Lock table
		// If an error occurs we retry the transaction
		if err != nil {
			log.Println(err, "...continuing...")
			continue
		}

		// Fetch the balance and source account currency
		var sBalance string
		var sCurrency string
		txString := "SELECT Balance , Currency FROM " + s.accountsTable + " WHERE AccountID ='" + fromAccount + "';"
		err = tx.QueryRow(txString).Scan(&sBalance, &sCurrency)
		// Return error messages if the query finds that the indicated source account does not return any results
		if err != nil {
			if err == sql.ErrNoRows {
				var ErrNoSource = errors.New("The source account does not exist")
				return "error", ErrNoSource
			}
			// Otherwise return a relevant error message
			var ErrUnexpect = errors.New("err: unexpected error")
			cErr := errors.New(ErrUnexpect.Error() + err.Error())
			return "error", cErr
		}
		// If there is an issue with reading the balance return an appropriate error
		fBalance, err := strconv.ParseFloat(sBalance, 64)
		if err != nil {
			var ErrParse = errors.New("Error parsing blance")
			return "error", ErrParse
		}

		// If there is an issue with determining the transfered amout return an appropriate error
		fAmount, err := strconv.ParseFloat(transferAmount, 64)
		if err != nil {
			var ErrParse = errors.New("err: error begining transaction in postgres")
			cErr := errors.New(ErrParse.Error() + err.Error())
			return "error", cErr
		}
		// If the balance is insuficcient to allow the indicated amount transfer return an appropriate message
		if fBalance < fAmount {
			var ErrBalance = errors.New("Balance insuficient for transaction")
			return "error", ErrBalance
		}
		// Fetch currency of the destination account
		var dCurrency string
		txString = "SELECT Currency FROM " + s.accountsTable + " WHERE AccountID ='" + toAccount + "';"
		err = tx.QueryRow(txString).Scan(&dCurrency)
		// if there is an error while fetching the currency retun an appropriate error
		if err != nil {
			if err == sql.ErrNoRows {
				var ErrNoSource = errors.New("The destination account does not exist")
				return "error", ErrNoSource
			}
			return "error", err
		}

		// If the source account currency is not the same as the destination account currency, then the transfer is not allowed
		if dCurrency != sCurrency {
			var ErrMissmatch = errors.New("Not same currency in transaction source and destination")
			return "error", ErrMissmatch
		}

		// Make query to implement in the Account table the substraction of the transfer amount from the source account
		txString = "UPDATE " + s.accountsTable + " SET balance = balance - " + transferAmount + " WHERE accountid = '" + fromAccount + "';"
		_, err = tx.Exec(txString)
		// In case of failures return apropriate error messages
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
		// Make query to implement in the Account table the addition of the transfer amount to the destination account
		txString = "UPDATE " + s.accountsTable + " SET balance = balance + " + transferAmount + " WHERE accountid= '" + toAccount + "';"
		_, err = tx.Exec(txString)
		// In case of failures if the error message is indicative of a db collision retry the transaction in a new iteration
		if err != nil {
			if strings.Contains(err.Error(), "could not serialize access due to") {
				continue
			}
			// otherwise return the error message to the outer function
			return "error", err
		}
		t0 := time.Now().Format(time.RFC3339)
		// Insert into the table responsibile for tracking transactions the information about this particular transfer:
		// Transaction ID, Source account, Destination Account, Amount transfered, Currency of amount transfered and Timestamp of transaction
		txString = "INSERT INTO " + s.transfersTable + " (transid, From_Account, To_Account, Amount, Currency, TTime) VALUES( nextval('Payment_counter'), '" + fromAccount + "', '" + toAccount + "', '" + transferAmount + "', '" + sCurrency + "', '" + t0 + "' );"
		_, err = tx.Exec(txString)

		if err != nil {
			// In case of a db write conflict retry the transaction in a new iteration
			if strings.Contains(err.Error(), "could not serialize access due to") {
				continue
			} else {
				// otherwise return the error to the outer function
				return "error", err
			}
		}
		// If we've gotten this far without any errors we can commit our transaction, break out of the transaction loop as the transaction was successful
		// and return an apropriate status mesage with a nil error
		tx.Commit()
		isCommitted = true
	}
	return "success", nil
}
