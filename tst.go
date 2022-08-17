package tst

import (
	"bytes"
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jayacarlson/dbg"
	"github.com/jayacarlson/env"
	"github.com/jayacarlson/hex"
)

/*
	Some functions useful for testing routines
*/

// Tests use this type, returning the name of the test and the result (Failed as TRUE)
type (
	TestFunc  func() (testName, description string, testResult bool)
	TestTFunc func(t *testing.T) (testName, description string, testResult bool)

	Chk      struct{ failed bool }
	TestOpts struct{ ShowPassed bool }
)

var (
	initFunc                  func() // setable function to run before each test
	finiFunc                  func() // setable function to run after each test
	normClr, testClr, passClr string
	failClr, skipClr, redClr  string
	greenClr, magentaClr      string

	Options = TestOpts{ShowPassed: false}
)

func init() {
	// individual test.go files will need to do "flag.Parse()" to enable this, or they
	// can alter Options.ShowPassed directly
	flag.BoolVar(&Options.ShowPassed, "tst.ShowPassed", false, "Output PASSED messages")
}

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
		fmt.Printf(testClr+" Testing: "+normClr+" %-20s %s\n", testName, description)
	} else {
		fmt.Printf(skipClr+" Disabled "+normClr+" %-20s %s\n", testName, description)
	}
	return enabled
}

// Show PASSED message
func Passed(t *testing.T, _ string, fstr string, a ...interface{}) {
	if Options.ShowPassed {
		fmt.Printf(passClr+"  Passed  "+normClr+" "+fstr+"\n", a...)
	}
}

// Show FAILED message
func Failed(t *testing.T, who string, fstr string, a ...interface{}) {
	fmt.Printf(failClr+"  Failed  "+normClr+" "+fstr+"\n", a...)
	t.Errorf("%s", who)
}

// call the given 'test' function and output whether it FAILED
func Func(t *testing.T, test TestFunc) bool {
	initFunc()
	defer finiFunc()
	st := time.Now()
	o, d, f := test()
	elapsed := time.Now().Sub(st)
	if o == "" {
		fl, ln := dbg.ErrWasAt()
		o = fmt.Sprintf("%s @ %d", fl, ln)
	}
	if f {
		Failed(t, o, "%15s: %s", elapsed.String(), o)
	} else {
		Passed(t, o, "%15s: %s %s", elapsed.String(), o, d)
	}
	return f
}

// call the given 'test' function and output whether it FAILED
func FuncT(t *testing.T, test TestTFunc) bool {
	initFunc()
	defer finiFunc()
	st := time.Now()
	o, d, f := test(t)
	elapsed := time.Now().Sub(st)
	if o == "" {
		fl, ln := dbg.ErrWasAt()
		o = fmt.Sprintf("%s @ %d", fl, ln)
	}
	if f {
		Failed(t, o, "%15s: %s", elapsed.String(), o)
	} else {
		Passed(t, o, "%15s: %s %s", elapsed.String(), o, d)
	}
	return f
}

// Returns "FAILED" as true to keep consistent with dbg.Chk
func Bool(t *testing.T, w string, passed bool) bool {
	if passed {
		Passed(t, w, w)
	} else {
		Failed(t, w, w)
	}
	return !passed
}

// Returns "FAILED" as true to keep consistent with dbg.Chk
func Bin(t *testing.T, w string, a, b []byte) bool {
	return Bool(t, w, 0 == bytes.Compare(a, b))
}

// Returns "FAILED" as true to keep consistent with dbg.Chk
func Panic(t *testing.T, w string, f func()) bool {
	return Bool(t, w, !panicTest(f))
}

func Slice(t *testing.T, w string, a, b interface{}) bool {
	ev := reflect.ValueOf(a)
	vv := reflect.ValueOf(b)
	ek := ev.Kind()
	vk := vv.Kind()
	et := ev.Type()
	vt := vv.Type()

	if ek != vk || reflect.Slice != ek {
		Failed(t, w+": mismatch in kind or not slices", w+": mismatch in kind or not slices")
		return true
	}
	if et != vt {
		Failed(t, w+": mismatch in types", w+": mismatch in types")
		return true
	}

	c := ev.Cap()
	if c == vv.Cap() {
		failed := false
		{
			aa, aok := a.([]byte)
			bb, bok := b.([]byte)
			if aok && bok {
				for i := 0; i < c && !failed; i++ {
					failed = failed || (aa[i] != bb[i])
				}
				return Bool(t, w, !failed)
			}
		}
		{
			aa, aok := a.([]int)
			bb, bok := b.([]int)
			if aok && bok {
				for i := 0; i < c && !failed; i++ {
					failed = failed || (aa[i] != bb[i])
				}
				return Bool(t, w, !failed)
			}
		}
		{
			aa, aok := a.([]float32)
			bb, bok := b.([]float32)
			if aok && bok {
				for i := 0; i < c && !failed; i++ {
					failed = failed || (aa[i] != bb[i])
				}
				return Bool(t, w, !failed)
			}
		}
		{
			aa, aok := a.([]float64)
			bb, bok := b.([]float64)
			if aok && bok {
				for i := 0; i < c && !failed; i++ {
					failed = failed || (aa[i] != bb[i])
				}
				return Bool(t, w, !failed)
			}
		}
		{
			aa, aok := a.([]int32)
			bb, bok := b.([]int32)
			if aok && bok {
				for i := 0; i < c && !failed; i++ {
					failed = failed || (aa[i] != bb[i])
				}
				return Bool(t, w, !failed)
			}
		}
		{
			aa, aok := a.([]uint32)
			bb, bok := b.([]uint32)
			if aok && bok {
				for i := 0; i < c && !failed; i++ {
					failed = failed || (aa[i] != bb[i])
				}
				return Bool(t, w, !failed)
			}
		}
		Failed(t, w+": invalid slice type", w+": invalid slice type")
		return true
	} else {
		Failed(t, w+": mismatch in sizes", w+": mismatch in sizes")
		return true
	}
	Passed(t, w, w)
	return false
}

