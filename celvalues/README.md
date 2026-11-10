# Typed CEL dates and currency

`celvalues` supplies civil dates and CEL integration for `commons/decimal.Money` and `commons/bigdecimal.Money`. Gomplate enables its CEL library automatically. Applications using cel-go directly add `celvalues.Library()`; custom function callbacks adapt their results with `celvalues.NewAdapter(env.CELTypeAdapter()).NativeToValue(value)`.

Declare direct cel-go record variables as `cel.DynType` or `cel.MapType(cel.StringType, cel.DynType)` so CEL's static checker permits comparisons between their fields and primitive literals.

```go
dueDate, err := celvalues.NewDate("2026-11-10")
if err != nil {
    return err
}
premium, err := decimal.NewMoney(decimal.MoneyOptions{
    Amount: "125.005",
    Currency: "USD",
    DecimalPlaces: 2,
})
if err != nil {
    return err
}
result, err := gomplate.RunExpression(map[string]any{
    "policy": map[string]any{"DueDate": dueDate, "PremiumDue": premium},
}, gomplate.Template{
    Expression: `policy.DueDate <= "2026-11-10" && policy.PremiumDue > 100`,
    Strict: true,
})
```

`NewDate` accepts an exact `YYYY-MM-DD` calendar date. Dates compare to other dates, ISO date strings, or the UTC calendar day of a CEL timestamp. `string(policy.DueDate)` returns `2026-11-10`. Actual timestamp values retain CEL's timestamp behavior.

Import `decimal` from `github.com/flanksource/commons/decimal`. `Amount` is a finite decimal string rounded half-even to the backend precision: `decimal` supports 19 digits; `bigdecimal` provides decimal128 precision. `Currency` is a three-letter uppercase code. `DecimalPlaces` is the minimum display scale, between 0 and 31; extra stored fractional precision is retained. `string(policy.PremiumDue)` above returns `USD 125.005`. A value of `125` with scale 2 returns `USD 125.00`.

Money compares to numbers or money of the same precision type and currency. All comparison operators work in either operand position. Comparing different currencies or precision types is an error. Implicit currency conversion is not provided.

`bigdecimal.Money` supports exact arithmetic, and the result is always a Money in the operand's currency:

| Expression | Result |
| --- | --- |
| `money + money`, `money - money` | Same currency only; keeps whichever operand shows more decimal places |
| `money + n`, `n + money`, `money - n`, `n - money` | `n` (int, uint or double) is an amount in the Money's currency |
| `money * n`, `n * money`, `money / n` | Scales the amount |
| `-money` | Negates the amount |

Doubles are read through their shortest decimal form, so `money - 0.1 - 0.2 == money - 0.3` holds exactly. Multiplying or dividing two Money values, dividing a number by Money, mixing currencies, dividing by zero, and any arithmetic on `decimal.Money` are errors. Numeric arithmetic that involves no Money keeps CEL's own behaviour. Dates and money remain native values through CEL-created maps/lists and custom functions; converting a value to JSON explicitly renders its string representation.

String collection `.sum()` uses `commons/decimal` parsing and summation, rounding inputs and the final result to its precision without intermediate rounding of additions. Results have no insignificant trailing zeros. Invalid inputs and overflow return errors; sufficiently small inputs round to zero.

Run the scalar and Gomplate integration specs from the module root:

```sh
go test ./celvalues
```
