# codeql query run --external demo

See `codeql query run` documentation and [github discussion](https://github.com/github/codeql/discussions/21050#discussioncomment-15277223).

`--external` can do many magic things.

To run this demo, a codeql database is needed.

run `codeql query run demo_ext.ql --external=foo=test.csv -d <database>`. The result will be:

```
|    a    |     b     |
+---------+-----------+
| hello   |  world    |
| goodbye |  universe |
```
