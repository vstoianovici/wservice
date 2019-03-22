# wService

`Wallet Service or wService for short` provides a generic basic Wallet service with a RESTful API implemented in Go (using the go-kit library/toolkit) the employs Postgres as a db solution.

Here are a few basic functionalities that are covered:

- Sending a payment from one account to another (same currency)
- Seeing all comitted payments since the initial balance value
- Seeing all available accounts

Assumptions and requirements:

- Only payments within the same currency are supported (no exchanges)
- There are no users in the system (no auth)
- Balance can't go below zero
- More than one instance of the application can be launched


## Get started with docker


Get the source code:

```
$ go get -u github.com/vstoianovici/wStart
```

Build the environment from the `docker-compose.yml` file in the root (`gowebapp` and `postgresdb` will be deployed):

```
$ docker-compose up -d

```
The `postgresdb` will already have a database called `Postgres` that has the `Accounts` and `Transfers` tables. The `Account` table will look something like this:

poza

while the `Transfers` table will be empty:

poza

One can visualize both tables by accessing the follwing links:

- The `Account` table: http://127.0.0.1:8080/accounts

- The `Transfers` table http://127.0.0.1:8080/transfers

- Additionally one can see the `Metrics & Instrumentation` endpoint here: http://127.0.0.1:8080/metrics


## Get started with building the Go binary and deploying a postgres DB

Get the source code:

```
$ go get -u github.com/vstoianovici/wStart
```

Build the binary by running the following command in the root:

```
$ make build

```

If the build succeeds, the resulting binary named "wService" should be found in the `/cmd` directory.

To build the Postgres db as a Docker container run these 2 commands:

```
docker build -t postgresdb -f ./Dockerfile_postgres .
```
followed by

```
docker run --rm --name postgresdb -e POSTGRES_PASSWORD=password -d -p 5432:5432 postgresdb
```
In case Postgres is installed in any other way other than the ones described above the user needs to create a database named `Postgres` and run the SQL queries contained in the init-user-db.sh script from the `/docker` folder.


Addtionally there is a `postgresql.cfg` file contained in `/cmd` that is used for configuring the connection between the Go webapp and the Postgres db. The content is pretty self-explanatory:

```yaml
sqlDriver : postgres,
sqlHost : 127.0.0.1,
sqlPort : 5432,
sqlUser : postgres,
sqlPassword : password,
sqlDbName : postgres,
sslmode : disable,
accountsTable : Accounts,
transfersTable : Transfers
```yaml

Run the tests:

```
$ make test
```

Feel free to explore the `Makefile` available in the root directory.

### Runtime

If you went the docker-compose route you can just skip to the below curl commands.

Otherwise, once `wService` is built and ready for runtime it can run (/cmd/wService) without any parameters (default should be fine) but there is the option of passing in a different port or a different `postgres.cfg` file:

```
$ ./wService -h
time=2019-03-22T20:40:18.099917Z tag=start msg="created logger"
Usage of ./wService:
  -file string
        Path of postgresql config file to be parsed. (default "./postgresql.cfg")
  -port int
        Port on which the server will listen and serve. (default 8080)
```

At runtime the easiest way to actually create fund transfers is to run a curl command against the `submittransfer` API endpoint such as the following:

```
curl  -d'{"from":"bob123","to":"alice456","amount":"20"}' "127.0.0.1:8080/submittransfer"
```

The other touchpoints of the API are, as mentioned earlier `/accounts`, `/transfers` and `/metrics` reachable with:

curl "127.0.0.1:8080/transfers"

curl "127.0.0.1:8080/accounts"

curl "127.0.0.1:8080/metics"


### Build your own wallet

Anybody can use this resource as a library to create their own implementation of a micro Wallet Service as long as they mimic what is being done in `/cmd/main.go`

For the future, a nice feature to implement would be a gRPC endpoint in addition to the Json over HTTP REST API so that the wallet can be a service in a microservice architecture solution.

