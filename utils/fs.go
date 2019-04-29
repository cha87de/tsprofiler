package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cha87de/tsprofiler/models"
)

// ReadProfileFromFile returns the TSProfile model read from the given json file
func ReadProfileFromFile(filepath string) models.TSProfile {
	filehandler, err := os.Open(filepath)
	if err != nil {
		fmt.Println(err)
	}
	defer filehandler.Close()
	byteValue, _ := ioutil.ReadAll(filehandler)

	var profile models.TSProfile
	json.Unmarshal(byteValue, &profile)

	return profile
}
