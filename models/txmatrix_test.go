package models

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDiff(t *testing.T) {

	tx1s := `{
		"transitions": {
			"0": {
				"nextProbs": [
					0,
					100,
					0,
					0
				]
			},
			"1": {
				"nextProbs": [
					0,
					0,
					100,
					0
				]
			},
			"2": {
				"nextProbs": [
					0,
					100,
					0,
					0
				]
			},
			"3": {
				"nextProbs": [
					100,
					0,
					0,
					0
				]
			}
		}
	}`
	var tx1 TxMatrix
	json.Unmarshal([]byte(tx1s), &tx1)

	tx2s := `{
		"transitions": {
			"0": {
				"nextProbs": [
					100,
					0,
					0,
					0
				]
			},
			"1": {
				"nextProbs": [
					0,
					0,
					0,
					100
				]
			},
			"2": {
				"nextProbs": [
					50,
					0,
					0,
					50
				]
			},
			"3": {
				"nextProbs": [
					0,
					0,
					0,
					100
				]
			}
		}
	}`
	var tx2 TxMatrix
	json.Unmarshal([]byte(tx2s), &tx2)

	tx3s := `{
		"transitions": {
			"0": {
				"nextProbs": [
					50,
					20,
					30,
					0
				]
			},
			"1": {
				"nextProbs": [
					0,
					0,
					50,
					50
				]
			},
			"2": {
				"nextProbs": [
					100,
					0,
					0,
					0
				]
			},
			"3": {
				"nextProbs": [
					50,
					50,
					0,
					0
				]
			}
		}
	}`
	var tx3 TxMatrix
	json.Unmarshal([]byte(tx3s), &tx3)

	Convey("Should compute diff between two TxMatrizes correctly", t, func() {
		So(
			tx1.Diff(tx1),
			ShouldEqual,
			1.0,
		)
		So(
			tx1.Diff(tx2),
			ShouldEqual,
			0.0,
		)
		So(
			tx1.Diff(tx3),
			ShouldEqual,
			0.3,
		)
	})

}
