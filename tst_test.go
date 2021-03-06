package tst

import (
	"testing"

	"github.com/jayacarlson/dbg"
)

var (
	initCount = 0
	finiCount = 0
)

func initFunction() {
	initCount += 1
}

func finiFunction() {
	finiCount += 1
}

func testEnabledPassed1() (string, bool) {
	// do some kind of test
	return dbg.IAm(), false // result is "Failed", so false
}

func testEnabledPassed2() (string, bool) {
	return dbg.IAm(), false // result is "Failed", so false
}

func testEnabledPassed3() (string, bool) {
	return dbg.IAm(), false // result is "Failed", so false
}

func testEnabledFailed() (string, bool) {
	return dbg.IAm(), true // result is "Failed"
}

func testDisabled() (string, bool) {
	return dbg.IAm(), true // should not get here
}

func testCounts() (string, bool) {
	failed := false
	failed = dbg.ChkTru(initCount == 3, "initFunction count incorrect") || failed
	failed = dbg.ChkTru(finiCount == 3, "finiFunction count incorrect") || failed
	return dbg.IAm(), failed
}


func TestTheEnabledTests(t *testing.T) {
	SetInitFunc(initFunction)
	SetFiniFunc(finiFunction)
	if Testing(dbg.IAm(), "Should see 3 successful tests", true) {
		Test(t, testEnabledPassed1)
		Test(t, testEnabledPassed2)
		Test(t, testEnabledPassed3)
	}
	ResetTestWrappers()
}

func TestCounts(t *testing.T) {
	if Testing(dbg.IAm(), "", true) {
		Test(t, testCounts)
	}
}

func TestIgnored(t *testing.T) {
	if Testing(dbg.IAm(), "Should see 'Ignored'", false) {
		Test(t, testDisabled)
	}
}

func TestFailedTest(t *testing.T) {
	if Testing(dbg.IAm(), "Should see a 'Failed' test", true) {
		Test(t, testEnabledFailed)
	}
}

func TestUserHandled(t *testing.T) {
	iam := dbg.IAm()
	if Testing(iam, "Should see 1 'Success' & 2 'Failed'", true) {
		// do some kind of test, then if passed:
		Passed(t, iam, "Passed user test...")

		// do some kind of test, then if failed:
		Failed(t, iam, "Failed user test...")

		// do a md5 check sum test on a file
		if Md5SumFile("tst_test.go", "failed-md5-sum-value") {
			Failed(t, iam, "Failed md5 test...  -- THIS IS A SUCCESS")
		} else {
			Passed(t, iam, "Passed md5 test...  -- THIS IS A FAILURE")
		}
	}
}
