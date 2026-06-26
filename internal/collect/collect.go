/*
Package collect [LLM-manifest]
*/
package collect

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/RnnoSd/etrds/internal/cache"
	"github.com/RnnoSd/etrds/internal/session"
	"github.com/spf13/cobra"
)

func Run(cmd *cobra.Command, args []string) {
	var (
		s   *session.AliveSession
		err error
	)
	// Here we do bring up to live a cached session or create a new-one
	// Depending a flag (LIKE git switch)
	create, _ := cmd.Flags().GetBool("create")
	createOrReset, _ := cmd.Flags().GetBool("createOrReset")
	switch {
	case create:
		cs := cache.NewCachedSession(args[0])

		s, err = session.SetUpSessionCached(cs)
		cs.Save()
	case createOrReset:
		cs := cache.NewCachedSession(args[0])
		s, err = session.SetUpSessionCached(cs)
		if err := os.Remove(cs.WritingPath); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "error al resetear la session")

			return
		}

		s, err = session.SetUpSessionCached(cs)
		cs.Save()
	default:
		cs, rerr := cache.ReadCachedSession(args[0])
		if rerr != nil {
			fmt.Fprintf(os.Stderr, "error al leer una session en el cache: %v\n", rerr)
			return
		}
		s, err = session.SetUpSessionCached(cs)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error al iniciar la session: %v\n", err)
		return
	}

	// The session owns the single read_write connection to the local store;
	// release it when we're done.
	defer s.Close()

	// Track the fetch goroutines so we can wait for every write to finish before
	// the session connection closes.
	var wg sync.WaitGroup

	for _, arg := range args[1:] {
		if consult, ok := s.Consults[arg]; ok {
			consultContent := consult.Content
			consultFormatingName, err := session.ExtractFormat(string(consultContent))
			if err != nil {
				fmt.Fprintf(os.Stderr, "error en parsing: %v", err)
			}

			Queries := strings.Split(string(consultContent), ";")

			for _, query := range Queries {
				DSN := s.Fetching[arg].DSN
				DBType := s.Fetching[arg].Type
				// EXPLAIN runs against the source (metadata); the result is a
				// SELECT qualified with the ATTACH catalog (arg) so the data
				// itself is read through the attached source on s.DB.
				fetchQuery, err := session.ParseSELECTlightFetch(query, DBType, DSN, arg)
				if err != nil {
					fmt.Fprintf(os.Stderr, "parsing Fetch query %s: %v\n", arg, err)
					continue
				}
				wg.Add(1)
				go func() {
					defer wg.Done()

					queryName, err := session.ExtractNamedQuery(query, consultFormatingName)
					if err != nil {
						fmt.Fprintf(os.Stderr, "parsing Fetch query %s: %v\n", arg, err)
						return
					}

					// Read through the ATTACHed source and write into the local
					// store in a single server-side statement (no round-trip
					// through Go memory).
					ctx := context.Background()
					schema := quoteIdent(arg + "Backups")
					table := quoteIdent(queryName)

					if _, err := s.DB.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS "+schema); err != nil {
						fmt.Fprintf(os.Stderr, "creando schema destino %s: %v\n", arg, err)
						return
					}
					create := fmt.Sprintf("CREATE OR REPLACE TABLE %s.%s AS %s", schema, table, fetchQuery)
					if _, err := s.DB.ExecContext(ctx, create); err != nil {
						fmt.Fprintf(os.Stderr, "materializando %s.%s: %v\n", schema, table, err)
					}
				}()
			}
		} else {
			fmt.Fprintf(os.Stderr, "%v is not a registered consult in the current session %v\n", arg, s.ID)
			wg.Wait()
			return
		}
	}

	// Wait for every fetch/write goroutine before the deferred dbWrite.Close().
	wg.Wait()
}

// quoteIdent wraps an identifier in double quotes for safe use as a SQL schema
// or table name, escaping any embedded double quotes.
func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
