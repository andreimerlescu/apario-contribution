# Project Apario Contribution Ingestion

This project is responsible for scanning an incoming CSV/PSV or XLSX metadata file that contains predictable rows of records that
presumably are PDF files that can be ingested/processed for consumption by Project Apario.

## Build

```shell
go build -a -race -v -o apario-contribution .
```

### Runtime Requirements

You'll need to download and install the dependencies:

```shell
git clone git@github.com:andreimerlescu/apario-contribution.git
cd apario-contribution
make install
make build
```

You'll also need to ensure that the following binaries are installed on the host that will execute this program.

```go
var RawBinaries = []string{
	"pdfcpu",
	"gs",
	"pdftotext",
	"convert",
	"composite",
	"pdftoppm",
	"tesseract",
}
```

It is required that you have accessible to the runtime of this executable the aforementioned binaries as they are utilized
throughout the compilation process of generating assets. For each of these, you should be able to run with success:

```shell
$ which pdfcpu
/usr/local/bin/pdfcpu

$ which gs
/usr/local/bin/gs

$ which pdftotext
/usr/local/bin/pdftotext

$ which convert
/usr/local/bin/convert

$ which composite
/usr/local/bin/composite

$ which pdftoppm
/usr/local/bin/pdftoppm

$ which tesseract
/usr/local/bin/tesseract
```

If you're missing any of these binaries, please consult with your preferred search engine for further assistance.

## Data Sets

| Name | Filename | Rows    | Pages   | Notes                      |
|------|----------|---------|---------|----------------------------|
| JFK 2018 | `importable/jfk2018.xlsx` | 54,636  | 334,031 | First Set of JFK Files     |
| JFK 2021 | `importable/jfk2021.xlsx` | 1,491   | 18,870  | Second Set of JFK Files    |
| JFK 2022 | `importable/jfk2022.xlsx` | 13,263  | 165,644 | Third Set of JFK Files     |
| JFK 2023 | `importable/jfk2023.xlsx` | 1,279   | 16,174  | Fourth Set of JFK Files    |
| STARGATE | `importable/stargate.psv` | 12,339  | 91,604  | DIA Remote Viewing Program |

## Running

There are two ways you can run the program, the first and most pure way of doing it is to directly call the built binary
with your own arguments, such as:

```shell
time ./apario-contribution -dir tmp -file importable/jfk2023.xlsx -limit 369 -buffer 4550819
```

The other way of doing it is much easier, this looks like:

```shell
make run jfk2023.xlsx
```

You can replace `jfk2023.xlsx` with the file inside the `importable/` directory. The script supports XLSX, CSV and PSV
file extensions, but the formatting of the data does matter. For instance, if you're running on an XLSX, it will only
process the first sheet. All other sheets are ignored by the script. It also assumes that the first line is the headers.

Do not be surprised if this process takes a very long to complete depending on how fast your system is. The script is
designed to technically be resumable but it does not do a very good job with it at the moment. For instance, if the PDF
is already downloaded, it won't re-download it. However, if the process exits before it's completed, and you resume it,
it'll regenerate the manifest files, thumbnails, and re-perform the OCR on the project.

### Command Line Arguments

