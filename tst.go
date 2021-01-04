package tst

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/jayacarlson/dbg"
	"github.com/jayacarlson/env"
)

/*
	Some functions useful for test routines
*/

// Tests use this type, returning the name of the test and the result (Failed as TRUE)
type TestFunc func() (testName string, testResult bool)
type TestTFunc func(t *testing.T) (testName string, testResult bool)

var (
	initFunc                  func() // setable function to run before each test
	finiFunc                  func() // setable function to run after each test
	normClr, testClr, passClr string
	failClr, skipClr, redClr  string
	greenClr, magentaClr      string
)

//  Set the 'initFunc' to run before each test
func SetInitFunc(init func()) { initFunc = init }

//  Set the 'initFunc' to the stub function
func ClearTestInit() { initFunc = stubFunc }

//  Set the 'finiFunc' to run after each test
func SetFiniFunc(fini func()) { finiFunc = fini }

//  Set the 'finiFunc' to the stub function
func ClearTestFini() { finiFunc = stubFunc }

//	Reset the init & fini funcs to the stub function
func ResetTestWrappers() {
	initFunc = stubFunc
	finiFunc = stubFunc
}

// output "Testing" or "Ignore" for given test, just returns the given bool
func Testing(testName, description string, enabled bool) bool {
	if enabled {
		fmt.Printf(testClr+"  Testing "+normClr+" %-20s %s\n", testName, description)
	} else {
		fmt.Printf(skipClr+"  Ignore  "+normClr+" %-20s %s\n", testName, description)
	}
	return enabled
}

// black on green text to output
func Passed(t *testing.T, _ string, fstr string, a ...interface{}) {
	fmt.Printf(passClr+"  Passed  "+normClr+" "+fstr+"\n", a...)
}

// black on red text to output
func Failed(t *testing.T, who string, fstr string, a ...interface{}) {
	fmt.Printf(failClr+"  Failed  "+normClr+" "+fstr+"\n", a...)
	t.Errorf("%s", who)
}

// call the given 'test' function and output whether it Passed or Failed
func Test(t *testing.T, test TestFunc) bool {
	initFunc()
	defer finiFunc()
	st := time.Now()
	o, f := test()
	elapsed := time.Now().Sub(st)
	if o == "" {
		fl, ln := dbg.ErrWasAt()
		o = fmt.Sprintf("%s @ %d", fl, ln)
	}
	if f {
		Failed(t, o, "%15s: %s", elapsed.String(), o)
	} else {
		Passed(t, o, "%15s: %s", elapsed.String(), o)
	}
	return f
}

// call the given 'test' function and output whether it Passed or Failed
func TestT(t *testing.T, test TestTFunc) bool {
	initFunc()
	defer finiFunc()
	st := time.Now()
	o, f := test(t)
	elapsed := time.Now().Sub(st)
	if o == "" {
		fl, ln := dbg.ErrWasAt()
		o = fmt.Sprintf("%s @ %d", fl, ln)
	}
	if f {
		Failed(t, o, "%15s: %s", elapsed.String(), o)
	} else {
		Passed(t, o, "%15s: %s", elapsed.String(), o)
	}
	return f
}

// do an MD5 sum on the given file and compare it agains the list of possible results
func Md5SumFile(in string, cmp ...string) bool { // returns FAILED condition
	fl, err := os.Open(in)
	if dbg.ChkErr(err, "File `%s` not found.", in) {
		return true
	}
	defer fl.Close()

	sum := md5.New()
	_, err = io.Copy(sum, fl)
	dbg.ChkErrX(err)

	chk := fmt.Sprintf("%x", sum.Sum(nil))
	failed := true

	for _, v := range cmp {
		if v == chk {
			failed = false
			break
		}
	}

	if failed {
		fnm, ln := dbg.ErrWasAt()
		out := fmt.Sprintf(magentaClr+"MD5CHK @ %d in %s: "+normClr+"%s got( "+redClr+"%s"+normClr+" )", ln, fnm, in, chk)
		for _, v := range cmp {
			out += fmt.Sprintf("\n   exp( "+greenClr+"%s"+normClr+" )", v)
		}
		if 1 == len(cmp) {
			out += "      -- endianess issue?"
		}
		fmt.Println(out)
	}

	return failed
}

// default init/fini function (do nothing)
func stubFunc() {}

// enable color output if Linux machine
func init() {
	if env.IsLinux() {
		normClr = "\033[0m"      // reset to normal text
		testClr = "\033[97;44m"  // BRIGHT WHITE on BLUE
		failClr = "\033[30;41m"  // BLACK on RED
		passClr = "\033[30;42m"  // BLACK on GREEN
		skipClr = "\033[93;100m" // YELLOW on GREY
		redClr = "\033[31m"      // simple red text
		greenClr = "\033[32m"    // simple green text
		magentaClr = "\033[35m"  // simple magenta text
	}
	initFunc = stubFunc // set initFunc to the stub function
	finiFunc = stubFunc // set finiFunc to the stub function
}
