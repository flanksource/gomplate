// Package celvalues provides civil dates and Commons Money integration for CEL.
//
// NewDate accepts YYYY-MM-DD. Date comparisons accept another Date, an ISO date
// string, or the UTC calendar day of a CEL timestamp. Construct currency amounts
// with commons/decimal.NewMoney or commons/bigdecimal.NewMoney. Money may be
// compared with a number or money in the same currency; different currencies
// produce an error. Both types support string() and preserve their native Go
// values through Gomplate evaluations, including nested maps and lists.
//
// Gomplate includes Library automatically. Plain cel-go callers add Library()
// to their environment options. Custom callbacks adapt Money through NewAdapter.
package celvalues
