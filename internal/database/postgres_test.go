package database

import (
	"testing"
)

func TestPostgres(t *testing.T) {
	postgresdb := InitTest("postgres://olakunle:habeeb1234@localhost:5432/prism?sslmode=disable")

	var count int

	postgresdb.QueryRow(`
		SELECT COUNT(*) 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_TYPE = 'BASE TABLE' 
			AND TABLE_SCHEMA = 'public'
	`).Scan(&count)

	
	want := 3
	
	if (count != want) {
		t.Fatal("error")
	}
}