// A way to do tests that require multiple prep steps with validation
func (z *Chk) Reset()   { z.failed = false }
func (z *Chk) Failed()  { z.failed = true }  // simply signal failure
func (z *Chk) Ok() bool { return !z.failed } // return "hasn't failed yet"

// Returns "FAILED" as true to keep consistent with dbg.Chk
func (z *Chk) Tru(b bool, a ...interface{}) bool {
	c := func() {}
	if !b {
		// see if last arg is a closer func -- func()
		if len(a) > 0 {
			if cl, ok := a[len(a)-1].(func()); ok {
				// pull off the closer
				c = cl
				a = a[:len(a)-1]
			}
		}
		if len(a) > 0 {
			if f, ok := a[0].(string); ok {
				dbg.Error(f, a[1:]...)
			} else {
				dbg.Fatal("bad arg(s) for chk.Tru")
			}
		}
		c()
		z.failed = true
	}
	return !b
}

// Returns "FAILED" as true to keep consistent with dbg.Chk
func (z *Chk) Err(err error, a ...interface{}) bool {
	c := func() {}
	if nil != err {
		// see if last arg is a closer func -- func()
		if len(a) > 0 {
			if cl, ok := a[len(a)-1].(func()); ok {
				// pull off the closer
				c = cl
				a = a[:len(a)-1]
			}
		}
		if len(a) > 0 {
			if f, ok := a[0].(string); ok {
				dbg.Error(f, a[1:]...)
			} else {
				dbg.Fatal("bad arg(s) for chk.Err")
			}
		}
		c()
		z.failed = true
	}
	return nil != err
}

// Returns "FAILED" as true to keep consistent with dbg.Chk
func (z *Chk) ErrIs(err, cmp error) bool {
	if cmp != err {
		dbg.Error("Received error: %v", err)
		dbg.Warning("Should have received: %v", cmp)
		z.failed = true
	}
	return cmp != err
}

// Show pass/fail, returns the failed flag
func (z Chk) ShowPassFail(t *testing.T, who string) bool {
	if z.failed {
		Failed(t, who, who)
	} else {
		Passed(t, who, who)
	}
	return z.failed
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

func CmpBytes(t *testing.T, desc string, a, b []byte) bool {
	who := dbg.IWas()
	failed := 0 != bytes.Compare(a, b)
	if failed {
		Failed(t, who, who)
		fmt.Print(greenClr)
		hex.DumpC(a)
		fmt.Print(redClr)
		hex.DumpC(b)
		fmt.Print(normClr)
	} else {
		Passed(t, who, who)
	}
	return failed
}

func AsRed(src string) {
	for len(src) > 0 {
		if i := strings.Index(src, "\n"); i >= 0 {
			fmt.Printf(failClr + src[:i] + normClr + "\n")
			src = src[i+1:]
		} else {
			fmt.Printf(failClr + src + normClr + "\n")
			return
		}
	}
}

func AsYellow(src string) {
	for len(src) > 0 {
		if i := strings.Index(src, "\n"); i >= 0 {
			fmt.Printf(skipClr + src[:i] + normClr + "\n")
			src = src[i+1:]
		} else {
			fmt.Printf(skipClr + src + normClr + "\n")
			return
		}
	}
}

func AsGreen(src string) {
	for len(src) > 0 {
		if i := strings.Index(src, "\n"); i >= 0 {
			fmt.Printf(passClr + src[:i] + normClr + "\n")
			src = src[i+1:]
		} else {
			fmt.Printf(passClr + src + normClr + "\n")
			return
		}
	}
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

func panicTest(f func()) bool {
	defer func() { _ = recover() }() // a panic returns FALSE
	f()
	return true // not panicing will return TRUE
}
