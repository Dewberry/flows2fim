package validate

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/csv"
	"flag"
	"flows2fim/pkg/utils"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	_ "modernc.org/sqlite"
)

var usage string = `Usage of validate:
Given a fim library folder and a scenarios database,
validate there is one to one correspondence between the entries of scenarios table and fim library objects.
GDAL VSI paths can be used, given GDAL must have access to cloud creds.
Intermediate folders for output files are created if they do not exist.

Correspondence is judged by matching each scenario's fim_path against the .tif files found
in the library. Only scenarios with map_exists=1 carry a fim_path. Two disagreements are
reported, each to its own CSV:
        o_fims             scenarios with map_exists=1 whose fim_path has no file (or is empty)
        o_scenarios        FIM files whose path is not the fim_path of any scenario
Scenarios with map_exists=0 and no FIM file are the expected case and are not reported.

FIM Library Specifications:
- All maps should have same CRS, Resolution, vertical units (if any), and nodata value
- fim_path values are relative to the library root and use '/' as separator
- <reach_id>/domain.tif files are reach domains, not FIMs, and are ignored

Database file must have a table 'scenarios' and contain following columns
        reach_id INTEGER
        us_flow REAL
        us_depth REAL
        us_wse Real
        ds_depth REAL
        ds_wse REAL
        boundary_condition TEXT CHECK(boundary_condition IN ('nd','kwse'))
        map_exists BOOL CHECK(map_exists IN (0, 1))
        fim_path TEXT
        UNIQUE(reach_id, us_flow, ds_wse, boundary_condition)


Arguments:`

// SQL Query Constants
const (
	queryCreateFIMEntTable = `
	CREATE TABLE memdb.fim_entries (
		fim_path TEXT PRIMARY KEY
	);
	`

	// Only scenarios whose depth grid was written are expected to have a FIM.
	queryMissingFims = `
	SELECT
		s.reach_id,
		s.us_flow,
		printf('%.1f', s.ds_wse) AS ds_wse,
		s.boundary_condition,
		COALESCE(s.fim_path, '') AS fim_path
	FROM
		scenarios s
	LEFT JOIN
		memdb.fim_entries f
		ON s.fim_path = f.fim_path
	WHERE
		f.fim_path IS NULL
		AND s.map_exists = 1
	ORDER BY
		s.reach_id, s.boundary_condition, s.ds_wse, s.us_flow;
	`

	queryMissingScenarios = `
	SELECT
		f.fim_path
	FROM
		memdb.fim_entries f
	LEFT JOIN
		scenarios s
		ON s.fim_path = f.fim_path
	WHERE
		s.fim_path IS NULL
	ORDER BY
		f.fim_path;
	`
)

var extIgnore = []string{".aux", ".aux.xml", ".ovr", ".xml", ".tfw"}

// dirEntry holds a path + info about whether it's a directory
type dirEntry struct {
	path  string
	isDir bool
}

// readDir is the wrapper that calls either gatherLocalEntries or gatherCloudEntries
// to get all paths (files + dirs).
// If recursive is true, it will recursively list all files and directories.
func readDir(dir string, recursive bool) ([]dirEntry, error) {
	var allEntries []dirEntry
	var err error

	if strings.HasPrefix(dir, "/vsi") {
		allEntries, err = gatherVSIEntries(dir, recursive)
	} else {
		allEntries, err = gatherLocalEntries(dir, recursive)
	}
	if err != nil {
		return nil, fmt.Errorf("error gathering entries from %s: %v", dir, err)
	}

	return allEntries, nil
}

// gatherLocalEntries uses either os.ReadDir (non-recursive) or filepath.WalkDir (recursive)
func gatherLocalEntries(dir string, recursive bool) ([]dirEntry, error) {
	if !recursive {
		// Non-recursive approach: just top-level entries
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		var results []dirEntry
		for _, e := range entries {
			results = append(results, dirEntry{
				path:  filepath.Join(dir, e.Name()),
				isDir: e.IsDir(),
			})
		}
		return results, nil
	}

	// Recursive approach with WalkDir
	var results []dirEntry
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		results = append(results, dirEntry{path: path, isDir: d.IsDir()})
		return nil
	})
	return results, err
}

