A core object of `etrds` which holds the state of an interaction between the end user and the **ET(L)** process. A session keeps track of an in-memory `duckdb` instance, the collection of `Consult`s the user has registered and the lifecycle of each one of them. In short, the session is the runtime glue that the subcommands ([[showCommand]], [[parseFeature]], [[CollectCommand]]) operate over.

>[!info] Why a session?
> Most subcommands need the same things: an open `duckdb` connection, the read-only attachments to the storage devices and the set of queries (`Consults`) being worked on. Instead of re-opening and re-attaching on every call, the session centralizes that lifecycle so the commands stay thin.

# Anatomy of a Session
A session is made of:
+ A **name** identifying the working session.
+ An in-memory `duckdb` connection used as the OLAP/storage motor.
+ A map of **`Consults`**, keyed by name. Each `Consult` carries its query, the local (`duckdb`) connection, the fetch database type/connection and its current `State`.

## Consult lifecycle
Each `Consult` advances through a small state machine:

```
registered -> staged -> running -> fetched -> stored -> done
```

# Brief Description of `duckdb` ATTACH
The session opens an in-memory `duckdb` and attaches every known storage device in **read-only** mode:

```sql
ATTACH '<location>' (READ_ONLY) AS <consultName>;
```

This lets a single session query across several attached databases without mutating any source, keeping the tool safe by default. See `parseFeature` for how the queries themselves are turned into a fetch plan.

# Goal overall
+ [ ] Provide an exported constructor/getter so sub-commands can obtain a ready session (`GetSession`/`NewETRDSSession`).
+ [ ] Make `etrdsSession` satisfy the `Session` interface by implementing `Scheduler() error`.
+ [ ] Fix the `ATTACH` statement (read-only syntax and the missing location field on `Consult`).
+ [ ] Implement the health check.
+ [ ] Connect the session with the fetch parsing.
+ [ ] Drive the `Consult` state machine (`registered -> ... -> done`) as work progresses.
