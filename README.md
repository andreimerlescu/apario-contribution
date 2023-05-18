# Project Apario Contribution Ingestion

This project is responsible for scanning an incoming CSV/PSV or XLSX metadata file that contains predictable rows of records that
presumably are PDF files that can be ingested/processed for consumption by Project Apario.

## Build

```shell
go build -a -race -v -o apario-contribution .
```

## Data Sets

| Name | Filename | Rows   | Pages | Notes |
|------|----------|--------|-------|-------|
| JFK 2018 | `importable/jfk2018.xlsx` | 54,636 | 334,031 | First Set of JFK Files |
| JFK 2021 | `importable/jfk2021.xlsx` | 1,491  | 18,870 | Second Set of JFK Files |
| JFK 2022 | `importable/jfk2022.xlsx` | 13,263 | 165,644 | Third Set of JFK Files |
| JFK 2023 | `importable/jfk2023.xlsx` | 1,279 | 16,174 | Four Set of JFK Files |
| STARGATE | `importable/stargate.psv` | 12,339 | | DIA Remote Viewing Program |

## Running

```shell
time ./apario-contribution -dir tmp -file importable/jfk2023.xlsx -limit 369 -buffer 4550819
```

Do not be surprised if this process takes a very long to complete depending on how fast your system is.

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