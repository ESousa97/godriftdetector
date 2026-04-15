// Package domain provides the core business logic for godriftdetector.
//
// It contains the models representing the desired and actual states of the
// infrastructure, as well as the [Comparator] responsible for detecting
// discrepancies (drifts) between them.
//
// The primary entry point for comparisons is the [Comparator.Compare] method,
// which returns a [ComparisonResult] detailing any missing services,
// shadow IT, or configuration mismatches.
package domain
