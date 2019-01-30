package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDiscretize(t *testing.T) {
	Convey("Should discretize corret state", t, func() {
		So(Discretize(24, 4, 0, 100).Value, ShouldEqual, 0)
		So(Discretize(55, 4, 0, 100).Value, ShouldEqual, 2)
		So(Discretize(70, 2, 0, 100).Value, ShouldEqual, 1)
		So(Discretize(0, 4, 0, 0).Value, ShouldEqual, 0)
	})
}
