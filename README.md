# CYBERCRYPT D1 GORM

This integration can be used to encrypt and decrypt data transparently when reading and writing to the database, using [CYBERCRYPT D1 Generic](https://github.com/cybercryptio/d1-service-generic/). The data is encrypted in the application layer in such a way that the database itself never receives the data in plain text.

This protects the data in the database from being read by third parties and tampering.

## Supported databases

All databases supported by GORM are supported by the d1gorm package. These include:

- MySQL
- PostgreSQL
- SQL Server
- SQLite
- any database that is compatible with the `mysql` or `postgres` dialects.

## Installation

You can install the d1gorm package in your go project by running:
```
go get github.com/cybercryptio/d1-gorm
```

## Usage

For examples of how to use the integration [see our examples in the godoc](https://pkg.go.dev/github.com/cybercryptio/d1-gorm).

## Limitations

- Currently only `string` and `[]byte` data fields can be encrypted.
- Encrypted data is not searchable by the database.

## License

The software in the CYBERCRYPT d1-gorm repository is licensed under the Apache License 2.0.