// gatherVSIEntries calls gdal_ls (with or without -r) to list entries in a VSI path
func gatherVSIEntries(dir string, recursive bool) ([]dirEntry, error) {
	var args []string
	if recursive {
		args = []string{"-r", dir}
	} else {
		args = []string{dir}
	}

	cmd := exec.Command(gdalLSName, args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error gathering entries from %s: %v", dir, err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		slog.Warn("Error reading gdal_ls output", "tool", gdalLSName, "dir", dir, "error", err)
	}

	var results []dirEntry
	for _, line := range lines {
		if line == "" || !strings.HasPrefix(line, "/") { // ignore lines not starting with /
			continue
		}
		isDir := strings.HasSuffix(line, "/")
		results = append(results, dirEntry{path: line, isDir: isDir})
	}
	return results, nil
}

// processLibEntry sends the library-relative, '/'-separated path of a FIM file to fimChan.
func processLibEntry(e dirEntry, absFimLibDir string, fimChan chan<- string) {
	// Skip directories
	if e.isDir {
		return
	}

	relPath, relErr := filepath.Rel(absFimLibDir, e.path)
	if relErr != nil {
		slog.Error("Relative path resolution failed", "path", e.path, "error", relErr)
		return
	}
	// fim_path values in the database always use forward slashes
	relPath = filepath.ToSlash(relPath)

	name := filepath.Base(e.path)
	ext := filepath.Ext(name)
	if utils.SliceContains(extIgnore, ext) {
		return
	} else if !strings.HasSuffix(name, ".tif") {
		slog.Warn("Non-TIFF file found", "path", relPath)
		return
	}

	if name == "domain.tif" && strings.Count(relPath, "/") == 1 {
		return
	}

	fimChan <- relPath
}

// batchInsertFIMs insert FIM paths in batches
func batchInsertFIMs(db *sql.DB, fimChan <-chan string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	const batchSize = 1000
	batch := make([]string, 0, batchSize)

	commitBatch := func() error {
		if len(batch) == 0 {
			return nil
		}

		// Can't do prepared statement as the final batch would not be of same size
		// Build single statement for multi-VALUES insert
		// INSERT OR IGNORE INTO memdb.fim_entries(fim_path) VALUES (?), (?) ...
		sqlStr := "INSERT OR IGNORE INTO memdb.fim_entries(fim_path) VALUES "
		vals := make([]interface{}, 0, len(batch))
		placeholders := make([]string, 0, len(batch))

		for _, p := range batch {
			placeholders = append(placeholders, "(?)")
			vals = append(vals, p)
		}
		sqlStr += strings.Join(placeholders, ",")

		if _, err := tx.Exec(sqlStr, vals...); err != nil {
			return err
		}
		batch = batch[:0]
		return nil
	}

	for row := range fimChan {
		batch = append(batch, row)
		if len(batch) >= batchSize {
			if err := commitBatch(); err != nil {
				_ = tx.Rollback()
				return err
			}
		}
	}

	// Do final flush when fimChan is closed in main routine
	if err := commitBatch(); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// writeCSV is a generic function to write results to CSV.
// It is atomic i.e. it either succeeds or no file is created.
// It create intermediate directories if they do not exist.
// It returns number of data rows written in CSV.
func writeCSV(rows *sql.Rows, outFile string, skipEmpty bool) (int, error) {
	// Try reading the first row to check if there's data. We don't want to create empty intermediate directories.
	// An approach that was first adopted and discarded was to write to a temp file in tmp folder and then rename it only if rows exist,
	// but that approach cause inter-device rename error on conatinerized environment with file mounts.

	// Columns must be read before Next, which closes rows when there are none
	cols, err := rows.Columns()
	if err != nil {
		return 0, fmt.Errorf("error reading columns: %v", err)
	}

	hasRow := rows.Next()
	if !hasRow {
		// No row was found if rows.Err() == nil.
		if err := rows.Err(); err != nil {
			return 0, fmt.Errorf("error reading rows: %v", err)
		}
		if skipEmpty {
			return 0, nil
		}
	}

	// Create intermediate directories if they do not exist
	if err := os.MkdirAll(filepath.Dir(outFile), 0755); err != nil {
		return 0, fmt.Errorf("could not create directories for %s: %v", outFile, err)
	}

	// On the same filesystem, os.Rename is atomic so will create a temp file and rename it later.
	tempFile, err := os.CreateTemp(filepath.Dir(outFile), "~f2f_*.tmp")
	if err != nil {
		return 0, fmt.Errorf("error creating temp file: %v", err)
	}
	tempFilePath := tempFile.Name()

	defer func() {
		_ = tempFile.Close()        // Always attempt to close file tempfile even if file is already closed
		_ = os.Remove(tempFilePath) // Always attempt to remove tempfile even if file is already renamed
	}()

	w := csv.NewWriter(tempFile)
	if err := w.Write(cols); err != nil {
		return 0, fmt.Errorf("error writing CSV header: %v", err)
	}

	vals := make([]sql.NullString, len(cols))
	ptrs := make([]interface{}, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	record := make([]string, len(cols))

	rowCount := 0
	// rows is already positioned on the first row, if there is one
	for ; hasRow; hasRow = rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return 0, fmt.Errorf("error scanning row: %v", err)
		}
		for i, v := range vals {
			record[i] = v.String
		}
		if err := w.Write(record); err != nil {
			return 0, fmt.Errorf("error writing CSV record: %v", err)
		}
		rowCount++
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("error from rows iteration: %v", err)
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return 0, fmt.Errorf("error flushing CSV: %v", err)
	}

	if err := tempFile.Close(); err != nil {
		return 0, fmt.Errorf("error closing temp file: %v", err)
	}

	if err := os.Rename(tempFilePath, outFile); err != nil {
		return 0, fmt.Errorf("error renaming temp file %s to %s: %v", tempFilePath, outFile, err)
	}
	return rowCount, nil
}

