A sub-command which collect data from a *storage device* (databases: MySQL, DuckDB) using a `.sql` file which is previously parsed into a plan of execution. Basically, `collect` from `etrds` is the complement of `parse` command.

>[!warning] The heart of `etrds`
> Through the planning of these command implementation, I decided that the **heart** of `etrds` is the parsing. To be precise, allocate the logic of certain queries (complex queries, rutinary queries) and ease the interactions of the end user, who is the person which requires some information, with the **ET(L)** process.
> Therefore, I incorporate a new command named `parse`.
