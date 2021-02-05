package orderby

import (
	"github.com/Aoi-hosizora/ahlib/xtesting"
	"testing"
)

func TestGenerateOrderByExp(t *testing.T) {
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
		xtesting.Equal(t, dict[tc.giveKey].Reverse(), tc.wantReverse)
		xtesting.Equal(t, dict[tc.giveKey].Destinations(), tc.wantDestinations)
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
		{"uid asc asc", dict, "uid ASC"},
		{"uid xxx", dict, "uid ASC"},
		{"uid,", dict, "uid ASC"},
		{"uid  , id", dict, "uid ASC"},
		{"uid,, ,", dict, "uid ASC"},

		{"uid,username", dict, "uid ASC, firstname ASC, lastname ASC"},
		{"uid, username", dict, "uid ASC, firstname ASC, lastname ASC"},
		{"username desc, age", dict, "firstname DESC, lastname DESC, birthday DESC"},
		{"username, age desc", dict, "firstname ASC, lastname ASC, birthday ASC"},
	} {
		xtesting.Equal(t, GenerateOrderByExp(tc.giveSource, tc.giveDict), tc.want)
	}
}