| Flag         | Default   | Notes                                                                  |
|--------------|-----------|------------------------------------------------------------------------|
| `-file`      | __blank__ | CSV file of URL + Metadata.                                            | 
| `-dir`       | __blank__ | Path of the directory you want the export to be generated into.        | 
| `-limit`     | `1`       | Number of rows to concurrently process.                                | 
| `-buffer`    | `172032`  | Memory allocation for CSV buffer (min 168 * 1024 = 168KB)              | 
| `-tesseract` | `1`       | Semaphore Limiter for `tesseract` binary.                              | 
| `-download`  | `2`       | Semaphore Limiter for downloading PDF files from URLs.                 | 
| `-pdfcpu`    | `17`      | Semaphore Limiter for `pdfcpu` binary.                                 | 
| `-gs`        | `17`      | Semaphore Limiter for `gs` binary.                                     | 
| `-pdftotext` | `17`      | Semaphore Limiter for `pdftotext` binary.                              | 
| `-convert`   | `17`      | Semaphore Limiter for `convert` binary.                                | 
| `-pdftoppm`  | `17`      | Semaphore Limiter for `pdftoppm` binary.                               | 
| `-png2jpg`   | `17`      | Semaphore Limiter for converting PNG images to JPG.                    | 
| `-resize`    | `17`      | Semaphore Limiter for resize PNG or JPG images.                        | 
| `-shafile`   | `36`      | Semaphore Limiter for calculating the SHA256 checksum of files.        | 
| `-watermark` | `36`      | Semaphore Limiter for adding a watermark to an image.                  | 
| `-darkimage` | `36`      | Semaphore Limiter for converting an image to dark mode.                | 
| `-filedata`  | `369`     | Semaphore Limiter for writing metadata about a processed file to JSON. | 
| `-shastring` | `369`     | Semaphore Limiter for calculating the SHA256 checksum of a string.     | 
| `-wjsonfile` | `369`     | Semaphore Limiter for writing a JSON file to disk.                     | 

## Output

When the program executes, you'll supply a `-dir` flag argument which will be a string to an existing directory with 
write permissions on your system. If the directory does not reflect that level of permission, you must do whatever
is necessary to ensure that its in a writable and working state prior to running this code.

As the script processes the CSV or XLSX file supplied with the `-file` flag argument, you'll begin to see files created.
For the example provided, the dir is called `./tmp` and it will look something like this:

```log
├── apario-contribution
├── logs
│   └── engine-2023-05-17-17-17-17.log
├── tmp
│   ├── f9928f63d48bf43aa8222b0eaed10c8f51b92a39a0e9db4370cedbe72ee682ac
│   │   ├── 2023_104-10143-10058.pdf
│   │   ├── ocr.txt
│   │   ├── pages
│   │   │   ├── 2023_104-10143-10058_page_1.pdf
│   │   │   ├── 2023_104-10143-10058_page_2.pdf
│   │   │   ├── manifest.000001.json
│   │   │   ├── manifest.000002.json
│   │   │   ├── page.light.000001.original.png
│   │   │   ├── page.light.000002.original.png
│   │   │   ├── page.dark.000001.original.png
│   │   │   └── page.dark.000002.original.png
│   │   └── record.json

```

Inside the `tmp/` directory you'll see something like `f9928f63d48bf43aa8222b0eaed10c8f51b92a39a0e9db4370cedbe72ee682ac`.
This happens to be the SHA256 checksum of the URL that the specific documents belongs to as its source. For this example,
`https://www.archives.gov/files/research/jfk/releases/2023/104-10143-10058.pdf`. Inside of that is the PDF file itself,
and in the case of the JFK files, the subfolder "2023" is just merged into "2023_" using a simple `strings.ReplaceAll()`.
In addition to the downloaded original PDF, the ocr.txt file is the output of `pdftotext` if the PDF has text objects
inside it. Most if not all DECLAS OSINT from JFK/STARGATE are flattened images making them impossible to search, thus
why this project is needed in the first place. The `pages/` subdirectory is responsible for hosting a `<filename>_page_#.pdf`
file that is just an extracted page, a manifest.0000#.json that contains paths to image assets and metadata, and then
the actual image assets as `page.(light|dark).(pageNumber).(png|jpg)`. When the process is complete, you'll see .JPG files.
If the process is incomplete, you may see .PNG files. The JPG images are progressive at 75% quality, the PNG are uncompressed
but resampled to 369px/in. 

In addition to these assets, a `record.sql` file will soon be added that will provide the necessary PostgreSQL insert
statements required to ensure that the row scanned from the input file is accessible via the Project Apario database/GUI.