func Run(args []string) error {
	flags := flag.NewFlagSet("validate", flag.ExitOnError)
	flags.Usage = func() {
		fmt.Println(usage)
		flags.PrintDefaults()
	}

	var (
		dbPath       string
		fimLibDir    string
		outFims      string
		outScenarios string
		concurrent   int
		skipEmpty    bool
	)

	flags.StringVar(&dbPath, "db", "", "Path to the scenarios database file")
	flags.StringVar(&fimLibDir, "lib", "", "Path to the FIM library directory")
	flags.StringVar(&outFims, "o_fims", "missing_fims.csv", "Output CSV for scenario entries with map_exists=1 missing corresponding FIM files")
	flags.StringVar(&outScenarios, "o_scenarios", "missing_scenarios.csv", "Output CSV for FIM entries missing corresponding scenario records")
	flags.IntVar(&concurrent, "cc", 25, "Concurrent Count, number of top-level reach directories to process concurrently (default 25)")
	flags.BoolVar(&skipEmpty, "skip_empty", false, "If true, do not create an empty output CSV file")

	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("error parsing flags: %v", err)
	}

	// Validate required flags
	if dbPath == "" || fimLibDir == "" {
		flags.PrintDefaults()
		return fmt.Errorf("missing required flags")
	}

	// Check if gdalbuildvrt or GDAL tool is available
	if strings.HasPrefix(fimLibDir, "/vsi") && !utils.CheckGDALToolAvailable(gdalLSName) {
		return fmt.Errorf(`%[1]s is not available. Please install GDAL and ensure %[1]s is in your PATH. %[1]s is not available in PATH
		by default. Please refer to docs for instructions on how to add it to Path`, gdalLSName)
	}

	// 1) Open the input DB ( we won't modify it).
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("database file does not exist: %s", dbPath)
	}
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return fmt.Errorf("error opening DB: %v", err)
	}

	slog.Debug("Database connection established")
	defer db.Close()

	if err := utils.CheckScenariosTable(db); err != nil {
		return err
	}
	if err := utils.CheckFimPathColumn(db); err != nil {
		return err
	}

	// 2) Attach an in-memory DB for fim_entries
	_, err = db.Exec(`ATTACH ':memory:' AS memdb;`)
	if err != nil {
		return fmt.Errorf("error attaching in-memory db: %v", err)
	}

	// Create table memdb.fim_entries
	// to do: move query to a constant at the top of the file
	_, err = db.Exec(queryCreateFIMEntTable)
	if err != nil {
		return fmt.Errorf("error creating memdb.fim_entries: %v", err)
	}

	var absFimLibDir string
	if strings.HasPrefix(fimLibDir, "/vsi") {
		absFimLibDir = fimLibDir
	} else {
		absFimLibDir, err = filepath.Abs(fimLibDir)
		if err != nil {
			return fmt.Errorf("error getting absolute path for fim library: %v", err)
		}
	}

	// 3) Setup concurrency
	fimChan := make(chan string, 2000) // buffer for discovered FIM paths
	var batchWG sync.WaitGroup

	// Single writer goroutine that batch-inserts rows into memdb.fim_entries
	batchWG.Add(1)
	go func() {
		defer batchWG.Done()
		if err := batchInsertFIMs(db, fimChan); err != nil {
			slog.Error("Could not insert FIM rows into memory db", "error", err)
		}
	}()

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrent) // limit concurrency to 'cc'
	// sync/semaphore could also have been used here

	// 4) Find top-level directories (reach folders) and process them
	libEntries, err := readDir(absFimLibDir, false)
	if err != nil {
		return fmt.Errorf("error reading fim library directory: %v", err)
	}

	var reachDirs []dirEntry
	for _, de := range libEntries {
		if de.isDir {
			reachDirs = append(reachDirs, de)
		}
	}

	if len(reachDirs) == 0 {
		if strings.HasPrefix(fimLibDir, "/vsi") {
			return fmt.Errorf("no entries found in VSI path. Is it a valid FIM library? Does GDAL have access to cloud credentials?")
		}
		return fmt.Errorf("no entries found in fim library directory. Not a valid fim library")
	}
	slog.Debug("Finished processing reach directories", "reach_dir_count", len(reachDirs))

	var reachDir string
	for _, de := range libEntries {
		if !de.isDir {
			processLibEntry(de, absFimLibDir, fimChan)
			continue
		}
		wg.Add(1)
		sem <- struct{}{} // Acquire concurrency token
		go func(reachDir string) {
			defer wg.Done()
			defer func() { <-sem }() // Release token
			reachEntries, err := readDir(de.path, true)
			if err != nil {
				slog.Warn("Reach directory read error", "path", de.path, "error", err)
				return
			}
			for _, e := range reachEntries {
				processLibEntry(e, absFimLibDir, fimChan)
			}
		}(reachDir)
	}

	// Wait for all reach processing goroutines to finish
	wg.Wait()
	close(fimChan) // no more FIM rows
	batchWG.Wait() // wait for the DB writer goroutine

	// 5) Query DB for missing data and write to CSV
	tasks := []struct {
		outFile string
		query   string
		label   string
	}{
		{outFims, queryMissingFims, "missing FIMs"},
		{outScenarios, queryMissingScenarios, "missing scenarios"},
	}
	for _, task := range tasks {
		rows, err := db.Query(task.query)
		if err != nil {
			return fmt.Errorf("error executing query for %s: %v", task.outFile, err)
		}
		defer rows.Close()

		rowCount, err := writeCSV(rows, task.outFile, skipEmpty)
		if err != nil {
			return fmt.Errorf("error writing %s: %v", task.outFile, err)
		}
		fmt.Printf("Number of %s records found: %d\n", task.label, rowCount)
		if rowCount > 0 || !skipEmpty {
			fmt.Printf("File for %s created at %s\n", task.label, task.outFile)
		}
	}

	fmt.Println("Validation complete")
	return nil
}
