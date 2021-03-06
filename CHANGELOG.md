These entries focus on changes we think are relevant to users of Brim,
zq, or pcap.  For all changes to zqd, its API, or to other components in the
zq repo, check the git log.

## v0.22.0
* zq: Change the implementation of the `union` type to conform with the [ZNG spec](https://github.com/brimsec/zq/blob/master/zng/docs/spec.md#3114-union-typedef) (#1245)
* zq: Make options/flags and version reporting consistent across CLI tools (#1249, #1254, #1256, #1296, #1323, #1334, #1328)
* zqd: Fix an issue that was preventing flows in nanosecond pcaps from opening in Brim (#1243, #1241)
* zq: Fix an issue where the TZNG reader did not recognize a bad record type as a syntax error (#1260)
* zq: Add a CSV writer (`-f csv`) (#1267, #1300)
* zqd: Add an endpoint for returning results in CSV format (#1280)
* zqd: Add an endpoint for returning results in NDJSON format (#1283)
* zapi: Add an option to return results as a JSON array (`-e json`) (#1285)
* zapi: Add output format options/flags to `zapi get` (#1278)
* zqd: Add an endpoint for creating/querying search indexes (#1272)
* zapi: Add commands `zapi index create|find` for creating/querying search indexes (#1289)
* pcap: Mention ICMP protocol filtering (`-p icmp`) in help text (#1281)
* zq: Point to new Slack community URL https://www.brimsecurity.com/join-slack/ in docs (#1304)
* zqd: Fix an issue where starting `zqd listen` created excess error messages when subdirectories were present (#1303)
* zql: Add the [`fuse` processor](https://github.com/brimsec/zq/tree/master/zql/docs/processors#fuse) for unifying records under a single schema (#1310, #1319, #1324)
* zql: Fix broken links in documentation (#1321, #1339)
* zst: Introduce the [ZST format](https://github.com/brimsec/zq/blob/master/zst/README.md) for columnar data based on ZNG (#1268, #1338)
* pcap: Fix an issue where certain pcapng files could fail import with a `bad option length` error (#1341)
* zql: [Document the `**` operator](https://github.com/brimsec/zq/tree/master/zql/docs/search-syntax#wildcard-field-names) for type-sepcific searches that look within nested records (#1337)
* zar: Change the archive data file layout to prepare for handing chunk files with overlapping ranges and improved S3 support (#1330)
* zar: Support archive data files with overlapping time spans (#1348)
* zqd: Add a page containing guidance for users that directly access the root `zqd` endpoint in a browser (#1350)
* pcap: Add a `pcap info` command to print summary/debug details about a packet capture file (#1354)
* zqd: Fix an issue with empty records (#1353)
* zq: Fix an issue where interrupted aggregations could leave behind temporary files (#1357)
* zng: Add a marshaler to generate ZNG streams from native Go values (#1327)

## v0.21.0
* zq: Improve performance by making fewer API calls in S3 reader (#1191)
* zq: Use memory more efficiently by reducing allocations (#1190, #1201)
* zqd: Fix an issue where a pcap moved/deleted after import caused a 404 response and white screen in Brim (#1198)
* zqd: Include details on [adding observability](https://github.com/brimsec/zq/tree/master/k8s#adding-observability) to the docs for running `zqd` in Kubernetes (#1173)
* zq: Improve performance by removing unnecessary type checks (#1192, #1205)
* zq: Add additional Boyer-Moore optimizations to improve search performance (#1188)
* zq: Fix an issue where data import would sometimes fail with a "too many files" error (#1210)
* zq: Fix an issue where error messages sometimes incorrectly contained the text "(MISSING)" (#1199)
* zq: Fix an issue where non-adjacent record fields in Zeek TSV logs could not be read (#1225, #1218)
* zql: Fix an issue where `cut -c` sometimes returned a "bad uvarint" error (#1227)
* zq: Add support for empty ZNG records and empty NDJSON objects (#1228)
* zng: Fix the tag value examples in the [ZNG spec](https://github.com/brimsec/zq/blob/master/zng/docs/spec.md) (#1230)
* zq: Update LZ4 dependency to eliminate some memory allocations (#1232)
* zar: Add a `-sortmem` flag to allow `zar import` to use more memory to improve performance (#1203)
* zqd: Fix an issue where file paths containing URI escape codes could not be opened in Brim (#1238)

## v0.20.0
* zqd: Publish initial [docs](https://github.com/brimsec/zq/blob/master/k8s/README.md) for running `zqd` in Kubernetes (#1101)
* zq: Provide a better error message when an invalid IP address is parsed (#1106)
* zar: Use single files for microindexes (#1110)
* zar: Fix an issue where `zar index` could not handle more than 5 "levels" (#1119)
* zqd: Fix an issue where `zapi pcappost` incorrectly reported a canceled operation as a Zeek exit (#1139)
* zar: Add support for empty microindexes, also fixing an issue where `zar index` left behind empty files after an error (#1136)
* zar: Add `zar map` to handle "for each file" operations (#1138, #1148)
* zq: Add Boyer-Moore filter optimization to ZNG scanner to improve performance (#1080)
* zar: Change "zdx" to "microindex" (#1150)
* zar: Update the [`zar` README](https://github.com/brimsec/zq/blob/master/cmd/zar/README.md) to reflect recent changes in commands/output (#1149)
* zqd: Fix an issue where text stack traces could leak into ZJSON response streams (#1166)
* zq: Fix an issue where an error "slice bounds out of range" would be triggered during attempted type conversion (#1158)
* pcap: Fix an issue with pcapng files that have extra bytes at end-of-file (#1178)
* zqd: Add a hidden `-brimfd` flag to `zqd listen` so that `zqd` can close gracefully if Brim is terminated abruptly (#1184)
* zar: Perform `zar zq` queries concurrently where possible (#1165, #1145, #1138, #1074)

## v0.19.1

* zq: Move third party license texts in zq repo to a single [acknowledgments.txt](https://github.com/brimsec/zq/blob/master/acknowledgments.txt) file (#1107)
* zq: Automatically load AWS config from shared config file `~/.aws/config` by default (#1109)
* zqd: Fix an issue with excess characters in Space names after upgrade (#1112)

## v0.19.0
* zq: ZNG output is now LZ4-compressed by default (#1050, #1064, #1063, [ZNG spec](https://github.com/brimsec/zq/blob/master/zng/docs/spec.md#313-compressed-value-message-block))
* zar: Adjust import size threshold to account for compression (#1082)
* zqd: Support starting `zqd` with datapath set to an S3 path (#1072)
* zq: Fix an issue with panics during pcap import (#1090)
* zq: Fix an issue where spilled records were not cleaned up if `zq` was interrupted (#1093, #1099)
* zqd: Add `-loglevel` flag (#1088)
* zq: Update help text for `zar` commands to mention S3, and other improvements (#1094)
* pcap: Fix an out-of-memory issue during import of very large pcaps (#1096)

## v0.18.0
* zql: Fix an issue where data type casting was not working in Brim (#1008)
* zql: Add a new [`rename` processor](https://github.com/brimsec/zq/tree/master/zql/docs/processors#rename) to rename fields in a record (#998, #1038)
* zqd: Fix an issue where API responses were being blocked in Brim due to commas in Content-Disposition headers (#1014) 
* zq: Improve error messaging on S3 object-not-found (#1019)
* zapi: Fix an issue where `pcappost` run with `-f` and an existing Space name caused a panic (#1042)
* zqd: Add a `-prometheus` option to add [Prometheus](https://prometheus.io/) metrics routes the API (#1046)
* zq: Update [README](https://github.com/brimsec/zq/blob/master/README.md) and add docs for more command-line tools (#1049)

## v0.17.0
* zq: Fix an issue where the inferred JSON reader crashed on multiple nested fields (#948)
* zq: Introduce spill-to-disk groupby for performing very large aggregations (#932, #963)
* zql: Use syntax `c=count()` instead of `count() as c` for naming the field that holds the value returned by an aggregate function (#950)
* zql: Fix an issue where attempts to `tail` too much caused a panic (#958)
* zng: Readability improvements in the [ZNG specification](https://github.com/brimsec/zq/blob/master/zng/docs/spec.md) (#935)
* zql: Fix an issue where use of `cut`, `put`, and `cut` in the same pipeline caused a panic (#980)
* zql: Fix an issue that was preventing the `uniq` processor from  working in the Brim app (#984)
* zq: Fix an issue where spurious type IDs were being created (#964)
* zql: Support renaming a field via the `cut` processor (#969)

## v0.16.0
* zng: Readability improvements in the [ZNG specification](https://github.com/brimsec/zq/blob/master/zng/docs/spec.md) (#897, #910, #917)
* zq: Support directory output to S3 (#898)
* zql: Group-by no longer emits records in "deterministic but undefined" order (#914)
* zqd: Revise constraints on Space names (#853, #926, #944, #945)
* zqd: Fix an issue where a file replacement race could cause an "access is denied" error in Brim during pcap import (#925)
* zng: Revise [Zeek compatibility](https://github.com/brimsec/zq/blob/master/zng/docs/zeek-compat.md) doc (#919)
* zql: Clarify [`cut` processor documentation](https://github.com/brimsec/zq/tree/master/zql/docs/processors#cut) (#924)
* zqd: Fix an issue where an invalid 1970 Space start time could be created in Brim during pcap inport (#938)

## v0.15.0
* pcap: Report more detailed error information (#844)
* zql: Add a new function `Time.trunc()` (#842)
* zql: Support grouping by computed keys (#860)
* zq: Change implementation of `every X` to use a computed groupby key (#893)
* zql: Clean up the [ZQL docs](https://github.com/brimsec/zq/tree/master/zql/docs) (#884)
* zql: Change `cut` processor to emit any matching fields (#899)
* zq: Allow output to an S3 bucket (#889)

## v0.14.0
* zq: Add support for reading from S3 buckets (#733, #780, #783)
* zq: Add initial support for reading Parquet files (only via `-i parquet`, no auto-detection) (#736, #754, #774, #780, #782, #820, #813, #830, #825, #834)
* zq: Fix an issue with reading/writing recursively-nested NDJSON events (#748)
* zqd: Begin using a "runner" to invoke Zeek for processing imported pcaps (#718, #788)
* zq: Fix issues related to reading NDJSON during format detection (#752)
* zqd: Include stack traces on panic errors (#732)
* zq: Handle `\r\n` line endings generated by MinGW (Windows) Zeek (#775)
* zq: Support scientific notation for integer types (#768)
* zql: Add cast syntax to expressions (#765, #784)
* zq: Fix an issue where reads from stdin were described as being from `-` (#777)
* zq: Improve an NDJSON parsing error to be more detailed than "bad format" (#776)
* zjson: Fix an issue with aliases in the zjson writer (#793)
* zq: Fix an issue where typed JSON reads could panic when a field that was expected to contain an array instead contained a scalar (#799)
* zq: Fix an issue with ZNG handling of aliases on records (#801)
* zq: Fix an issue with subnet searches (#807)
* zapi: Introduce `zapi`, a simple CLI for interacting with `zqd` servers (#802, #809, #812)
* zq: Add arguments to generate CPU/memory profiles (#814)
* zql: Introduce time conversion functions (#822)
* zq: Ensure Spaces have non-blank names (#826)

## v0.13.1
* zq: Fix an issue with stream reset that was preventing the pcap button in Brim from activating (#725)
* zql: Allow multiple fields to be written from `put` processor (#697)

## v0.13.0
* zqd: Enable time indexing to provide faster query response in narrower time ranges (#647)
* zql: Make ipv4 subnet bases contain 4 octets to remove ambiguity between fractions & CIDR (#670)
* zq: Use an external sort for large inputs (removes the 10-million line `sort` limit) (#527)
* zq: Fix an issue where duplicate field names could be produced by aggregate functions & group-by (#676)
* zar: Introduce an experimental prototype for working with archived logs ([README](https://github.com/brimsec/zq/blob/master/cmd/zar/README.md)) (#700)
* zq: Support recursive record nesting in Zeek reader/writer (#715)
* zqd: Zeek log import support needed for Brim (#616, #517, #608, #592, #592, #582, #709)

## v0.12.0
* zql: Introduce `=~` and `!~` operators in filters for globs, regexps, and matching addresses against subnets (#604, #620)
* zq: When input auto-detect fails, include each attempted format's error (#616)
* zng: Binary format is now called "ZNG" and text format is called "TZNG" ("BZNG" has been retired) (#621, #630, #656)
* zql: `cut` now has a `-c` option to show all fields _not_ in the provided list (#639, #655)
* zq: Make `-f zng` (binary ZNG) the default `zq` output format, and introduce `-t` as shorthand for `-f tzng` (#654)

## v0.11.1
* zqd: Send HTTP status 200 for successful pcap search (#605)

## v0.11.0
* zql: Improve string search matching on field names (#570)
* pcap: Better handling of empty results (#572)
* zq: Introduce `-e` flag to allow for continued reads during input errors (#577)
* pcap: Allow reading of pcap files that have a capture length that exceeds the original length of the packet (#584)
* zqd: Fix an issue that was causing the histogram to draw incorrectly in Brim app (#602)

## v0.10.0

* zql: Let text searches match field names as well as values (#529)
* zql: Fix an issue where ZQL queries exceeding 255 chars caused a crash (#543)
* zql: Make searches case-insensitive by default (#536)
* Fix an issue where the Zeek reader failed to read whitespace from the rightmost column (#552)

## v0.9.0

* zql: Emit warnings from `put` processor (#477)
* zql: Add string functions (#475)
* zql: Narrow the use of `len()` to only sets/vectors, introduce new functions for string length (#485)
* zql: Add ternary conditional operator (#484)
* zqd: Add waterfall logger (#492)
* zqd: Make http shutdown more graceful (#500)
* zqd: Make space deletion cancel and await other operations (#451)

## v0.8.0

* zql: add the `put` processor that adds or updates fields using a computed
  expression. (#437)
* zql: add functions for use with put, like `Math.min`, `Math.max`, and others.
  (#453, #459, #461, #472)
* zq: support reading ndjson with user supplied type information. (#441)
* Fix an issue reading pcaps with snaplen=0. (#462)

## v0.7.0

* Address ingest issues for packet captures in legacy pcap format.
* Calculate and respond with packet capture time range at the start of ingest,
  so that Brim can immediately display the space's time range.

## v0.6.0

* zq now displays warnings by default; the "-W" flag is removed, replaced by
  the "-q" for quieting warnings.
* Update license to reflect new corporate name.
* Address ingest issues for some pcapng packet captures.
* Address ingest issues for file or path names that required uri encoding.

## v0.5.0

* Support search queries during pcap ingestion.
* Improved error reporting in zqd, especially during pcap ingestion.
* Improved performance of space info api.
* zqd supports ingesting pcapng formatted packet capture files.

## v0.4.0
  
* zqd adds an endpoint to create a new empty space via post
* zqd adds an endpoint to post packet captures that are indexed and turned into Zeek logs

## v0.3.0

* zqd adds -datadir flag for space root directory.
* zqd adds -version flag.
* Add pcap command to interact with packet capture files.

## v0.2.0

* Per-platform binaries will be available as Github release assets.
* zql examples under zql/docs are now verified via `make test-heavy`.
* Negative integers and floats are accepted in zql expressions.
* Internal integer types now match the ZNG specification.
* Fixed comparisons of aliased types.

## v0.1.0

* zq moves from github.com/mccanne/zq to github.com/brimsec/zq.
* Parser and AST moved to zq repo from github.com/looky-cloud/lookytalk.
* Query language name changed to ZQL.
* ZNG specification added.

## v0.0.1

* Initial release of zq.
