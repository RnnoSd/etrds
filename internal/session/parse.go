package session

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type FetchOptions int

const (
	LightFetch FetchOptions = iota + 1
	FullFetch

	// Total options
	TotalOptions
)

type FetchState struct {
	Name FetchOptions
}

func (cons Consult) ExtractFetchQUERY() (string, error) {
	var fetchQuery string
	strToParse := cons.Query

	deriveRegex := regexp.MustCompile(`^#\[derive\(fetch\.([a-zA-Z]+)\)\]`)

	if deriveExps := deriveRegex.FindAllStringSubmatch(strToParse, -1); len(deriveExps) > 0 {
		fetchedStr := deriveExps[0][1]

		switch fetchedStr {
		case "LightFetch":
			var ParseErr error
			fetchQuery, ParseErr = ParseSELECTlightFetch(deriveRegex.ReplaceAllLiteralString(strToParse, ""), cons.FetchDBType, cons.FetchDBConnection)
			if ParseErr != nil {
				return "", ParseErr
			}
		case "FullFetch":
			fetchQuery = deriveRegex.ReplaceAllLiteralString(strToParse, "")
		default:
			err := fmt.Errorf("not valid fetching workflow option")
			return "", err
		}
	} else {
		fetchQuery = ""
	}

	return fetchQuery, nil
}

const (
	MySQL = "MySQL"
)

var DBsupported = []string{MySQL}

type TableMySQL struct {
	TableName         string   `json:"table_name"`
	AccessType        string   `json:"access_type"`
	UsedColumns       []string `json:"used_columns"`
	AttachedCondition string   `json:"attached_condition"`
}

type QueryBlockMySQL struct {
	SelectId int        `json:"select_id"`
	Table    TableMySQL `json:"table"`
}

type QueryExplainMySQL struct {
	QueryBlock QueryBlockMySQL `json:"query_block"`
}

type QueryExplain interface {
	GetUsedColumns() []string
	GetAttachedConditions() string
}

func (q QueryExplainMySQL) GetUsedColumns() []string {
	return q.QueryBlock.Table.UsedColumns
}
func (q QueryExplainMySQL) GetAttachedConditions() string {
	return q.QueryBlock.Table.AttachedCondition
}

func unmarshalAndBuild[T QueryExplain](jsonRaw []byte) (string, error) {
	var target T
	if err := json.Unmarshal(jsonRaw, &target); err != nil {
		return "", err
	}

	columns := strings.Join(target.GetUsedColumns(), ", ")
	condition := target.GetAttachedConditions()

	if condition == "" {
		return fmt.Sprintf("SELECT %s;", columns), nil
	}
	return fmt.Sprintf("SELECT %s WHERE %s;", columns, condition), nil
}

func ParseSELECTlightFetch(q string, dbType string, dbConnection *sql.DB) (string, error) {
	var statementSELECTWHERE string
	var jsonRaw []byte
	var err error

	switch dbType {
	case MySQL:
		err = dbConnection.QueryRow(fmt.Sprintf("EXPLAIN FORMAT=JSON %s", q)).Scan(&jsonRaw)
		if err != nil {
			return "", err
		}
		statementSELECTWHERE, err = unmarshalAndBuild[QueryExplainMySQL](jsonRaw)
		if err != nil {
			return "", err
		}

	// case PostgreSQL:
	// return "", fmt.Errorf("postgresql implementation pending")

	default:
		err := fmt.Errorf(`%s do not exist, try %v`, dbType, DBsupported)
		return "", err
	}

	return statementSELECTWHERE, nil
}
