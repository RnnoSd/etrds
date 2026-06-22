A feature of `etrds` API which translate SQL custom queries to full end-to-end procedure execution.

The main goal is to convert simple instruction into full pipeline guidelines for execution. This means how to retrieve data from an `rds`, how to modify and how to present it. All must be achieve by the user through RUST like attributes' inside `sql` which are parsed. This feature centralized its work on the orchestration and delegate the actual transforming and parsing common SQL syntax to `duckDB`.
> [!tip] Simplify thinks
> Delegating responsibilities to other tools make the tool more secure at the start. And improves the final goal shaping of the team.
# Brief `duckdb` Description
## Web page
### Why `duckdb`?
There are many database management systems (DBMS) out there. But there is no one-size-fits-all database system. All take different trade-offs to better adjust to specific use cases. `duckdb` is no different. Here, we try to explain what goals `duckdb` has and why and how we try to achieve those goals through technical means. To start with, `duckdb` is a relational (table-oriented) DBMS that supports the Structured Query Language (SQL).
#### Key Characteristics of `duckdb`
##### Simple
`duckdb` adopted the ideas of simplicity and in-process operation – but later stepped out of the realm of in-process operations through the **Quack protocol**.

`duckdb` has no external dependencies, neither for compilation nor during run-time. For releases, the entire source tree of `duckdb` is compiled into two files, a header and an implementation file, a so-called "amalgamation". This greatly simplifies deployment and integration in other build processes.

For `duckdb`, there is no DBMS server software to install, update and maintain. `duckdb` does not run as a separate process, but completely embedded within a host process.
##### Portable
Thanks to having no dependencies, DuckDB is extremely portable. It can be compiled for all major operating systems and CPU architectures.
#### Feature-Rich
`duckdb` provides serious data management features. There is extensive support for complex queries in SQL with a large function library, window functions, etc. `duckdb` provides transactional guarantees (ACID properties) through our custom, bulk-optimized Multi-Version Concurrency Control (MVCC). Data can be stored in `duckdb`'s native format, a single-file database or in one of the supported **lakehouse formats**. `duckdb`'s native format supports secondary indexes to speed up queries trying to find a single table entry, while lakehouse formats, including the **DuckLake** format can scale up to petabytes of data.
#### Fast
`duckdb` is designed to support analytical query workloads, also known as OLAP. These workloads are characterized by complex, relatively long-running queries that process significant portions of the stored dataset, for example aggregations over entire tables or joins between several large tables.
#### Extensible
`duckdb` offers a flexible extension mechanism that allows defining new data types, functions, file formats and new SQL syntax. In fact, many of `duckdb`'s key features, such as support for the Parquet file format, JSON, time zones, and support for the HTTP(S) and S3 protocols are implemented as extensions.
#### Free
`duckdb`'s development started while the main developers were public servants in the Netherlands. We see it as our responsibility and duty to society to make the results of our work freely available to anyone in the Netherlands or elsewhere.
#### Thoroughly Tested
While `duckdb` was originally created by a research group, it was never intended to be a research prototype. Instead, it was intended to become a stable and mature database system. To facilitate this stability, `duckdb` is intensively tested using Continuous Integration. `duckdb`'s test suite currently contains millions of queries, and includes queries adapted from the test suites of SQLite, PostgreSQL, and MonetDB. Tests are repeated on a wide variety of platforms and compilers. Every pull request is checked against the full test setup and only merged if it passes.
# Goal overall 
+ [ ] Connect the `collectFeature` with this part. Create a first part parsing which divides the SQL queries into *RETRIEVING section*, the *OLAP section* and the *STORAGE section*.
+ [ ] Connect the *OLAP section* to `duckdb`motor.
+ [ ] Connect the STORAGE section to `duckdb` format of storage.