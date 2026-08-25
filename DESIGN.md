## Purpose of Directories

- `cmd`: Contains individual folders for different commands.
- `pkg`: Houses reusable packages potentially useful in other projects.
- `internal`: For private application code not intended for external use.
- `scripts`: Includes useful scripts for building, testing, and more.

## Design Thoughts

### FIM

1. GDAL will always use relative paths if output and inputs are in same lineage, this was the reason for removing -rel flag
1. gdalbuildvrt don't support cloud relative paths. This does not work `gdalbuildvrt /vsis3/fimc-data/fim2d/prototype/2024_03_13/vsi_relative.vrt ./8489318/z0_0/f_1560.tif ./8490370/z0_0/f_130.tif`
1. To simplify fim.go, all paths are converted to absolute paths and the relative logic is left to `gdalbuildvrt`
1. We looked into `gdal_merge`, `gdalwarp`, `gdalbuildvrt`. None of them have a way to merge rasters with maximum value of each pixel. The only possible option out there is pixel function with VRT, which was used in v0.3.0 for depth library type. It was extremely slow because https://gis.stackexchange.com/a/491960/142232. Hence in 0.4.0, the FIM library is modified to store FIMs and domains separetely, this would eliminate the need of pixel by pixel calculations completely, turmendoulsy improving speed. This is also a better design because domains are not always needed anyways and were an unnecessary burden in NRT executions.

### Validate

1. A  `-o_<type>` convention is used to allow for multiple outputs across subcommands.
2. An in-memory SQLite database is attached in read-only mode, ensuring the input database is not altered.
3. Batch inserts through a single goroutine are employed to minimize concurrency issues with SQLite and to reduce the number of round trips. This is necessary since consistency is desired over speed.
4. A single SQL query with rc_exist and fim_exist columns is relied upon, allowing the database to handle processing instead of in-memory go logic.
5. CSV rows are written one by one to avoid substantial memory usage at the end of processing.
6. Specified output files, even if empty, are generated to keep API consistent.
7. Atomic writes with a temporary file are performed for CSVs so no partial CSVs remain if errors occur.
8. Local directories are not processed with gdal_ls, preserving fast local operations and avoiding unnecessary dependencies.



### Backward Compatibility

- We are not supporting backward compatibility generally, but in case of migration from 0.4.1 to 0.5.0 (rating_curves to scenarios and map_exist support), we provided helpful error messages because it was easy to do that.