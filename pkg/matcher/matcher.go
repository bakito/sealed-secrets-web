package matcher

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/onsi/gomega"
	types2 "github.com/onsi/gomega/types"
	"github.com/pmezard/go-difflib/difflib"
)

func EqualDiff(expected string) types2.GomegaMatcher {
	return &EqualDiffMatcher{
		Expected: expected,
	}
}

func EqualFileDiff(path ...string) types2.GomegaMatcher {
	return &EqualFileDiffMatcher{
		Expected: filepath.Join(path...),
	}
}

type EqualDiffMatcher struct {
	Expected any
	diff     string
}

func (matcher *EqualDiffMatcher) Match(actual any) (success bool, err error) {
	if actual == nil && matcher.Expected == nil {
		return false, errors.New(
			"refusing to compare <nil> to <nil>.\nBe explicit and use BeNil() instead. " +
				"This is to avoid mistakes where both sides of an assertion are erroneously uninitialized",
		)
	}
	if actualByteSlice, ok := actual.([]byte); ok {
		if expectedByteSlice, ok := matcher.Expected.([]byte); ok {
			diff, err := unifiedDiff(string(expectedByteSlice), "Expected", string(actualByteSlice), "Actual")
			if err != nil {
				return false, err
			}
			matcher.diff = diff
			return diff == "", nil
		}
	}
	if actualString, ok := actual.(string); ok {
		if expectedString, ok := matcher.Expected.(string); ok {
			diff, err := unifiedDiff(expectedString, "Expected", actualString, "Actual")
			if err != nil {
				return false, err
			}
			matcher.diff = diff
			return diff == "", nil
		}
	}
	return false, fmt.Errorf("expected %s to be of type string or []byte", reflect.TypeOf(actual))
}

func (matcher *EqualDiffMatcher) FailureMessage(_ any) (message string) {
	return matcher.diff
}

func (matcher *EqualDiffMatcher) NegatedFailureMessage(_ any) (message string) {
	return matcher.diff
}

type EqualFileDiffMatcher struct {
	Expected any
	diff     string
}

func (matcher *EqualFileDiffMatcher) Match(actual any) (success bool, err error) {
	if actual == nil && matcher.Expected == nil {
		return false, errors.New(
			"refusing to compare <nil> to <nil>.\nBe explicit and use BeNil() instead. " +
				"This is to avoid mistakes where both sides of an assertion are erroneously uninitialized",
		)
	}
	if actualString, ok := actual.(string); ok {
		if expectedString, ok := matcher.Expected.(string); ok {
			actualFile := readFile(actualString)
			expectedFile := readFile(expectedString)

			diff, err := unifiedDiff(expectedFile, expectedString, actualFile, actualString)
			if err != nil {
				return false, err
			}
			matcher.diff = diff
			return diff == "", nil
		}
	}
	return false, fmt.Errorf("expected fiename %s to be of type string", reflect.TypeOf(actual))
}

func (matcher *EqualFileDiffMatcher) FailureMessage(_ any) (message string) {
	return matcher.diff
}

func (matcher *EqualFileDiffMatcher) NegatedFailureMessage(_ any) (message string) {
	return matcher.diff
}

func unifiedDiff(a, nameA, b, nameB string) (string, error) {
	ud := difflib.UnifiedDiff{
		FromFile: nameA,
		A:        difflib.SplitLines(a),
		ToFile:   nameB,
		B:        difflib.SplitLines(b),
		Context:  3,
	}
	return difflib.GetUnifiedDiffString(ud)
}

func readFile(path ...string) string {
	b, err := os.ReadFile(filepath.Join(path...))
	gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
	return string(b)
}
