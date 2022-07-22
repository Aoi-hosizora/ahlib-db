package xdbutils_orderby

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"runtime"
	"testing"
)

func TestGenerateOrderByExpr(t *testing.T) {
	dict := PropertyDict{
		"uid":      NewPropertyValue(false, "uid"),
		"username": NewPropertyValue(false, "firstname", "lastname"),
		"age":      NewPropertyValue(true, "birthday"),
	}
	for _, tc := range []struct {
		giveKey          string
		wantReverse      bool
		wantDestinations []string
	}{
		{"uid", false, []string{"uid"}},
		{"username", false, []string{"firstname", "lastname"}},
		{"age", true, []string{"birthday"}},
	} {
		xtestingEqual(t, dict[tc.giveKey].Reverse(), tc.wantReverse)
		xtestingEqual(t, dict[tc.giveKey].Destinations(), tc.wantDestinations)
	}

	for _, tc := range []struct {
		giveSource string
		giveDict   PropertyDict
		want       string
	}{
		{"", dict, ""},
		{" ", dict, ""},
		{"uid", nil, ""},
		{"id", dict, ""},

		{"uid", dict, "uid ASC"},
		{"  uid ", dict, "uid ASC"},
		{"uid asc", dict, "uid ASC"},
		{"uid desc", dict, "uid DESC"},
		{"uid DESC", dict, "uid DESC"},
		{"uid asc xxx", dict, "uid ASC"},
		{"uid desc xxx", dict, "uid DESC"},
		{"uid xxx", dict, "uid ASC"},
		{"uid,", dict, "uid ASC"},
		{"uid  , id", dict, "uid ASC"},
		{"uid,, ,", dict, "uid ASC"},

		{"uid,username", dict, "uid ASC, firstname ASC, lastname ASC"},
		{"uid, username", dict, "uid ASC, firstname ASC, lastname ASC"},
		{"username desc, age", dict, "firstname DESC, lastname DESC, birthday DESC"},
		{"username, age desc", dict, "firstname ASC, lastname ASC, birthday ASC"},
	} {
		xtestingEqual(t, GenerateOrderByExpr(tc.giveSource, tc.giveDict), tc.want)
	}
}


// =============================
// simplified xtesting functions
// =============================

func failTest(t testing.TB, failureMessage string) bool {
	_, file, line, _ := runtime.Caller(2)
	_, _ = fmt.Fprintf(os.Stderr, "%s:%d %s\n", path.Base(file), line, failureMessage)
	t.Fail()
	return false
}

func xtestingEqual(t testing.TB, give, want interface{}) bool {
	if give != nil && want != nil && (reflect.TypeOf(give).Kind() == reflect.Func || reflect.TypeOf(want).Kind() == reflect.Func) {
		return failTest(t, fmt.Sprintf("Equal: invalid operation `%#v` == `%#v` (cannot take func type as argument)", give, want))
	}
	if !reflect.DeepEqual(give, want) {
		return failTest(t, fmt.Sprintf("Equal: expect to be `%#v`, but actually was `%#v`", want, give))
	}
	return true
}