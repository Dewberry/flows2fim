# Test Matrix

|                                                 |                      |                      |                   |
| :---------------------------------------------- | :------------------- | :------------------- | :---------------- |
|                                                 | **generate outputs** | **regression tests** | **Assert errors** |
|                                                 |                      |                      |                   |
| **controls**                                    |                      |                      |                   |
| Generate 2,5,10,25,50,100 yr controls.csv files | &#9745;              | &#9745;              |                   |
| missing `-scsv`                                 |                      |                      | &#9745;           |
| missing `-scvs`, `-o`                           |                      |                      | &#9745;           |
| `-db` only                                      |                      |                      | &#9745;           |
| `-scsv` empty                                   |                      |                      | &#9745;           |
| `-f` empty                                      |                      |                      | &#9745;           |
| `-f` columns swapped                            |                      |                      | &#9745;           |
| `-f` no values                                  |                      |                      | &#9745;           |
|                                                 |                      |                      |                   |
| **fim**                                         |                      |                      |                   |
| Generate 2,5,10,25,50,100 yr fim.tif files      | &#9745;              | &#9745;              |                   |
| Generate depth fims in different output formats | &#9745;              | &#9745;              |                   |
| Generate extent fims in different output formats|                      |                      |                   |
| missing `-c`                                    |                      |                      | &#9745;           |
| missing `-l`                                    |                      |                      | &#9745;           |
| missing `-o`                                    |                      |                      | &#9745;           |
| missing `-type`                                 |                      |                      |                   |
| empty `-c` file                                 |                      |                      | &#9745;           |
| -with_domain                                    |                      |                      |                   |
|                                                 |                      |                      |                   |
| **validate**                                    |                      |                      |                   |
|                                                 |                      |                      |                   |
