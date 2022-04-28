package storage

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/denisenkom/go-mssqldb"
)

type MSSQL struct {
	connString  string
	sqlUsername string
	sqlPassword string
	sqlHost     string
	sqlInstance string
	sqlDatabase string
	database    *sql.DB
}

func NewMSSQL(sqlUsername, sqlPassword, sqlHost, sqlInstance, sqlDatabase string) (*MSSQL, error) {
	if sqlUsername == "" {
		return nil, fmt.Errorf("mssql: username cannot be empty")
	}

	if sqlPassword == "" {
		return nil, fmt.Errorf("mssql: password cannot be empty")
	}

	if sqlHost == "" {
		return nil, fmt.Errorf("mssql: host cannot be empty")
	}

	if sqlDatabase == "" {
		return nil, fmt.Errorf("mssql: database cannot be empty")
	}

	query := url.Values{}
	query.Add("database", sqlDatabase)

	u := url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(sqlUsername, sqlPassword),
		Host:     sqlHost,
		RawQuery: query.Encode(),
	}

	if sqlInstance != "" {
		u.Path = sqlInstance
	}

	return &MSSQL{
		connString:  u.String(),
		sqlUsername: sqlUsername,
		sqlPassword: sqlPassword,
		sqlHost:     sqlHost,
		sqlInstance: sqlInstance,
		sqlDatabase: sqlDatabase,
	}, nil
}

func (s *MSSQL) Bootstrap(key string) error {
	if s.database == nil {
		var dberr error
		s.database, dberr = sql.Open("sqlserver", s.connString)
		if dberr != nil {
			return fmt.Errorf("unable to open SQLServer DB: %w", dberr)
		}
	}

	if err := s.database.Ping(); err != nil {
		return fmt.Errorf("unable to ping SQLServer DB: %w", err)
	}

	table := `if not exists (select * from sysobjects where name = '{tableName}' and xtype = 'U')
	create table {tableName} (
		{tableName} varchar(max) not null
	);

	insert into {tableName} values ('');
`

	createTableSQL := strings.NewReplacer("{tableName}", key).Replace(table)

	if _, err := s.database.Exec(createTableSQL); err != nil {
		return fmt.Errorf("unable to bootstrap records table: %w", err)
	}

	return nil
}

func (s *MSSQL) Get(key string) (string, error) {
	selectSQL := strings.NewReplacer("{tableName}", key).Replace(`select {tableName} from {tableName}`)
	rows, err := s.database.Query(selectSQL)
	if err != nil {
		return "", fmt.Errorf("unable to query database: %w", err)
	}

	defer rows.Close()

	if rows.Next() {
		var value string

		if err := rows.Scan(&value); err != nil {
			return "", fmt.Errorf("unable to scan database results: %w", err)
		}

		return value, nil
	}

	return "", nil
}

func (s *MSSQL) Set(key, value string) error {
	insertSQL := strings.NewReplacer("{tableName}", key).Replace(`update {tableName} set {tableName} = @VALUE`)
	_, err := s.database.Exec(insertSQL, sql.Named("VALUE", value))
	return err
}

func (r *MSSQL) IsValidKey(key string) error {
	if key != strings.ToLower(key) {
		return fmt.Errorf("key %q must be only alphabetic characters, all lowercase", key)
	}

	if !alphaOnly(key) {
		return fmt.Errorf("key %q must be only alphabetic characters, all lowercase", key)
	}

	return nil
}

func (r *MSSQL) ConfigString() string {
	query := url.Values{}
	query.Add("database", r.sqlDatabase)

	u := url.URL{
		Scheme:   "sqlserver",
		Host:     r.sqlHost,
		RawQuery: query.Encode(),
	}

	if r.sqlInstance != "" {
		u.Path = r.sqlInstance
	}

	return fmt.Sprintf(
		"SQL Server: %s; user: %s; password: %s",
		u.String(),
		obfuscate(r.sqlUsername, 3),
		obfuscate(r.sqlPassword, 3),
	)
}
