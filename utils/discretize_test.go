package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSimpleDiscretize(t *testing.T) {
	Convey("Should SimpleDiscretize corret state", t, func() {
		So(SimpleDiscretize(24, 4, 0, 100).Value, ShouldEqual, 0)
		So(SimpleDiscretize(55, 4, 0, 100).Value, ShouldEqual, 2)
		So(SimpleDiscretize(70, 2, 0, 100).Value, ShouldEqual, 1)
		So(SimpleDiscretize(0, 4, 0, 0).Value, ShouldEqual, 0)
	})
}

func TestClosestDiscretize(t *testing.T) {
	Convey("Should SimpleDiscretize corret state", t, func() {
		So(ClosestDiscretize(24, 4, 0, 100).Value, ShouldEqual, 1)
		So(ClosestDiscretize(55, 4, 0, 100).Value, ShouldEqual, 2)
		So(ClosestDiscretize(70, 2, 0, 100).Value, ShouldEqual, 1)
		//So(ClosestDiscretize(0, 4, 0, 0).Value, ShouldEqual, 3)
		So(ClosestDiscretize(91, 4, 0, 100).Value, ShouldEqual, 3)
	})
}
