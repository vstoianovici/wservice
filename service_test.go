package wservice

import (
	"os"
	"sync"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestNewServiceDefault(t *testing.T) {
	fileName := "./cmd/postgresql.cfg"
	svc, err := getDbConfig(fileName)
	assert.Nil(t, err)
	assert.NotNil(t, svc)
}

func TestNewServiceNoFile(t *testing.T) {
	fileName := "./postgresql.cfg_"
	// Open Postgres configuration file
	file, err := os.Open(fileName)
	assert.NotNil(t, err)
	defer file.Close()
}

func TestNewServiceWrongFormatDelimeter(t *testing.T) {
	fileName := "./cmd/test/postgresql_delim.cfg"
	_, err := getDbConfig(fileName)
	assert.NotNil(t, err)
}

func TestNewServiceWrongFormatFewLines(t *testing.T) {
	fileName := "./cmd/test/postgresql_line_few.cfg"
	_, err := getDbConfig(fileName)
	assert.NotNil(t, err)
}

func TestNewServiceWrongFormatManyLines(t *testing.T) {
	fileName := "./cmd/test/postgresql_line_many.cfg"
	_, err := getDbConfig(fileName)
	assert.NotNil(t, err)
}
func TestGetTable(t *testing.T) {
	svc, _ := getDbConfig("./cmd/postgresql.cfg")
	vSlice, err := svc.GetTable("Accounts")
	assert.Contains(t, vSlice, "Success.")
	assert.Nil(t, err)
	vSlice, err = svc.GetTable("Transfers")
	assert.Contains(t, vSlice, "Success.")
	assert.Nil(t, err)
	vSlice, err = svc.GetTable("SomeOtherTable")
	assert.NotContains(t, vSlice, "[]")
}

func TestDoTransferRegular(t *testing.T) {
	svc, _ := getDbConfig("./cmd/postgresql.cfg")
	status, err := svc.DoTransfer("bob123", "alice456", "30")
	assert.Contains(t, status, "success")
	assert.Nil(t, err)
	status, err = svc.DoTransfer("alice456", "bob123", "30")
	assert.Contains(t, status, "success")
	assert.Nil(t, err)
}

func TestDoTransferNoDestAccount(t *testing.T) {
	svc, _ := getDbConfig("./cmd/postgresql.cfg")
	status, err := svc.DoTransfer("alice456", "amockaccount123", "30")
	assert.Contains(t, status, "error")
	assert.EqualError(t, err, "The destination account does not exist")
}

func TestDoTransferNoSourceAccount(t *testing.T) {
	svc, _ := getDbConfig("./cmd/postgresql.cfg")
	status, err := svc.DoTransfer("amockaccount123", "alice456", "30")
	assert.Contains(t, status, "error")
	assert.EqualError(t, err, "The source account does not exist")
}

func TestDoTransferSameAccount(t *testing.T) {
	svc, _ := getDbConfig("./cmd/postgresql.cfg")
	status, err := svc.DoTransfer("alice456", "alice456", "30")
	assert.Contains(t, status, "error")
	assert.EqualError(t, err, "the source account is the same as the destination account. ")
}

func TestDoTransferBalanceMinus(t *testing.T) {
	svc, _ := getDbConfig("./cmd/postgresql.cfg")
	status, err := svc.DoTransfer("alice456", "bob123", "300000")
	assert.Contains(t, status, "error")
	assert.EqualError(t, err, "Balance insuficient for transaction")
}

func TestDoTransferWrongCurrency(t *testing.T) {
	svc, _ := getDbConfig("./cmd/postgresql.cfg")
	status, err := svc.DoTransfer("alice456", "marcy789", "30")
	assert.Contains(t, status, "error")
	assert.EqualError(t, err, "Not same currency in transaction source and destination")
}

func TestDoTransferConcurent(t *testing.T) {
	var wg sync.WaitGroup
	svc, _ := getDbConfig("./cmd/postgresql.cfg")
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 5; j++ {
				status, err := svc.DoTransfer("bob123", "alice456", "1")
				//log.Println("status", status, " and ", j)
				assert.Contains(t, status, "success")
				assert.Nil(t, err)
			}
			wg.Done()
		}()
	}
	wg.Wait()

}

func TestDoTransferConcurentRevertChanges(t *testing.T) {
	var wg sync.WaitGroup
	svc, _ := getDbConfig("./cmd/postgresql.cfg")
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 5; j++ {
				status, err := svc.DoTransfer("alice456", "bob123", "1")
				//log.Println("status", status, " and ", j)
				assert.Contains(t, status, "success")
				assert.Nil(t, err)
			}
			wg.Done()
		}()
	}
	wg.Wait()

}
