Translating a single `Consult` query into the concrete query that will actually hit the storage device is the main functionality of the fetch parsing. This is the part of [[parseFeature]] that lives inside the session and decides *how* the data is retrieved.

The user expresses intent through a RUST-like attribute placed at the top of the query, and the parser turns that attribute into a fetch plan.

>[!tip] Delegate the heavy lifting
> The parser only decides the *shape* of the retrieval. The actual SQL execution (`EXPLAIN`, the final `SELECT`) is delegated to the source database and to `duckdb`, keeping our surface small and secure.

# The `derive` attribute
A query opts into a fetch workflow with:

```sql
#[derive(fetch.LightFetch)]
SELECT ... FROM ... WHERE ... ;
```

The attribute is parsed with the regex `^\s*#\[derive\(fetch\.([a-zA-Z]+)\)\]`, the captured option selects the workflow, and the attribute itself is stripped before the query is forwarded.

>[!warning] Leading whitespace matters
> Queries read from a `.sql` file are split on `;`, so each one keeps the newline left by the previous statement. The anchor must tolerate leading whitespace (`^\s*`) or every query after the first silently parses to an empty fetch.

# Fetch workflows
+ [x] **FullFetch** — retrieve the query as written. The attribute is removed and the remaining SQL is passed straight through.
+ [ ] **LightFetch** — retrieve only what is needed. The source database is asked for its plan (`EXPLAIN FORMAT=JSON`), and a minimal `SELECT` is rebuilt from the columns and conditions the plan actually touches.

## LightFetch reconstruction
From the `EXPLAIN` plan we read the table name, the used columns and the attached condition, then rebuild:

```sql
SELECT <used_columns> FROM <table_name> [WHERE <attached_condition>] ;
```

>[!warning] The `FROM` clause must not be dropped
> The reconstruction has to include `FROM <table_name>`; a `SELECT` with only columns and a `WHERE` is not valid SQL. The table name comes from the plan and must be carried through.

# Supported sources
+ [x] `MySQL` — via `EXPLAIN FORMAT=JSON`, unmarshalled through `QueryExplainMySQL`.
+ [ ] `PostgreSQL` — pending.

>[!tip] Extending to a new source
> Implement the `QueryExplain` interface (`GetTableName`, `GetUsedColumns`, `GetAttachedConditions`) for the new engine and add its case to the dispatch, so the reconstruction logic stays shared.
