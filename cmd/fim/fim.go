package fim

import (
	"encoding/csv"
	"flag"
	"flows2fim/pkg/utils"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var usage string = `Usage of fim:
Given a control table and a fim library folder, create a composite flood inundation map for the control conditions.
GDAL VSI paths can be used (only for library and not for output), given GDAL must have access to cloud creds.

FIM Library Specifications:
- All maps should have same CRS, Resolution, data type, vertical units (if any), and nodata value
- Should have following folder structure:
.
├── 2821866
│   ├── z_nd
│   │   ├── f_10283.tif
│   │   ├── f_104569.tif
│   │   ├── f_11199.tif
│   │   ├── f_112807.tif
│   │   ...
│   ├── z_53_5
│   │   ├── f_102921.tif
│   │   ├── f_10485.tif
│   │   ├── f_111159.tif
│   │   ├── f_11309.tif
│   │   ...
│   ...
│   └── domain.tif (optional)
├── 2821867
│   ├── z_nd
...

Arguments:` // Usage should be always followed by PrintDefaults()

func Run(args []string) (gdalArgs []string, err error) {
	flags := flag.NewFlagSet("fim", flag.ExitOnError)
	flags.Usage = func() {
		fmt.Println(usage)
		flags.PrintDefaults()
	}

	var controlsFile, fimLibDir, libType, outputFormat, outputFile string
	var withDomain bool

	// Define flags using flags.StringVar
	flags.StringVar(&fimLibDir, "lib", "", "Directory containing FIM Library. GDAL VSI paths can be used, given GDAL must have access to cloud creds")
	flags.StringVar(&controlsFile, "c", "", "Path to the controls CSV file")
	flags.StringVar(&outputFormat, "fmt", "VRT", "Output format: 'VRT', 'COG' or 'GTIFF'") // follows GDAL format names, case insensitive
	flags.StringVar(&libType, "type", "", "Library type: 'depth' or 'extent'")             // was only required for v0.3.0, but keeping it for backward compatibility
	flags.StringVar(&outputFile, "o", "", "Output FIM file path")
	flags.BoolVar(&withDomain, "with_domain", false, "If true, domain is added behind FIMs")

	// Parse flags from the arguments
	if err := flags.Parse(args); err != nil {
		return []string{}, fmt.Errorf("error parsing flags: %v", err)
	}

	outputFormat = strings.ToUpper(outputFormat) // COG, cog, VRT, vrt all okay

	// Validate required flags
	if controlsFile == "" || fimLibDir == "" || outputFile == "" {
		fmt.Println(controlsFile, fimLibDir, outputFile)
		fmt.Println("Missing required flags")
		flags.PrintDefaults()
		return []string{}, fmt.Errorf("missing required flags")
	}

	// Check if required GDAL tools are available
	requiredTools := []string{"gdalbuildvrt"}
	if outputFormat != "VRT" {
		requiredTools = append(requiredTools, "gdal_translate")
	}

	for _, tool := range requiredTools {
		if !utils.CheckGDALToolAvailable(tool) {
			slog.Error("GDAL tool missing", "tool", tool)
			return []string{}, fmt.Errorf("%[1]s is not available. Please install GDAL and ensure %[1]s is in your PATH", tool)
		}
	}

	var absOutputPath, absFimLibPath string
	if strings.HasPrefix(outputFile, "/vsi") {
		absOutputPath = outputFile
	} else {
		absOutputPath, err = filepath.Abs(outputFile)
		if err != nil {
			return []string{}, fmt.Errorf("error getting absolute path for output file: %v", err)
		}
	}

	if strings.HasPrefix(fimLibDir, "/vsi") {
		absFimLibPath = fimLibDir
	} else {
		absFimLibPath, err = filepath.Abs(fimLibDir)
		if err != nil {
			return []string{}, fmt.Errorf("error getting absolute path for FIM library directory: %v", err)
		}
	}

	// Processing CSV
	file, err := os.Open(controlsFile)
	if err != nil {
		return []string{}, fmt.Errorf("error opening controls file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return []string{}, fmt.Errorf("error reading CSV file: %v", err)
	}

	if len(records) < 2 {
		return []string{}, fmt.Errorf("no records in controls file")
	}

	if len(records[1]) < 3 {
		return []string{}, fmt.Errorf("not enough columns in controls file, need at least 3")
	}

	// Locate map_exists by header name rather than position, so additional
	// columns can be added to the controls file later without breaking this.
	mapExistsIdx := -1
	for i, name := range records[0] {
		if strings.TrimSpace(name) == "map_exists" {
			mapExistsIdx = i
			break
		}
	}
	if mapExistsIdx == -1 {
		slog.Warn("controls file has no map_exists column; assuming every record has a map", "file", controlsFile)
	}

	var domainFiles, fimFiles []string
	var skipped int
	for _, record := range records[1:] { // Skip header row
		reachID := record[0]

		// Records without a depth grid are hydraulically correct but have no raster
		// to composite, so they are left out of the VRT rather than pointing at a
		// file that does not exist.
		if mapExistsIdx != -1 && mapExistsIdx < len(record) {
			if mapExists := strings.TrimSpace(record[mapExistsIdx]); mapExists == "0" || strings.EqualFold(mapExists, "false") {
				slog.Debug("Skipping record with no map", "reach_id", reachID, "flow", record[1], "control_stage", record[2])
				skipped++
				continue
			}
		}

		record[2] = strings.Replace(record[2], ".", "_", -1) // Replace '.' with '_'
		folderPath := filepath.Join(absFimLibPath, reachID, fmt.Sprintf("z_%s", record[2]))
		fileName := fmt.Sprintf("f_%s.tif", record[1])
		absFIMPath := filepath.Join(folderPath, fileName)
		absDomainPath := filepath.Join(absFimLibPath, reachID, "domain.tif")

		// join on windows may cause \vsi
		if strings.HasPrefix(absFIMPath, `\vsi`) {
			absDomainPath = strings.ReplaceAll(absDomainPath, `\`, "/")
			absFIMPath = strings.ReplaceAll(absFIMPath, `\`, "/")
		}

		fimFiles = append(fimFiles, absFIMPath)
		if withDomain {
			domainFiles = append(domainFiles, absDomainPath)
		}
	}

	if skipped > 0 {
		slog.Debug("Excluded records with no map from the composite", "count", skipped, "included", len(fimFiles))
	}
	if len(fimFiles) == 0 {
		return []string{}, fmt.Errorf("no records in controls file have a map; nothing to composite")
	}

	// Write file paths to a temporary file
	inputFileListPath, err := utils.WriteListToTempFile(append(domainFiles, fimFiles...))
	if err != nil {
		return []string{}, fmt.Errorf("error writing file list to temporary file: %v", err)
	}
	defer os.Remove(inputFileListPath)

	tempVRTPath, err := utils.CreateTempVRT(inputFileListPath, absOutputPath)
	if err != nil {
		return []string{}, fmt.Errorf("error creating temp vrt: %v", err)
	}
	defer os.Remove(tempVRTPath)

	if outputFormat == "VRT" {
		// For VRT, simply move the temporary file to the final destination for atomicity
		slog.Debug("Moving temporary VRT to final destination",
			"from", tempVRTPath,
			"to", absOutputPath)

		if err := os.Rename(tempVRTPath, absOutputPath); err != nil {
			return []string{}, fmt.Errorf("error renaming temp file %s to %s: %v", tempVRTPath, absOutputPath, err)
		}

	} else {
		// For TIF or COG, use gdal_translate to convert the VRT
		translateArgs := []string{
			"-co", "COMPRESS=LZW",
			"-co", "NUM_THREADS=ALL_CPUS",
			"-of", outputFormat,
			tempVRTPath,
			absOutputPath,
		}

		translateCmd := exec.Command("gdal_translate", translateArgs...)
		translateCmd.Stdout = os.Stdout
		translateCmd.Stderr = os.Stderr

		slog.Debug(fmt.Sprintf("Converting VRT to %s", outputFormat),
			"command", fmt.Sprintf("gdal_translate %s", strings.Join(translateArgs, " ")),
			"format", outputFormat,
		)

		if err := translateCmd.Run(); err != nil {
			return []string{}, fmt.Errorf("error converting VRT to %s: %v", outputFormat, err)
		}

	}

	fmt.Printf("Composite FIM created at %s\n", absOutputPath)

	return gdalArgs, nil
}
