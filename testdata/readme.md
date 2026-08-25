# Updating /testdata  

### To update collection being used for testing:

- First pull collection data from S3 into /testdata folder
- Move qc/controls_* to testdata/reference_data/controls/
- Move qc/fim_* to testdata/reference_data/fim/
- Move ripple.gpkg to testdata/reference_data/db
- Run test script (`test_suite_linux.sh`) to generate fim_output_formats files, and move to testdata/reference_data/fim_output_formats (don't forget `fim_2year_test_rel_false.vrt`)
- Move library/ to testdata/reference_data/library
- Move start_reaches.csv to testdata/reference_data/start_reaches