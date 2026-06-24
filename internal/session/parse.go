package session

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/apache/arrow-go/v18/parquet/variant"
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

func (sess *aliveSession) ExtractFetchQUERIES(name string) ([]string, error) {
	var fetchQueries []string
	strToParse := string(sess.Consults[name].Content)
	defer sess.Consults[name].Queued()
	queries := strings.Split(strToParse, ";")

	for _, query := range queries {
		var fetchQuery string

		dbType := sess.Fetching[name].Type
		dbDSN := sess.Fetching[name].DSN

		deriveRegex := regexp.MustCompile(`^s*#\[derive\(fetch\.([a-zA-Z]+)\)\]`)

		if deriveExps := deriveRegex.FindAllStringSubmatch(query, -1); len(deriveExps) > 0 {
			fetchedStr := deriveExps[0][1]

			switch fetchedStr {
			case "LightFetch":
				var ParseErr error
				fetchQuery, ParseErr = ParseSELECTlightFetch(deriveRegex.ReplaceAllLiteralString(query, ""), dbType, dbDSN)
				if ParseErr != nil {
					return []string{""}, ParseErr
				}
			case "FullFetch":
				fetchQuery = deriveRegex.ReplaceAllLiteralString(query, "")
			default:
				err := fmt.Errorf("not valid fetching workflow option")
				return []string{""}, err
			}
		} else {
			fetchQuery = ""
		}
		if !slices.Conatins(fetchQueries, fetchQuery) {
			fetchQueries = append(fetchQueries, fetchQuery)
		}
	}

	return fetchQueries, nil
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

func ParseSELECTlightFetch(q string, dbType string, dsn string) (string, error) {
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
