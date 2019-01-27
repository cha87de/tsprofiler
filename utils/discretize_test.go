package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDiscretize(t *testing.T) {
	Convey("Should discretize corret state", t, func() {
		So(discretize(24, 4, 0, 100).value, ShouldEqual, 0)
		So(discretize(55, 4, 0, 100).value, ShouldEqual, 2)
		So(discretize(70, 2, 0, 100).value, ShouldEqual, 1)
		So(discretize(0, 4, 0, 0).value, ShouldEqual, 0)
	})
}
