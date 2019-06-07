package cmd

import "testing"

func assertNoErr(t *testing.T, e error) {
	if e != nil {
		t.Error(e)
	}
}
