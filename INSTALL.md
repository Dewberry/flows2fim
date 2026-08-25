# Installation Instructions


## Docker
1. **Build Docker Image**
   - cd into root of repository
   - Build container `docker build -t flows2fim:latest --target prod .`

2. **Verify**
    - Run `docker run --rm -it flows2fim:latest --version` to confirm everything works.

## Windows

1. **Install GDAL**
   - The easiest way is via [**OSGeo4W**](https://trac.osgeo.org/osgeo4w/#QuickStartforOSGeo4WUsers)

2. **Setup flows2fim**
    - Go to the [**Releases**](https://github.com/NGWPC/flows2fim/releases) page and download the `flows2fim-windows-amd64.zip`
    - Unzip and copy `flows2fim.exe` into `C:\OSGeo4W\bin`

3. **(Optional) Enable `gdal_ls`**
   - This step is **only needed** if you plan to use the `flows2fim validate` command with a FIM library on cloud storage.

   _Your actual paths might be slightly different based on the version of python_
   - Copy `C:\OSGeo4W\apps\Python312\Lib\site-packages\osgeo_utils\gdal_ls.py` to `C:\OSGeo4W\apps\Python312\Scripts`.
   - In `C:\OSGeo4W\apps\Python312\Scripts`, make a copy of `gdal_merge.bat` → `gdal_ls.bat`.
   - Open `gdal_ls.bat` and replace all occurrences of `gdal_merge.py` with `gdal_ls.py`.

4. **Verify**
    - Open the **OSGeo4W Shell**.
    - Run `flows2fim --version` to confirm everything works.
    - Run `gdalinfo --version` to confirm everything works.
    - (Optional) Run `gdal_ls --version` if you set it up.


## Linux

1. **Install GDAL**

    _For Debian based distros:_
   ```bash
   sudo apt-get update && sudo apt-get install -y gdal-bin
   ```

   There are various ways to install GDAL. See GDAL documentation for more details. If you are struggling with GDAL installation, consider using Docker.

2. **Setup flows2fim**
    - **Option 1 - Download**
        - Go to the [**Releases**](https://github.com/NGWPC/flows2fim/releases) page and download the `flows2fim-linux-<your-cpu-arch>.tar.gz`
        - Extract it and move to a directory in `$PATH` (e.g., `/usr/local/bin`) and make it executable:
            ```bash
            tar -xvf flows2fim-linux-amd64.tar.gz
            sudo mv flows2fim /usr/local/bin/
            sudo chmod +x /usr/local/bin/flows2fim
            ```
    - **Option 2 - Build from source**
        - cd into root of repository
        - Build container
            - `docker compose up -d`
        - Get flows2fim CONTAINER_ID
            - `CONTAINER_ID=$(docker ps -q | head -n 1)`
        - Issue build script
            - `docker exec $CONTAINER_ID /bin/bash -c "./scripts/build-linux-amd64.sh"`
        - Shutdown container
            - `docker compose down`
        - Move executable to `$PATH`
            - `sudo mv builds/linux-amd64/flows2fim /usr/local/bin/`
            - `sudo chmod +x /usr/local/bin/flows2fim`
3. **(Optional) Enable `gdal_ls`**
   - Only required if using `flows2fim validate` with FIM libraries stored on cloud.
   - On Ubuntu/Debian systems:
     ```bash
     sudo cp /usr/lib/python3/dist-packages/osgeo_utils/samples/gdal_ls.py /usr/local/bin
     sudo chmod +x /usr/local/bin/gdal_ls.py
     ```

4. **Verify**
    - Run `flows2fim --version` to confirm everything works.
    - Run `gdalinfo --version` to confirm everything works.
    - (Optional) Run `gdal_ls.py --version` if you set it up.

## Mac

1. **Install GDAL**

    The easiest way to install GDAL on a Mac is through homebrew.
    If [homebrew](https://brew.sh/) is not on your Mac, install it:
    ```bash
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    ```
    Verify homebrew installation:
    ```bash
        brew --version
    ```
    Install GDAL:
    ```
        brew install gdal
    ```
    Verify:
    ```
        gdalinfo --version
    ```

   There are various ways to install GDAL. See GDAL documentation for more details. If you are struggling with GDAL installation, consider using Docker.

2. **Setup flows2fim**
    - **Option 1 - Download**
        - Go to the [**Releases**](https://github.com/NGWPC/flows2fim/releases) page and download (double click) `flows2fim-darwin-arm64.tar.gz`
        - Open the Mac Finder application and navigate to your `Downloads` folder.
        - Right click > Open the `flows2fim-darwin-arm64.tar.gz` file. This *should* avoid errors arising from Apple security settings.
        - In a terminal window, move the flows2fim executable to a directory in your `$PATH` (e.g., `/usr/local/bin`), and add execute permissions:
            ```bash
            cd ~/Downloads
            sudo mv flows2fim /usr/local/bin/
            sudo chmod a+x /usr/local/bin/flows2fim
            ```
        - Verify `flows2fim` is executable and in your `$PATH`.
            ```bash
            flows2fim --version
            ```
        - If the error `"flows2fim" can't be opened because Apple cannot check it for malicious software.` or similar comes up, you may temporarily override your Mac's security settings with the following command:
            ```bash
            xattr -d com.apple.quarantine /usr/local/bin/flows2fim
            ```

    - **Option 2 - Build from source**
        - cd into root of repository
        - Build container
            - `docker compose up -d`
        - Get flows2fim CONTAINER_ID
            - `CONTAINER_ID=$(docker ps -q | head -n 1)`
        - Issue build script
            - `docker exec $CONTAINER_ID /bin/bash -c "./scripts/build-darwin-arm64.sh"`
        - Shutdown container
            - `docker compose down`
        - Move executable to `$PATH`
            - `sudo mv builds/darwin-arm64/flows2fim /usr/local/bin/`
            - `sudo chmod a+x /usr/local/bin/flows2fim`

3. **`gdal_ls` not supported on MacOS currently (gdal_ls is not included in gdal homebrew installation)**

4. **Verify**
    - Run `flows2fim --version` to confirm everything works.
