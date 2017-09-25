package ginkgoext

import (
	"github.com/onsi/ginkgo"
)

type AfterAll struct {
	Task     int
	Executed int
	Focused  int
	Body     func()
}

//AddTestType: This function used the ginkgo test function (IT, FIT) and wrap
//the body function to mark as executed or not. The idea behind this is to set
//a AfterAll struct and execute a function when the context/Describe is
//finished. There is no better way to do this, see ginkgo issue 70 to read
//more about this.
func addTestType(text string, body func(), testType string, control *AfterAll, timeout ...float64) bool {
	var ginkgoFunc func(text string, body interface{}, timeout ...float64) bool

	//FIXME: XIT functions
	switch testType {
	case "FIT":
		control.Focused++
		ginkgoFunc = ginkgo.FIt
	default:
		control.Task++
		ginkgoFunc = ginkgo.It
	}
	wrappedBody := func() {
		body()
		control.Executed++
		if (control.Executed == control.Task) || (control.Focused > 0 && control.Focused == control.Executed) {
			control.Body()
		}
	}
	return ginkgoFunc(text, wrappedBody, timeout...)
}

//Text block with AfterAll wrapper
func It(text string, body func(), afterAll *AfterAll, timeout ...float64) bool {
	return addTestType(text, body, "IT", afterAll, timeout...)
}

//You can focus individual Its using FIt. this contains AfterAll wrapper
func FIt(text string, body func(), afterAll *AfterAll, timeout ...float64) bool {
	return addTestType(text, body, "FIT", afterAll, timeout...)
}
