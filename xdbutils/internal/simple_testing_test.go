package internal

import (
	"errors"
	"fmt"
	"path"
	"runtime"
	"testing"
)

// ================================
// simplified xtesting's unit tests
// ================================

func fail(t *testing.T) {
	_, file, line, _ := runtime.Caller(1)
	fmt.Printf("!!! testing on %s:%d is failed !!!\n", path.Base(file), line)
	t.Fail()
}

type testFlag uint8

const (
	positive testFlag = iota
	negative
	abnormal
)

func TestTestEqual(t *testing.T) {
	mockT := &testing.T{}

	type myType string
	var m map[string]interface{}

	for _, tc := range []struct {
		giveG, giveW interface{}
		want         testFlag
	}{
		// expect to Equal
		{"Hello World", "Hello World", positive},
		{123, 123, positive},
		{123.5, 123.5, positive},
		{[]byte("Hello World"), []byte("Hello World"), positive},
		{nil, nil, positive},
		{int32(123), int32(123), positive},
		{uint64(123), uint64(123), positive},
		{myType("1"), myType("1"), positive},
		{&struct{}{}, &struct{}{}, positive},

		// expect to NotEqual
		{"Hello World", "Hello World!", negative},
		{123, 1234, negative},
		{123.5, 123.55, negative},
		{[]byte("Hello World"), []byte("Hello World!"), negative},
		{nil, new(struct{}), negative},
		{10, uint(10), negative},
		{m["bar"], "something", negative},
		{myType("1"), myType("2"), negative},

		// expect to fail in all cases
		{func() {}, func() {}, abnormal},
		{func() int { return 23 }, func() int { return 42 }, abnormal},
	} {
		pos := TestEqual(mockT, tc.giveG, tc.giveW)
		if (tc.want == positive && !pos) || (tc.want != positive && pos) {
			fail(t)
		}
	}
}

func TestTestPanic(t *testing.T) {
	mockT := &testing.T{}

	for _, tc := range []struct {
		give func()
		want testFlag
	}{
		// expect to Panic
		{func() { panic("Panic!") }, positive},
		{func() { panic(0) }, positive},
		{func() { panic(nil) }, positive},

		// expect to NotPanic
		{func() {}, negative},
	} {
		pos := TestPanic(mockT, true, tc.give)
		if (tc.want == positive && !pos) || (tc.want != positive && pos) {
			fail(t)
		}
		neg := TestPanic(mockT, false, tc.give)
		if (tc.want == negative && !neg) || (tc.want != negative && neg) {
			fail(t)
		}
	}

	for _, tc := range []struct {
		giveF func()
		giveW interface{}
		want  testFlag
	}{
		// expect to pass PanicWithValue
		{func() { panic("Panic!") }, "Panic!", positive},
		{func() { panic(0) }, 0, positive},
		{func() { panic(nil) }, nil, positive},
		{func() { panic(errors.New("panic")) }, errors.New("panic"), positive},

		// expect to fail PanicWithValue
		{func() {}, nil, negative},
		{func() { panic("Panic!") }, "Panic", negative},
		{func() { panic(uint8(0)) }, 0, negative},
		{func() { panic(errors.New("panic")) }, "panic", negative},
	} {
		pos := TestPanic(mockT, true, tc.giveF, tc.giveW)
		if (tc.want == positive && !pos) || (tc.want != positive && pos) {
			fail(t)
		}
	}
}
