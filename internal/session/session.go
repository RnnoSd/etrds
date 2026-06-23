package session

import (
	"database/sql"
	"fmt"
)

type Session interface {
	Scheduler() error
}

type etrdsSession struct {
	Name     string
	Consults map[string]Consult
}

type State int

const (
	registered State = iota + 1
	staged
	running
	fetched
	stored
	done
)

type Consult struct {
	Name              string
	Query             string
	LocalDBConnection *sql.DB
	FetchDBType       string
	FetchDBConnection *sql.DB
	State             State
}

func NewETRDSSession(sessionName string, initialConsults map[string]Consult) (etrdsSession, error) {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return etrdsSession{}, fmt.Errorf("error al abrir duckdb en memoria: %w", err)
	}

	initializedConsults := make(map[string]Consult)

	for consultName, consult := range initialConsults {
		attachStatement := fmt.Sprintf(`ATTACH %s connection_mode='READ_ONLY' AS %s`, consult.LocalDBLocation, consultName)
		_, attachmentErr := db.Exec(attachStatement)
		if attachmentErr != nil {
			err = fmt.Errorf("%v\ndb (%s) did not attach to the session %s since %v", err, consult.LocalDBLocation, sessionName, attachmentErr)
		} else {
			initializedConsults[consultName] = consult
		}
	}

	return etrdsSession{
		Name:     sessionName,
		Consults: initializedConsults,
	}, err
}

func GetSession() Session {

}