Finally, in the `worker.go` file are a few unused functions that will be expanded to include processing the OCR text from
the document to either clean it up, generate ngram sequences, and build a master dictionary of words->[]pages to help
improve the performance of complex searching that will be offered in the desktop build of Project Apario.



## Logging

```log
2023/05/17 17:17:17 headerFields = File Name,Record Num,NARA Release Date,Formerly Withheld,Agency,Doc Date,Doc Type,File Num,To Name,From Name,Title,Num Pages,Originator,Record Series,Review Date,Comments,Pages Released
2023/05/17 17:17:17 processRecord received for row [{File Name 2023/104-10061-10328.pdf} {Num Pages 1} {Record Series JFK} {Comments JFK9 : F60 : 1993.07.12.14:32:01:090580 :} {Record Num 104-10061-10328} {From Name SE/OP/S, CIA} {Pages Released 1} {NARA Release Date 05/11/2023} {Formerly Withheld Redact} {Doc Type PAPER - TEXTUAL DOCUMENT} {To Name SE/SAG/OP, CIA} {Review Date 03/31/2023} {Doc Date 06/27/1978} {File Num 80T01357A} {Title SERGEY ALEKSANDROVICH UZLOV.} {Originator CIA}]
2023/05/17 17:17:17 pdf_url = https://www.archives.gov/files/research/jfk/releases/2023/104-10061-10328.pdf
2023/05/17 17:17:17 downloading URL https://www.archives.gov/files/research/jfk/releases/2023/104-10061-10328.pdf to tmp/52a800b39d62a74b2ec3add1e1bb2b1b7a932e04834fbcace757f64f1ec28f63/2023_104-10061-10328.pdf
2023/05/17 17:17:17 sending URL https://www.archives.gov/files/research/jfk/releases/2023/104-10061-10328.pdf (rd struct) into the ch_ImportedRow channel
2023/05/17 17:17:17 processRecord received for row [{Title HSCA REQUEST} {Comments JFK9 : F64 : 1993.07.12.13:45:27:590530 :} {Pages Released 1} {Doc Type PAPER - TEXTUAL DOCUMENT} {Num Pages 1} {Doc Date 06/07/1978} {NARA Release Date 05/11/2023} {Formerly Withheld Redact} {File Num 80T01357A} {From Name WITHHELD} {Originator CIA} {Record Series JFK} {Review Date 03/31/2023} {File Name 2023/104-10061-10266.pdf} {Record Num 104-10061-10266}]
2023/05/17 17:17:17 pdf_url = https://www.archives.gov/files/research/jfk/releases/2023/104-10061-10266.pdf
2023/05/17 17:17:17 started validatePdf(2023UWG662) = tmp/52a800b39d62a74b2ec3add1e1bb2b1b7a932e04834fbcace757f64f1ec28f63/2023_104-10061-10328.pdf
...
```

## Running on Docker

If you wish to run this project purely within Docker and not rely on installing the required binaries directly to your
system, or if you're on a system that doesn't natively support the binary dependencies, then Docker is your friend.

### Build Docker Image

The file is `Dockerfile` and it can be compiled using:

```shell
make dbuild
```

### Running Docker Container

When running the docker container, you are required to supply the required argument `<filename>` in it, such as:

```shell
make drun jfk2023.xlsx
```

Inside the container, you're running as the user `apario` and not as `root`.

## Future Proves Past

Project Apario was a monolithic Rails application that delivered the first 2 sets of the JFK Assassination Records to 
the public in a fully searchable manner. As costs increased and funding stretched thin, a pivot was necessary to replace
the underlying technology with something more robust like Go. 

As is, this project only accomplishes half of what is necessary to actually USE the pages of OSINT that its capable of
ingesting. Future updates to this code will incorporate those changes.

## Contribution Policy

Feel free to fork the project, submit a contribution, and then create a pull request. I'll review it and incorporate it 
if It's something that the project needs. 

### Contributors

- Andrei Merlescu (inventor of Project Apario)

## License

This project is licensed with the MIT License.