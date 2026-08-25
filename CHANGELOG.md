# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.5.0] - 2026-08-04

**Note:** Compatible with outputs from Ripple1D Pipeline version 0.11.0 (the release containing the `scenarios` rename) to present. Not compatible with earlier outputs, which carry a `rating_curves` table instead, see the Changed section below.

### Added

- `controls` and `validate` now fail with an actionable message when the database has a pre-0.5.0 `rating_curves` table instead of `scenarios`
- `controls` and `validate` now fail with an actionable message when the database has a pre-0.5.0 `rating_curves_no_map` table
- `controls` output file now has a `map_exists` column
- `fim` leaves records whose `map_exists` is `0` out of the composite. If the controls file has no `map_exists` column, it warns about that and includes every record
- `validate` now has a third check and a `-o_unexpected_fims` flag (default `unexpected_fims.csv`), listing FIM files whose scenario records have `map_exists = 0`, that is, a raster exists where the database says none was written
- .github/scripts/generate-release-notes.sh - helper script to GHA workflow
- .github/workflows/release.yml - GHA workflow to automate releases
- scripts/create-release-binaries.sh - separate binary creation script - add error messaging and exit codes for use in GHA workflow

### Changed

- **BREAKING** - the database table read by `controls` and `validate` is renamed from `rating_curves` to `scenarios`. These records are computed model scenarios (a discharge, plus a downstream stage for `kwse`, with the resulting water surface elevations), not rating curves; a rating curve is a discharge-vs-depth relationship at a point. Requires a database produced by the corresponding ripple1d / ripple1d-pipeline release. There is no fallback to the old table name
- **BREAKING** - `validate` flag `-o_rcs` is renamed to `-o_scenarios`, and its default output file changes from `missing_rating_curves.csv` to `missing_scenarios.csv`
- `controls` output header is now `reach_id,flow,control_stage,map_exists`. The new column is appended last, so readers of the first three columns (including `flows2fim domain`) are unaffected
- `validate` progress messages now read `Number of <what> records found` and `File for <what> created at`, since not every check reports something missing
- CHANGELOG.md - adhere to [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) standard
- scripts/create-release-assets.sh - separate release assets and binaries creation, mkdir release_assets/ directory as part of script, and add checksum generation

### Removed

### Fixed

## [0.4.1] - 2025-09-17

### Added

- Binary build for `linux-arm64` platform added to releases
- Enhanced error reporting

### Changed

Build Support Matrix:

| OS \ ARCH   | AMD64 | ARM64 |
| :---------- | :---: | :---: |
| **Linux**   | ✅    | ✅    |
| **MacOS**   |       | ✅    |
| **Windows** | ✅    |       |

## [0.4.0] - 2025-05-15

**Note:** Compatible with outputs from Ripple1D Pipeline version 0.10.3 to 0.10.4

### Added

- New `domain` command to create a composite domain map for the given reaches

### Changed

- Format of extent library modified to improve performance of `fim` command
- API and default output for `fim` command changed - argument `-type depth|extent` is no longer required
- Argument `--with_domain` must now be added if domain should be included as background in composite map
- `fmt` argument for `fim` command now uses `GTiff` option for GTiff format (not `tif`) to be consistent with GDAL raster drivers short names
- GDAL version requirement relaxed - now works with GDAL 3.4.0 or greater (previously required 3.8.0+)

## [0.3.0] - 2025-03-26

**Note:** Compatible with outputs from Ripple1D Pipeline version unknown to 0.7.0

### Added

- New `validate` command to validate FIM libraries (stored locally or on cloud) against rating curves table in a database. See Install.md for setup details
- Configuration through environment variables:
    - `F2F_LOG_LEVEL`: Set logging level ('DEBUG', 'INFO', 'WARN', 'ERROR'). Default is 'INFO'
    - `F2F_NO_COLOR`: Set to 'TRUE' to disable colored output. Default is 'FALSE'
- Structured logging support

### Changed

- API for `fim` command modified - argument `-type depth|extent` is now required
- GDAL version requirement increased to 3.8.0 or greater when working with extent libraries

### Removed

- `-rel` argument from `fim` command

### Fixed

- Bug in `fim` command on extent libraries that was causing gaps in composite FIMs

## 0.2.1 - 2024-11-15

### Fixed

- Bug that was assuming no data value of -9999.0 for all raster types - no data value is now inferred from FIM library rasters
- Error now raised in `fim` command when controls file is empty

## 0.2.0 - 2024-11-01

**Note:** Compatible with outputs from Ripple1D pipeline version unknown to 0.7.0

### Changed

- Default database table name changed from `conflation` to `network`
- Default `to_id` name changed from `conflation_to_id` to `updated_to_id`
- Removed warnings for when stage difference is because of normal depth higher than target stage

### Added

- COG (Cloud Optimized GeoTIFF) option for `fim` command

## 0.1.0 - 2024-12-18

### Added

- Command-line utility for generating composite Flood Inundation Maps (FIMs) under varying flow conditions
- Core commands:
    - `controls`: Process flow files with rating curve data to generate control tables containing reach flows and downstream boundary conditions
    - `fim`: Combine control tables with FIM library folders to produce flood inundation maps matching specified flow conditions
    - `domain`: Generate composite domain maps from reach identifier lists or control tables using FIM library data
- FIM library support for local and cloud-stored libraries
- Rating curve database integration
- GDAL-based geospatial operations
- Support for English units with flow values in cubic feet per second (cfs)
- Multi-platform binary support (Linux AMD64, macOS ARM64, Windows AMD64)

**Note:** This initial release is based on the archived repository at [NGWPC/flows2fim-archive](https://github.com/NGWPC/flows2fim-archive/releases/tag/v0.1.0)

[Unreleased]: https://github.com/NGWPC/flows2fim/compare/v0.5.0...HEAD
[0.5.0]: https://github.com/NGWPC/flows2fim/releases/tag/v0.5.0