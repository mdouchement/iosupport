# TSV Sorter

## Compilation

```sh
$ go get ./...
$ go build -o tsvsorter *.go
```

## Usage

```sh
$ ./tsvsorter -h
```

## Usage examples

- From LocalFS to HDFS

```sh
./tsvsorter -i iris.csv -s=',' -H -f "Sepal Length" -o hdfs://localhost:9000/S_iris.csv
```

- From HDFS to HDFS

```sh
./tsvsorter -i hdfs://localhost:9000/iris.csv -s=',' -H -f "Sepal Length" -o hdfs://localhost:9000/S_iris.csv
```

- From HDFS to LocalFS

```sh
./tsvsorter -i hdfs://localhost:9000/iris.csv -s=',' -H -f "Sepal Length" -o S_iris.csv
```
