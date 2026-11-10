---
title: CEL Collections
weight: 14
menu: main
---

CEL expressions can group, select and join native collections without converting them to JSON or strings.

```cel
rows.groupBy(r, r.fund).map(g,
  {"fund": g.key, "amount": g.rows.map(r, r.amount).sum()})
```

`groupBy(row, key)` returns groups with `key` and `rows`, preserving first occurrence order. Keys must be strings, integers, unsigned integers or booleans. Null and error keys fail evaluation.

`distinctBy(row, key)` returns the first row for each key. `latestBy(row, key)` returns the row with the greatest comparable key, such as a timestamp. It rejects empty input and tied maximum keys. `one()` requires exactly one row.

```cel
funds.groupBy(r, r.fund).map(g, g.rows.latestBy(r, r.created))
```

`left.joinBy(right, leftRow, leftKey, rightRow, rightKey)` is an inner join. It returns `{left, right}` pairs for every matching key, preserving left order and then right order. Unmatched rows are omitted; duplicate keys produce all matching pairs. Key rules match `groupBy`.

Joins reject results exceeding 10,000 pairs. Decimal `sum()` rejects individual strings longer than 4,096 bytes before parsing them.

```cel
before.joinBy(after, b, b.fund, a, a.fund)
```

For a list of plain decimal strings, `sum()` uses exact decimal arithmetic and returns a decimal string. For example, `["0.1", "0.2"].sum()` returns `"0.3"`. Invalid decimal strings fail evaluation. Integer and floating point lists keep their existing overloads.

The Go API supports `Template{Strict: true}` for expressions that must fail on missing values. Strict and normal expression caches are separate. `CELValue` converts CEL lists and maps to native Go values while preserving integer precision and native host objects.
