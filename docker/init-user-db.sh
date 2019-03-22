#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	CREATE TABLE Accounts (
    AccountID varchar(255) PRIMARY KEY,
    Balance decimal(9,3) NOT NULL CHECK (Balance>=0),
    Currency varchar(255) NOT NULL,
	InitialBalance decimal(9,3) NOT NULL CHECK (Balance>=0)
	);

	CREATE TABLE Transfers (
	    TransID int NOT NULL PRIMARY KEY,
	    From_Account varchar(255) NOT NULL,
		To_Account varchar(255)  NOT NULL,
		Amount decimal(9,3) NOT NULL CHECK (Amount>=0),
		Currency varchar(255) NOT NULL,
		TTime varchar(255) NOT NULL,
		FOREIGN KEY (From_Account) REFERENCES Accounts(AccountID),
		FOREIGN KEY (To_Account) REFERENCES Accounts(AccountID)
	);


	INSERT INTO Accounts (AccountID, Balance, Currency, InitialBalance)
	VALUES ('bob123', '302.35', 'USD', '302.35');

	INSERT INTO Accounts (AccountID, Balance, Currency, InitialBalance)
	VALUES ('alice456', '573.81', 'USD', '573.81');

	INSERT INTO Accounts (AccountID, Balance, Currency, InitialBalance)
	VALUES ('marcy789', '4583.90', 'EUR', '4583.90');

	INSERT INTO Accounts (AccountID, Balance, Currency, InitialBalance)
	VALUES ('lucy0123', '14583.90', 'EUR', '14583.90');

	CREATE SEQUENCE Payment_Counter;
EOSQL