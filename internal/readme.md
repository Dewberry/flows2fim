# Internal Directory

## Overview
The `internal` directory is for private application and library code. This code is not intended to be imported by other applications.

## Guidelines
- **Restricted Access**: Code in `internal` is restricted by `go` and inaccessible to other projects.
- **Application-Specific**: Contains code specifically tailored for the internal workings of the application. Hence `internal` pkgs should not be used public pkgs located in `pkg` direcotry.

## Usage
This approach provides a clear boundary for code that is not meant to be exposed externally.
