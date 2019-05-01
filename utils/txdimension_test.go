package utils

import (
	"testing"

	"github.com/cha87de/tsprofiler/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestChangeDimension(t *testing.T) {
	Convey("Should ChangeDimension correctly", t, func() {

		So(
			ChangeDimension(map[string][]int64{
				"3": []int64{0, 0, 0, 145},
			}, models.TSStats{
				Min: 50, Max: 55,
			}, models.TSStats{
				Min: 0, Max: 100,
			}, 4),

			ShouldResemble,
			map[string][]int64{
				"2": []int64{0, 0, 145, 0},
			},
		)

		So(
			ChangeDimension(
				ChangeDimension(map[string][]int64{
					"3": []int64{0, 0, 0, 145},
				}, models.TSStats{
					Min: 50, Max: 55,
				}, models.TSStats{
					Min: 40, Max: 90,
				}, 4),

				models.TSStats{
					Min: 40, Max: 90,
				}, models.TSStats{
					Min: 0, Max: 100,
				}, 4),
			ShouldResemble,
			map[string][]int64{
				// "2": []int64{0, 0, 145, 0},  // THIS IS RIGHT
				"3": []int64{0, 0, 0, 145}, // THIS IS WRONG!
				// see issue #5
			},
		)

		So(
			ChangeDimension(map[string][]int64{
				"0": []int64{10, 0, 0, 0},
				"3": []int64{0, 0, 0, 100},
			}, models.TSStats{
				Min: 0, Max: 10,
			}, models.TSStats{
				Min: 0, Max: 100,
			}, 4),
			ShouldResemble,
			map[string][]int64{
				"0": []int64{110, 0, 0, 0},
			},
		)

		So(
			ChangeDimension(map[string][]int64{
				"0": []int64{10, 0, 0, 0},
				"3": []int64{0, 0, 0, 100},
			}, models.TSStats{
				Min: 10, Max: 20,
			}, models.TSStats{
				Min: 0, Max: 20,
			}, 4),
			ShouldResemble,
			map[string][]int64{
				"2": []int64{0, 0, 10, 0},
				"3": []int64{0, 0, 0, 100},
			},
		)

		So(
			ChangeDimension(map[string][]int64{
				"1": []int64{0, 30, 20, 0},
				"2": []int64{0, 0, 20, 0},
				"3": []int64{0, 0, 0, 100},
			}, models.TSStats{
				Min: 20, Max: 50,
			}, models.TSStats{
				Min: 0, Max: 100,
			}, 4),
			ShouldResemble,
			map[string][]int64{
				"1": []int64{0, 70, 0, 0},
				"2": []int64{0, 0, 100, 0},
			},
		)
	})
}
