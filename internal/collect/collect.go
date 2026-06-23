/*
Package collect [LLM-manifest]
*/
package collect

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/RnnoSd/etrds/internal/config"
	"github.com/RnnoSd/etrds/internal/session"
	"github.com/spf13/cobra"
)

func Run(cmd *cobra.Command, args []string) {
	s := session.etrdsSession()
	endRegexSQL := regexp.MustCompile(`.\.sql$`)

	for _, arg := range args {
		if sqlPattern := endRegexSQL.FindAllStringSubmatch(arg, -1); len(sqlPattern) > 0 {
			sqlQueries, err := os.ReadFile(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "read file error %s: %v\n", arg, err)
			}

			Queries := strings.Split(string(sqlQueries), ";")

			for _, query := range Queries {
				consult := session.Consult{
					Name:              endRegexSQL.ReplaceAllLiteralString(arg, ""),
					Query:             query,
					LocalDBConnection: config.GetDBConnection(),
					FetchDBType:       config.GetFetchDBType(),
					FetchDBConnection: config.GetFetchDBType(),
					State:             config.GetFetchDBType(),
				}

				fetchQuery, err := consult.ExtractFetchQUERY()
				if err != nil {
					fmt.Fprintf(os.Stderr, "parsing Fetch query %s: %v\n", arg, err)
				}
				go func() {
					sqlRows, err := consult.FetchDBConnection.Query(fetchQuery)
					if err != nil {
						fmt.Fprintf(os.Stderr, "fetch information error %s: %v\n", arg, err)
					}
					tableLocation := fmt.Sprintf(`backups.%s`, consult.Name)
					err := uploadLocalDBLocation(sqlRows, consult.LocalDBConnection, tableLocation)
					if err != nil {
						fmt.Fprintf(os.Stderr, "upload information local Database %s: %v\n", arg, err)
					}
				}()
			}
		} else {
			consult := s.Consults[arg]
			fetchQuery, err := consult.ExtractFetchQUERY()
			if err != nil {
				fmt.Fprintf(os.Stderr, "fetch information error %s: %v\n", arg, err)
			}
			go func() {
				sqlRows, err := consult.FetchDBConnection.Query(fetchQuery)
				if err != nil {
					fmt.Fprintf(os.Stderr, "fetch information error %s: %v\n", arg, err)
				}
				uploadLocalDBLocation(sqlRows)
			}()
		}
	}
}

func uploadLocalDBLocation(sqlRows *sql.Rows, db *sql.DB, tableDestination string) error {
	columnas, err := sqlRows.Columns()
	if err != nil {
		return fmt.Errorf("error al obtener columnas de origen: %w", err)
	}
	numColumnas := len(columnas)
	placeholders := make([]string, numColumnas)
	for i := range placeholders {
		placeholders[i] = "?"
	}

	insertQuery := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableDestination,
		strings.Join(columnas, ", "),
		strings.Join(placeholders, ", "),
	)

	stmt, err := db.Prepare(insertQuery)
	if err != nil {
		return fmt.Errorf("error al preparar inserción en destino: %w", err)
	}
	defer stmt.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txStmt := tx.Stmt(stmt)

	valores := make([]interface{}, numColumnas)
	punterosValores := make([]interface{}, numColumnas)
	for i := range valores {
		punterosValores[i] = &valores[i]
	}

	for sqlRows.Next() {
		if err := sqlRows.Scan(punterosValores...); err != nil {
			return fmt.Errorf("error escaneando fila de origen: %w", err)
		}

		if _, err := txStmt.Exec(valores...); err != nil {
			return fmt.Errorf("error insertando fila en destino: %w", err)
		}
	}

	if err := sqlRows.Err(); err != nil {
		return fmt.Errorf("error en el lector de filas: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error al hacer commit en el destino: %w", err)
	}

	return nil
}
