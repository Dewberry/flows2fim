# `flows2fim`

[![Build and Test - Linux](https://github.com/NOAA-OWP/flows2fim/actions/workflows/build_and_test_linux.yml/badge.svg?event=push)](https://github.com/NOAA-OWP/flows2fim/actions/workflows/build_and_test_linux.yml) [![Build and Test - Mac](https://github.com/NOAA-OWP/flows2fim/actions/workflows/build_and_test_mac.yml/badge.svg?event=push)](https://github.com/NOAA-OWP/flows2fim/actions/workflows/build_and_test_mac.yml)

![alt text](image.png)
## Overview
`flows2fim` is a command line utility program that creates composite FIMs for different flow conditions utilizing input FIM libraries and scenarios database.
`flows2fim` can work directly with FIM libraries stored on the cloud.

It has the following basic commands:

 - `controls`: Given a flow file and a scenarios database, create a control table of reach flows and downstream boundary conditions.
 - `fim`: Given a control table and a fim library folder. Create a flood inundation map for the control conditions.
 - `domain`: Given a reach_id list (or a control table) and a fim library folder, create a composite domain map for the given reaches.

The following advanced commands are available but are not commonly needed:
 - `validate`: Given a FIM library folder and a scenarios database, validate there is one-to-one correspondence between the entries of the scenarios table and FIM library objects.

### Dependencies:
 - `GDAL` must be installed and available in PATH.

### Units:
Current support is for English units. The flow values must be in `cfs`

## Quick Start For Users

1. Follow install instructions at [INSTALL.md](INSTALL.md)

1. Get familiar using `flows2fim -h` and `flows2fim COMMAND -h`.

1. Download Baxter sample data from ` s3://fimc-data/flows2fim/sample_data/v0_2_0/Baxter` if you don't have a dataset already.

1. To create a controls file from the 100yr flows file run `flows2fim controls -db "Baxter/ripple.gpkg" -f "Baxter/flows/flows_100yr_cfs.csv" -o "Baxter/controls_100yr.csv" -sids 2821866`

1. To create Depth VRT run `flows2fim fim -lib "Baxter/library" -type depth -c "Baxter/controls_100yr.csv" -o "Baxter/fim_100yr.vrt"`

## Quick Start For Developers

1. Clone the repository and perform the following steps from the root of the repo.

1. Download Baxter testdata from `s3://fimc-data/flows2fim/sample_data/v0_2_0/Baxter` to `testdata/Baxter` folder.

2. Launch a docker container using `docker compose up` and run the following commands inside the container

3. Run `go run main.go controls -db "testdata/Baxter/ripple.gpkg" -f "testdata/Baxter/flows/flows_100yr_cfs.csv" -o "testdata/Baxter/controls_100yr.csv" -sids 2821866` This will create a controls.csv file

4. Run `go run main.go fim -lib "testdata/Baxter/library" -type depth -c "testdata/Baxter/controls_100yr.csv" -o "testdata/Baxter/fim_100yr.vrt"` This will create a VRT file. VRT can be tested by loading in QGIS.

#### Testing

1. Run `go test ./...` to run automated tests.

#### Building

Run `./scripts/build-linux-amd64.sh` This will place the executable in `builds/linux-amd64`.

## Installation
[Installation](INSTALL.md)
