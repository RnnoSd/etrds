package session

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v3"
)

func (sess *AliveSession) ExtractFetchQUERIES(name string) ([]string, error) {
	var fetchQueries []string
	strToParse := string(sess.Consults[name].Content)
	defer sess.Consults[name].Queued()
	queries := strings.Split(strToParse, ";")

	for _, query := range queries {
		var fetchQuery string

		dbType := sess.Fetching[name].Type
		dbDSN := sess.Fetching[name].DSN

		deriveRegex := regexp.MustCompile(`^\s*#\[derive\(fetch\.([a-zA-Z]+)\)\]`)

		if deriveExps := deriveRegex.FindAllStringSubmatch(query, -1); len(deriveExps) > 0 {
			fetchedStr := deriveExps[0][1]

			switch fetchedStr {
			case "LightFetch":
				var ParseErr error
				fetchQuery, ParseErr = ParseSELECTlightFetch(deriveRegex.ReplaceAllLiteralString(query, ""), dbType, dbDSN, name)
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
		if !slices.Contains(fetchQueries, fetchQuery) {
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
	GetTableName() string
}

func (q QueryExplainMySQL) GetUsedColumns() []string {
	return q.QueryBlock.Table.UsedColumns
}

func (q QueryExplainMySQL) GetAttachedConditions() string {
	return q.QueryBlock.Table.AttachedCondition
}

func (q QueryExplainMySQL) GetTableName() string {
	return q.QueryBlock.Table.TableName
}

func unmarshalAndBuild[T QueryExplain](jsonRaw []byte, catalog string) (string, error) {
	var target T
	if err := json.Unmarshal(jsonRaw, &target); err != nil {
		return "", err
	}

	tableName := target.GetTableName()
	columns := strings.Join(target.GetUsedColumns(), ", ")
	condition := target.GetAttachedConditions()

	// Qualify the table with the ATTACH catalog so the query reads through the
	// attached source on the local read_write connection (no trailing `;`, the
	// result is embedded into a CREATE TABLE AS ...).
	from := tableName
	if catalog != "" {
		from = catalog + "." + tableName
	}

	if condition == "" {
		return fmt.Sprintf("SELECT %s FROM %s", columns, from), nil
	}
	return fmt.Sprintf("SELECT %s FROM %s WHERE %s", columns, from, condition), nil
}

type NameFormat struct {
	Formating        string           // template fed verbatim into fmt.Sprintf
	RegexpsCompilers []*regexp.Regexp // one regexp per verb in Formating, in order
}

var headerRegex = regexp.MustCompile(`(?s)\A\s*---\s*\n(.*?)\n\s*---\s*(?:\n|$)`)

func ExtractFormat(q string) (NameFormat, error) {
	m := headerRegex.FindStringSubmatch(q)
	if m == nil {
		return NameFormat{}, fmt.Errorf("no format header (--- ... ---) found")
	}

	var rawMap map[string]any
	if err := yaml.Unmarshal([]byte(m[1]), &rawMap); err != nil {
		return NameFormat{}, fmt.Errorf("parsing format header: %w", err)
	}

	// normalize keys to lower-case so field names are case-insensitive
	fields := make(map[string]any, len(rawMap))
	for k, v := range rawMap {
		fields[strings.ToLower(k)] = v
	}

	formating, ok := fields["formating"].(string)
	if !ok {
		return NameFormat{}, fmt.Errorf("format header missing 'formating' string")
	}
	nf := NameFormat{Formating: formating}

	rawRegexps, _ := fields["regexps"].([]any)
	for _, r := range rawRegexps {
		pattern, ok := r.(string)
		if !ok {
			return NameFormat{}, fmt.Errorf("'regexps' entries must be strings, got %T", r)
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return NameFormat{}, fmt.Errorf("compiling regexp %q: %w", pattern, err)
		}
		nf.RegexpsCompilers = append(nf.RegexpsCompilers, re)
	}

	return nf, nil
}

func ExtractNamedQuery(q string, fmtQ NameFormat) (string, error) {
	body := headerRegex.ReplaceAllString(q, "")

	values := make([]any, 0, len(fmtQ.RegexpsCompilers))
	for _, re := range fmtQ.RegexpsCompilers {
		match := re.FindStringSubmatch(body)
		if match == nil {
			return "", fmt.Errorf("regexp %q matched nothing in query", re.String())
		}
		value := match[0]
		if len(match) > 1 {
			value = match[1]
		}
		values = append(values, value)
	}

	named := fmt.Sprintf(fmtQ.Formating, values...)

	if strings.Contains(named, "%!") {
		return "", fmt.Errorf("format %q does not match its %d regexp arguments: %s", fmtQ.Formating, len(values), named)
	}
	return named, nil
}

func ParseSELECTlightFetch(q string, dbType string, dsn string, catalog string) (string, error) {
	var statementSELECTWHERE string
	var jsonRaw []byte

	switch dbType {
	case MySQL:
		dbConnection, err := sql.Open("mysql", dsn)
		if err != nil {
			return "", err
		}
		defer dbConnection.Close()
		err = dbConnection.QueryRow(fmt.Sprintf("EXPLAIN FORMAT=JSON %s", q)).Scan(&jsonRaw)
		if err != nil {
			return "", err
		}
		statementSELECTWHERE, err = unmarshalAndBuild[QueryExplainMySQL](jsonRaw, catalog)
		if err != nil {
			return "", err
		}

	default:
		err := fmt.Errorf(`%s do not exist, try %v`, dbType, DBsupported)
		return "", err
	}

	return statementSELECTWHERE, nil
}
