# Overview
This repo provides tools to decode an EU Digital COVID-19 Certificate (also referred to as Digital Green Certificates)
* [Decode using a public web page](#decode-using-the-Safetypass-verify-web-page)
* [Decode using a CLI tool on your local machine](#decode-using-the-CLI-tool)
* [EU DGC and CBOR Reference Documents](#reference-documents)
* [Appendix CLI Verbose Output](#appendix-CLI-verbose-output)

## Decoding Steps
The **decoding steps** are as follows:
1. Read the QR code to get the QR alphanumeric code. This is prefixed with "HC1:6BF...."
2. Base45 decode the part after "HC1:" to get the ZLIB compressed certificate
3. ZLIB inflate the compressed certificate to get a COSE tagged message (CBOR Object Signing and Encryption (COSE))
4. CBOR decode the COSE tagged message that contains
    - a Number key with value of 18 indicating that a "Cose_Sign1" object
    - a Content key with an array of containing the CBOR Web Token parts
        - protected header - cbor encoded []byte
        - unprotected header - map
        - payload  - cbor encoded []byte
        - digital signature (a signed sha256 digest) - []byte
5. CBOR decode the protected header to get the Signing Algorithm and KeyID
6. CBOR decode the payload to get the issuer, iat, exp, subject information, and vaccination information
7. NOT yet implemented check the COSE signature by getting signing key from issuing State and using it to check the CBOR signature.

## Decode using the SafetyPASS verify web page
This [web page](https://safetypass.sandbox.webshield.io/shc/web/v1/public/smart-health-card/upload) decodes an EU Digital 
COVID-19 Certificate and displays it contents. It **does not** save the data, or verify the signature (future)

The overview page
![overview page](https://raw.githubusercontent.com/webshield-dev/eudvcdecoder/main/images/display-dgc-page.png)

The show details section
![details page](https://raw.githubusercontent.com/webshield-dev/eudvcdecoder/main/images/details-1.png)
![details page](https://raw.githubusercontent.com/webshield-dev/eudvcdecoder/main/images/details-2.png)

The upload page (it stores no data)
![upload page](https://raw.githubusercontent.com/webshield-dev/eudvcdecoder/main/images/landing-page.png)

## Decode using the CLI tool
The CLI tool decodes an EU Digital COVID-19 Certificate QRCode from a file.(png|jpg)
It does NOT yet verify the Signature. Example Certificate Information

```
... run details removed for clarity ...
**** EU Covid-19 Certificate Summary **** 
Name:Erika Mustermann
DOB :1964-08-12
Vaccine Details
  Doses Administered: 2
  Doses Required:     2
  When:               2021-05-29
  Vaccine Product:    COVID-19 Vaccine Moderna
  Vaccine Type:       SARS-CoV-2 mRNA vaccine
  Vaccine Maker:      Moderna Biotech Spain S.L.
  Issuer:             Robert Koch-Institut
  ID:                 URN:UVCI:01DE/IZ12345A/5CWLU12RNOB9RXSEOP6FG8#W
```

## Setup
1. Git Clone/Fork this repo
2. The repo includes a **macOS binary** and a **linux binary**, otherwise you need to install Go.
   - If using go run then Install Go (1.14 or higher) see https://golang.org/doc/install


## Usage
The CLI flags are
1. `-qrfile <value>` the QRcode.png
2. `-verbose <level>` where level is 0 -> 9, default is zero

Using the executables macOS (./bin/decoder.mac) or Linux (./bin/decoder.linux)
- `./bin/decoder.mac -qrfile ./testfiles/dcc-testdata/DE/2DCode/png/1.png`     <-- Mac no verbose 
- `./bin/decoder.linux -qrfile ./testfiles/dcc-testdata/DE/2DCode/png/1.png`   <-- Linux no verbose

verbose
- `./bin/decoder.mac ./testfiles/dcc-testdata/DE/2DCode/png/1.png -verbose 1` <-- displays  protected header and common payload
- `./bin/decoder.mac . -qrfile ./testfiles/dcc-testdata/DE/2DCode/png/1.png -verbose 2` <-- displays decoding step details

Using Go run (all platforms)
- `go run . -qrfile ./testfiles/dcc-testdata/DE/2DCode/png/1.png`

Example output:
```
./bin/decoder.mac -qrfile ./testfiles/vaccine/dr_1.png
Decoding EU Covid-19 Certificate
  qrCodefile=./testfiles/vaccine/dr_1.png  ValueSetPath=./valuesetdata  verbose=0
  Step 1 - Read QR Code PNG ./testfiles/vaccine/dr_1.png Successfully...
  Step 2 - Base45 Decoded Successfully...
  Step 3 - ZLIB Inflated Successfully...
  Step 4 - CBOR UnMarshalled CBOR Web Token (CWT) using COSE tagged message COSE Number=18 Successfully...
    CWT CBOR UnMarshalled Protected Header Successfully...
    CWT Read UnProtected Header Successfully...
    CWT CBOR UnMarshalled Payload Successfully...
    CWT Read COSE Signature (single signer) Successfully...
Successfully Decoded EU Covid-19 Certificate

**** EU Covid-19 Certificate Summary **** 
Name:Erika Mustermann
DOB :1964-08-12
Vaccine Details
  Doses Administered: 2
  Doses Required:     2
  When:               2021-05-29
  Vaccine Product:    COVID-19 Vaccine Moderna
  Vaccine Type:       SARS-CoV-2 mRNA vaccine
  Vaccine Maker:      Moderna Biotech Spain S.L.
  Issuer:             Robert Koch-Institut
  ID:                 URN:UVCI:01DE/IZ12345A/5CWLU12RNOB9RXSEOP6FG8#W
```

## Testing
- Test QR.png(s) are from `https://github.com/eu-digital-green-certificates/dgc-testdata`
- `make test` runs local tests

## Verification
Verification covers
1. Verifying the COSE Signature to ensure message integrity and the issuers identity.
2. Verifying the issuer is a trusted party. 

The EU nation states both (1) issue the credentials and (2) control access to the public signing keys to verify the 
credentials, in much the same way as passports. So
    - The "trusted issuer" is an EU nation state.
    - Only parties blessed by the EU nation states can be "trusted verifiers", they are given access to the public signing key. 
Although technically a verifier could hand the public key to another verifier in a trusted fashion, it was unclear to me if 
this is some form of violation and penalty.

Hence, as it is a closed shop it was not possible verify the vaccine credentials.

Resources
- This was useful in understanding more https://github.com/Digitaler-Impfnachweis/certification-apis/blob/master/dsc-update/README.md

# Reference Documents

## EU Documents and Code

1. A good starting place has links to all other technical documents
    - https://ec.europa.eu/health/ehealth/covid-19_en

2. Github Repo
    - general - https://github.com/eu-digital-green-certificates
    - development - https://github.com/ehn-dcc-development
    - test data - https://github.com/eu-digital-green-certificates/dgc-testdata

3. EU Technical Specification Volumes
    - Volume 1 - Technical Specifications for Digital Green Certificates
        - https://ec.europa.eu/health/sites/default/files/ehealth/docs/digital-green-certificates_v1_en.pdf
    - Volume 2 - Technical Specifications for Digital Green Certificates
        - https://ec.europa.eu/health/sites/default/files/ehealth/docs/digital-green-certificates_v2_en.pdf
    - Volume 3 - Technical Specifications for Digital Green Certificates (Interoperable 2D Code, CBOR)
        - https://ec.europa.eu/health/sites/default/files/ehealth/docs/digital-green-certificates_v3_en.pdf
    - Volume 4 - European Digital Green Certificate Applications
      - https://ec.europa.eu/health/sites/default/files/ehealth/docs/digital-green-certificates_v4_en.pdf

4. EU Support
    - Certificate JSON Schema
        - https://github.com/ehn-dcc-development/ehn-dcc-schema
        - https://ec.europa.eu/health/sites/default/files/ehealth/docs/covid-certificate_json_specification_en.pdf
    - Value Sets for codes etc
        - https://ec.europa.eu/health/sites/default/files/ehealth/docs/digital-green-value-sets_en.pdf
        - https://github.com/ehn-dcc-development/ehn-dcc-schema/tree/release/1.3.0/valuesets

## CBOR Specifications
Used the following CBOR related specifications to unpack the credential

- Concise Binary Object Representation (CBOR)
    - `https://datatracker.ietf.org/doc/html/rfc7049`
- CBOR Object Signing and Encryption (COSE)
    - `https://datatracker.ietf.org/doc/html/rfc8152`
    - certificate uses COSE Single Signer (COSE_Sign1), which has a CBOR tag of 18
- CBOR Web Token (CWT)
    - `https://datatracker.ietf.org/doc/html/rfc8392`
- Decode CBOR tags
    - `https://datatracker.ietf.org/doc/html/draft-bormann-cbor-notable-tags-01`


# Appendix CLI Verbose Output
    
## Example Verbose 1 Output
```
go run . -qrfile ./testfiles/dcc-testdata/DE/2DCode/png/1.png -verbose 1
Decoding EU Covid-19 Certificate
  qrCodefile=./testfiles/vaccine/dr_1.png  ValueSetPath=./valuesetdata  verbose=1
  Step 1 - Read QR Code PNG ./testfiles/vaccine/dr_1.png Successfully...
  Step 2 - Base45 Decoded Successfully...
  Step 3 - ZLIB Inflated Successfully...
  Step 4 - CBOR UnMarshalled CBOR Web Token (CWT) using COSE tagged message COSE Number=18 Successfully...
    CWT CBOR UnMarshalled ProtectedHeader Successfully...
    CWT Read UnProtectedHeader Successfully...
    CWT CBOR UnMarshalled Payload Successfully...
    CWT Read COSE Signature (single signer) Successfully...
Successfully Decoded EU Covid-19 Certificate

**** EU Covid-19 Certificate Details **** 
Protected Header={
  "Alg": -7,
  "Kid": null
}
Common Payload={
  "iss": "DE",
  "iat": 1622316073,
  "exp": 1643356073,
  "hcert": {
    "1": {
      "ver": "1.0.0",
      "dob": "1964-08-12",
      "nam": {
        "fn": "Mustermann",
        "fnt": "MUSTERMANN",
        "gn": "Erika",
        "gnt": "ERIKA"
      },
      "v": [
        {
          "tg": "840539006",
          "vp": "1119349007",
          "mp": "EU/1/20/1507",
          "ma": "ORG-100031184",
          "dn": 2,
          "sd": 2,
          "dt": "2021-05-29",
          "co": "DE",
          "is": "Robert Koch-Institut",
          "ci": "URN:UVCI:01DE/IZ12345A/5CWLU12RNOB9RXSEOP6FG8#W"
        }
      ]
    }
  }
}
hex(signature)=218ebc2a2a77c1796c95a8c942987d461411b0075fd563447295250d5ead69f3b8f6083a515bd97656e87aca01529e6aa0e09144fc07e2884c93080f1419e82f

**** EU Covid-19 Certificate Summary **** 
Name:Erika Mustermann
DOB :1964-08-12
Vaccine Details
  Doses Administered: 2
  Doses Required:     2
  When:               2021-05-29
  Vaccine Product:    COVID-19 Vaccine Moderna
  Vaccine Type:       SARS-CoV-2 mRNA vaccine
  Vaccine Maker:      Moderna Biotech Spain S.L.
  Issuer:             Robert Koch-Institut
  ID:                 URN:UVCI:01DE/IZ12345A/5CWLU12RNOB9RXSEOP6FG8#W
```

## Example Verbose 2 Output
```
go run . -qrfile ./testfiles/dcc-testdata/DE/2DCode/png/1.png -verbose 2
Decoding EU Covid Certificate
  qrCodefile=./testfiles/vaccine/dr_1.png  ValueSetPath=./valuesetdata  verbose=2
  Step 1 - Read QR Code PNG ./testfiles/vaccine/dr_1.png Successfully...
    value=HC1:6BF+70790T9WJWG.FKY*4GO0.O1CV2 O5 N2FBBRW1*70HS8WY04AC*WIFN0AHCD8KD97TK0F90KECTHGWJC0FDC:5AIA%G7X+AQB9746HS80:54IBQF60R6$A80X6S1BTYACG6M+9XG8KIAWNA91AY%67092L4WJCT3EHS8XJC$+DXJCCWENF6OF63W5NW6WF6%JC QE/IAYJC5LEW34U3ET7DXC9 QE-ED8%E.JCBECB1A-:8$96646AL60A60S6Q$D.UDRYA 96NF6L/5QW6307KQEPD09WEQDD+Q6TW6FA7C466KCN9E%961A6DL6FA7D46JPCT3E5JDLA7$Q6E464W5TG6..DX%DZJC6/DTZ9 QE5$CB$DA/D JC1/D3Z8WED1ECW.CCWE.Y92OAGY8MY9L+9MPCG/D5 C5IA5N9$PC5$CUZCY$5Y$527B+A4KZNQG5TKOWWD9FL%I8U$F7O2IBM85CWOC%LEZU4R/BXHDAHN 11$CA5MRI:AONFN7091K9FKIGIY%VWSSSU9%01FO2*FTPQ3C3F
  Step 2 - Base45 Decoded Successfully...
    hex(value)=789c0163019cfed28443a10126a104480c4b15512be9140159010da401624445061a60b29429041a61f39fa9390103a101a4617681aa626369782f55524e3a555643493a303144452f495a3132333435412f3543574c553132524e4f4239525853454f5036464738235762636f62444562646e026264746a323032312d30352d323962697374526f62657274204b6f63682d496e737469747574626d616d4f52472d313030303331313834626d706c45552f312f32302f3135303762736402627467693834303533393030366276706a3131313933343930303763646f626a313936342d30382d3132636e616da462666e6a4d75737465726d616e6e62676e654572696b6163666e746a4d55535445524d414e4e63676e74654552494b416376657265312e302e305840218ebc2a2a77c1796c95a8c942987d461411b0075fd563447295250d5ead69f3b8f6083a515bd97656e87aca01529e6aa0e09144fc07e2884c93080f1419e82f1c66773a
  Step 3 - ZLIB Inflated Successfully...
    hex(value)=d28443a10126a104480c4b15512be9140159010da401624445061a60b29429041a61f39fa9390103a101a4617681aa626369782f55524e3a555643493a303144452f495a3132333435412f3543574c553132524e4f4239525853454f5036464738235762636f62444562646e026264746a323032312d30352d323962697374526f62657274204b6f63682d496e737469747574626d616d4f52472d313030303331313834626d706c45552f312f32302f3135303762736402627467693834303533393030366276706a3131313933343930303763646f626a313936342d30382d3132636e616da462666e6a4d75737465726d616e6e62676e654572696b6163666e746a4d55535445524d414e4e63676e74654552494b416376657265312e302e305840218ebc2a2a77c1796c95a8c942987d461411b0075fd563447295250d5ead69f3b8f6083a515bd97656e87aca01529e6aa0e09144fc07e2884c93080f1419e82f
  Step 4 - CBOR UnMarshalled CBOR Web Token (CWT) using COSE tagged message COSE Number=18 Successfully...
    value={Number:18 Content:[[161 1 38] map[4:[12 75 21 81 43 233 20 1]] [164 1 98 68 69 6 26 96 178 148 41 4 26 97 243 159 169 57 1 3 161 1 164 97 118 129 170 98 99 105 120 47 85 82 78 58 85 86 67 73 58 48 49 68 69 47 73 90 49 50 51 52 53 65 47 53 67 87 76 85 49 50 82 78 79 66 57 82 88 83 69 79 80 54 70 71 56 35 87 98 99 111 98 68 69 98 100 110 2 98 100 116 106 50 48 50 49 45 48 53 45 50 57 98 105 115 116 82 111 98 101 114 116 32 75 111 99 104 45 73 110 115 116 105 116 117 116 98 109 97 109 79 82 71 45 49 48 48 48 51 49 49 56 52 98 109 112 108 69 85 47 49 47 50 48 47 49 53 48 55 98 115 100 2 98 116 103 105 56 52 48 53 51 57 48 48 54 98 118 112 106 49 49 49 57 51 52 57 48 48 55 99 100 111 98 106 49 57 54 52 45 48 56 45 49 50 99 110 97 109 164 98 102 110 106 77 117 115 116 101 114 109 97 110 110 98 103 110 101 69 114 105 107 97 99 102 110 116 106 77 85 83 84 69 82 77 65 78 78 99 103 110 116 101 69 82 73 75 65 99 118 101 114 101 49 46 48 46 48] [33 142 188 42 42 119 193 121 108 149 168 201 66 152 125 70 20 17 176 7 95 213 99 68 114 149 37 13 94 173 105 243 184 246 8 58 81 91 217 118 86 232 122 202 1 82 158 106 160 224 145 68 252 7 226 136 76 147 8 15 20 25 232 47]]}
    CWT CBOR UnMarshalled ProtectedHeader Successfully...
      value={Alg:-7 Kid:[]}
    CWT Read UnProtectedHeader Successfully...
      value={Alg:0 Kid:[12 75 21 81 43 233 20 1]}
    CWT CBOR UnMarshalled Payload Successfully...
      value=map[-260:map[1:map[dob:1964-08-12 nam:map[fn:Mustermann fnt:MUSTERMANN gn:Erika gnt:ERIKA] v:[map[ci:URN:UVCI:01DE/IZ12345A/5CWLU12RNOB9RXSEOP6FG8#W co:DE dn:2 dt:2021-05-29 is:Robert Koch-Institut ma:ORG-100031184 mp:EU/1/20/1507 sd:2 tg:840539006 vp:1119349007]] ver:1.0.0]] 1:DE 4:1643356073 6:1622316073]
    CWT Read COSE Signature (single signer) Successfully...
      hex(value)=218ebc2a2a77c1796c95a8c942987d461411b0075fd563447295250d5ead69f3b8f6083a515bd97656e87aca01529e6aa0e09144fc07e2884c93080f1419e82f
Successfully Decoded EU Covid Certificate

**** EU Covid-19 Certificate Summary **** 
Name:Erika Mustermann
DOB :1964-08-12
Vaccine Details
  Doses Administered: 2
  Doses Required:     2
  When:               2021-05-29
  Vaccine Product:    COVID-19 Vaccine Moderna
  Vaccine Type:       SARS-CoV-2 mRNA vaccine
  Vaccine Maker:      Moderna Biotech Spain S.L.
  Issuer:             Robert Koch-Institut
  ID:                 URN:UVCI:01DE/IZ12345A/5CWLU12RNOB9RXSEOP6FG8#W
```

# Development
- Go version >= 1.14
- Is a module  
- Make targets
    - `make test` lint and run tests
- Files
    - decoder.go - main
    - helper - code to decode and display certificates
    - datamodel - the certificate structs
    - valuesetdata - copies of valueset data from https://github.com/ehn-dcc-development/ehn-dcc-schema/tree/release/1.3.0/valuesets
    - testfiles - example qr code png from https://github.com/eu-digital-green-certificates/dgc-testdata
