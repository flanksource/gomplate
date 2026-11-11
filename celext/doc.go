// Package celext copies cel-go's Strings, Encoders, Lists, Math, and Sets extension libraries.
// It preserves upstream behavior, including cost models, by default. Callers that never
// estimate cost can use StringsWithoutCostModels, EncodersWithoutCostModels,
// ListsWithoutCostModels, MathWithoutCostModels, and SetsWithoutCostModels to avoid
// compiling overload cost models on every environment configuration.
//
// Regenerate with make celext after changing the cel-go dependency. Remove this copy
// when cel-go compiles cost models once per model set.
//
//go:generate env GOWORK=off go run ./internal/sync
package celext
