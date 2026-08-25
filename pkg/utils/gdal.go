package utils

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Check GDAL tools available checks if gdalbuildvrt is available in the environment.
func CheckGDALToolAvailable(tool string) bool {
	cmd := exec.Command(tool, "--version")
	if out, err := cmd.Output(); err != nil {
		outStr := string(out)
		return strings.Contains(strings.ToLower(outStr), "released") // can't use exit code because https://github.com/OSGeo/gdal/issues/11550
	}
	return true
}

func CreateTempVRT(inputFileListPath, absOutputPath string) (string, error) {

	// Create intermediate directories if they do not exist
	if err := os.MkdirAll(filepath.Dir(absOutputPath), 0755); err != nil {
		return "", fmt.Errorf("could not create directories for %s: %v", absOutputPath, err)
	}

	// We don't really need os.CreateTemp, but we are using it to generate random file name
	tempVRTFile, err := os.CreateTemp(filepath.Dir(absOutputPath), "~f2f_*.tmp")
	if err != nil {
		return "", fmt.Errorf("error creating temporary VRT file: %v", err)
	}
	tempVRTPath := tempVRTFile.Name()
	tempVRTFile.Close() // Close now, gdalbuildvrt will write to it

	// Build temporary VRT file
	vrtArgs := []string{"-input_file_list", inputFileListPath, tempVRTPath}
	vrtCmd := exec.Command("gdalbuildvrt", vrtArgs...)
	vrtCmd.Stdout = os.Stdout
	vrtCmd.Stderr = os.Stderr

	slog.Debug("Creating temporary VRT file",
		"command", fmt.Sprintf("gdalbuildvrt %s", strings.Join(vrtArgs, " ")),
		"tempVRT", tempVRTPath,
	)

	if err := vrtCmd.Run(); err != nil {
		return "", fmt.Errorf("error running gdalbuildvrt: %v", err)
	}

	return tempVRTPath, nil
}
