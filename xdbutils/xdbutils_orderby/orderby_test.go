package xdbutils_orderby

import (
	"github.com/Aoi-hosizora/ahlib-db/xdbutils/internal"
	"strings"
	"testing"
)

func TestOptions(t *testing.T) {
	dict := PropertyDict{"a": NewPropertyValue(true, "aa"), "b": NewPropertyValue(false, "bb1", "bb2")}

	generated1 := GenerateOrderByExpr(
		"a|b", dict,
		WithSourceSeparator("|"),
		WithTargetSeparator(","),
	)
	internal.TestEqual(t, generated1, "aa DESC,bb1 ASC,bb2 ASC")

	generated2 := GenerateOrderByExpr(
		"a desc, b", dict,
		WithSourceSeparator(""), // ","
		WithTargetSeparator(""), // ", "
		WithSourceProcessor(func(source string) (field string, asc bool) { return "b", false }),
		WithTargetProcessor(func(destination string, asc bool) (target string) { return destination }),
	)
	internal.TestEqual(t, generated2, "bb1, bb2, bb1, bb2")

	generated3 := GenerateOrderByExpr(
		"a desc, b desc", dict,
		WithSourceProcessor(nil),
		WithTargetProcessor(nil),
	)
	internal.TestEqual(t, generated3, "aa ASC, bb1 DESC, bb2 DESC")

	generated4 := GenerateOrderByExpr(
		"a+|b-", dict,
		WithSourceSeparator("|"),
		WithTargetSeparator(",,"),
		WithSourceProcessor(func(source string) (field string, asc bool) {
			field = strings.TrimRightFunc(source, func(r rune) bool { return r == '+' || r == '-' })
			asc = strings.LastIndex(source, "-") == -1
			return field, asc
		}),
		WithTargetProcessor(func(destination string, asc bool) (target string) {
			if asc {
				return destination + " ASCENDING"
			}
			return destination + " DESCENDING"
		}),
	)
	internal.TestEqual(t, generated4, "aa DESCENDING,,bb1 DESCENDING,,bb2 DESCENDING")
}

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
		internal.TestEqual(t, dict[tc.giveKey].Reverse(), tc.wantReverse)
		internal.TestEqual(t, dict[tc.giveKey].Destinations(), tc.wantDestinations)
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
		internal.TestEqual(t, GenerateOrderByExpr(tc.giveSource, tc.giveDict), tc.want)
	}
}
