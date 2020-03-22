package models

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTreeGetNode(t *testing.T) {

	tree1json := `{
		"root": {
			"UUID": 840,
			"maxChilds": 2,
			"maxCounts": 32,
			"children": [
				{
					"UUID": 630,
					"maxChilds": 4,
					"maxCounts": 16,
					"children": [
						{
							"UUID": 878,
							"maxChilds": 4,
							"maxCounts": 4,
							"children": [],
							"txmatrix": []
						},
						{
							"UUID": 636,
							"maxChilds": 4,
							"maxCounts": 4,
							"children": [],
							"txmatrix": []
						},
						{
							"UUID": 407,
							"maxChilds": 4,
							"maxCounts": 4,
							"children": [],
							"txmatrix": []
						},
						{
							"UUID": 983,
							"maxChilds": 4,
							"maxCounts": 4,
							"children": [],
							"txmatrix": []
						}
					],
					"txmatrix": []
				},
				{
					"UUID": 203,
					"maxChilds": 4,
					"maxCounts": 16,
					"children": [
						{
							"UUID": 506,
							"maxChilds": 4,
							"maxCounts": 4,
							"children": [],
							"txmatrix": []
						},
						{
							"UUID": 20,
							"maxChilds": 4,
							"maxCounts": 4,
							"children": [],
							"txmatrix": []
						},
						{
							"UUID": 914,
							"maxChilds": 4,
							"maxCounts": 4,
							"children": [],
							"txmatrix": []
						},
						{
							"UUID": 272,
							"maxChilds": 4,
							"maxCounts": 4,
							"children": [],
							"txmatrix": []
						}
					],
					"txmatrix": []
				}
			],
			"txmatrix": []
		}
	}`
	var tree1 PeriodTree
	json.Unmarshal([]byte(tree1json), &tree1)

	Convey("Should navigate through nodes correctly", t, func() {
		// level 0
		So(
			tree1.GetNode([]int{}).UUID,
			ShouldEqual,
			840,
		)
		So(
			tree1.GetNode([]int{0}).UUID,
			ShouldEqual,
			630,
		)
		So(
			tree1.GetNode([]int{1}).UUID,
			ShouldEqual,
			203,
		)
		// level 1
		So(
			tree1.GetNode([]int{0, 0}).UUID,
			ShouldEqual,
			878,
		)
		So(
			tree1.GetNode([]int{0, 1}).UUID,
			ShouldEqual,
			636,
		)
		So(
			tree1.GetNode([]int{0, 2}).UUID,
			ShouldEqual,
			407,
		)
		So(
			tree1.GetNode([]int{0, 3}).UUID,
			ShouldEqual,
			983,
		)
		//
		So(
			tree1.GetNode([]int{1, 0}).UUID,
			ShouldEqual,
			506,
		)
		So(
			tree1.GetNode([]int{1, 1}).UUID,
			ShouldEqual,
			20,
		)
		So(
			tree1.GetNode([]int{1, 2}).UUID,
			ShouldEqual,
			914,
		)
		So(
			tree1.GetNode([]int{1, 3}).UUID,
			ShouldEqual,
			272,
		)
		// level 2: on TxMatrix
		So(
			tree1.GetNode([]int{0, 0, 0}).UUID,
			ShouldEqual,
			878,
		)
		So(
			tree1.GetNode([]int{0, 1, 0}).UUID,
			ShouldEqual,
			636,
		)
	})

}
