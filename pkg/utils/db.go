package utils

import (
	"database/sql"
	"fmt"
)

// ScenariosTable is the database table flows2fim reads reach scenarios from.
const ScenariosTable = "scenarios"

// legacyScenariosTable is the pre-0.5.0 name of ScenariosTable. It is referenced
// only to produce a helpful error for databases built by an older
// ripple1d-pipeline. It, and the branch using it in CheckScenariosTable, are
// safe to delete once no collections carrying the old schema remain in
// circulation.
const legacyScenariosTable = "rating_curves"

// legacyNoMapTable is the table that older ripple1d-pipeline releases used to
// hold scenarios whose depth grid did not exists, that was hydraulically incorrect.
// It, and the branch using it in CheckScenariosTable, are
// safe to delete once no collections carrying the old schema remain in
// circulation.
const legacyNoMapTable = "rating_curves_no_map"

// tableExists reports whether the connected SQLite database has a table with
// the given name.
func tableExists(db *sql.DB, name string) (bool, error) {
	var found string
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?;`, name).Scan(&found)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// CheckScenariosTable verifies the database exposes a 'scenarios' table in the
// shape flows2fim expects, and returns an actionable error when it does not.
// Databases produced by an older ripple1d-pipeline carry a 'rating_curves' table
// instead, and older ones again keep mapless scenarios in a separate
// 'rating_curves_no_map' table. Each is reported with the SQL that fixes it and
// one step at a time, so an old database is migrated by re-running the command
// and applying whatever it asks for next: first the rename, then the merge.
func CheckScenariosTable(db *sql.DB) error {
	hasScenarios, err := tableExists(db, ScenariosTable)
	if err != nil {
		return fmt.Errorf("error inspecting database tables: %w", err)
	}

	if !hasScenarios {
		legacy, err := tableExists(db, legacyScenariosTable)
		if err != nil {
			return fmt.Errorf("error inspecting database tables: %w", err)
		}
		if legacy {
			return fmt.Errorf(
				"database has a '%[2]s' table but no '%[1]s' table. flows2fim 0.5.0 renamed this table, "+
					"because its records are computed model scenarios rather than rating curves. "+
					"Regenerate the database with a matching ripple1d-pipeline release, or rename the table in place with: "+
					"`ALTER TABLE %[2]s RENAME TO %[1]s`",
				ScenariosTable, legacyScenariosTable,
			)
		}
		return fmt.Errorf("database does not have a '%s' table", ScenariosTable)
	}

	// Scenarios whose depth grid was never written used to live in their own
	// table. They now belong in ScenariosTable with map_exists = 0, so a database
	// still carrying the old table need to be updated.
	noMap, err := tableExists(db, legacyNoMapTable)
	if err != nil {
		return fmt.Errorf("error inspecting database tables: %w", err)
	}
	if noMap {
		return fmt.Errorf(
			"database has a '%[1]s' table, which flows2fim 0.5.0 merged into '%[2]s'. "+
				"Regenerate the database with a matching ripple1d-pipeline release, or merge it in place with "+
				"  PRAGMA foreign_keys = OFF;\n"+
				"  ALTER TABLE %[2]s ADD COLUMN %[3]s BOOL CHECK(%[3]s IN (0, 1));\n"+
				"  UPDATE %[2]s SET %[3]s = 1;\n"+
				"  INSERT INTO %[2]s (reach_id, us_flow, us_depth, us_wse, ds_depth, ds_wse, boundary_condition, %[3]s)\n"+
				"    SELECT reach_id, us_flow, us_depth, us_wse, ds_depth, ds_wse, boundary_condition, 0 FROM %[1]s;\n"+
				"  DROP TABLE %[1]s;",
			legacyNoMapTable, ScenariosTable, "map_exists",
		)
	}

	return nil
}
