# Overview
This CLI tool decodes an EU Digital COVID Certificate (also referred to as Digital Green Certificates), it does NOT
yet verify the Signature.

# How to run
The CLI flags are
1. -qrc_file <value> file containing the qr code png
2. -verbose <level> where level is 0 -> 9, default is zero

Example running with no verbose
- `go run . -qrfile ./testfiles/vaccine/de_1.png`            <-- no verbose
- `go run . -qrfile ./testfiles/vaccine/de_1.png -verbose 0` <-- no verbose
- `go run . -qrfile ./testfiles/vaccine/de_1.png -verbose 1` <-- displayed details on protected header and common payload
- `go run . -qrfile ./testfiles/vaccine/de_1.png -verbose 2` <-- display details on each decoding step  

No verbose output
```
go run . -qrfile ./testfiles/vaccine/dr_1.png
Decoding EU Covid Certificate
  qrCodefile=./testfiles/vaccine/dr_1.png  ValueSetPath=./valuesetdata  verbose=0
  Step 1 - Read QR Code PNG ./testfiles/vaccine/dr_1.png Successfully...
  Step 2 - Base45 Decoded Successfully...
  Step 3 - ZLIB Inflated Successfully...
  Step 4 - CBOR UnMarshalled CBOR Web Token (CWT) Successfully...
    CWT CBOR UnMarshalled ProtectedHeader Successfully...
    CWT Read UnProtectedHeader Successfully...
    CWT CBOR UnMarshalled Payload Successfully...
    CWT Read COSE Signature (single signer) Successfully...
Successfully Decoded EU Covid Certificate

**** EU Covid Certificate Summary **** 
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

# Decoding Steps
The **decoding steps** are as follows:
1. Read the QR code (.png) containing the Digital Certificate to get a base45 encoded certificate
2. Decode the base45 certificate to get a compressed certificate
3. ZLIB inflate the compressed certificate to get a CBOR Web Token
4. CBOR decode the CBOR Web Token to get the protected header, unprotected header, payload, and signature
5. CBOR decode the protected header to get the Signing Algorithm and KeyID
6. CBOR decode the payload to get the issuer, iat, exp, subject information, and vaccination information
7. NOT implemented check the COSE signature by getting signing key from issuing State and using it to check the CBOR signature.

Limitations
- Only pretty prints vaccine credentials not test or recovery, to see later use verbose mode
- Does **NOT** verify signature

Testing
- Test QR.png(s) are from `https://github.com/eu-digital-green-certificates/dgc-testdata`
- `make test` runs local tests

# Supporting EU Documents and Code

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

4. EU Support 
    - Certificate JSON Schema
        - https://github.com/ehn-dcc-development/ehn-dcc-schema
        - https://ec.europa.eu/health/sites/default/files/ehealth/docs/covid-certificate_json_specification_en.pdf
    - Value Sets for codes etc
        - https://ec.europa.eu/health/sites/default/files/ehealth/docs/digital-green-value-sets_en.pdf
        - https://github.com/ehn-dcc-development/ehn-dcc-schema/tree/release/1.3.0/valuesets
    
# CBOR Web Token Specifications
The certificate is a CBOR Web Token so used the following to unpack

- CBOR spec -  Concise Binary Object Representation (CBOR) -  CBOR Web Token (CWT)
    - `https://datatracker.ietf.org/doc/html/rfc7049`
- COSE spec -  CBOR Object Signing and Encryption (COSE)
    - `https://datatracker.ietf.org/doc/html/rfc8152`
    - certificate uses COSE Single Signer (COSE_Sign1), which has a CBOR tag of 18
- CBOR web token
    - `https://datatracker.ietf.org/doc/html/rfc8392`
- Decode CBOR tags
    - `https://datatracker.ietf.org/doc/html/draft-bormann-cbor-notable-tags-01`

# Verbose Output
```
go run . -qrfile ./testfiles/vaccine/dr_1.png -verbose 1
Decoding EU Covid Certificate
  qrCodefile=./testfiles/vaccine/dr_1.png  ValueSetPath=./valuesetdata  verbose=1
  Step 1 - Read QR Code PNG ./testfiles/vaccine/dr_1.png Successfully...
  Step 2 - Base45 Decoded Successfully...
  Step 3 - ZLIB Inflated Successfully...
  Step 4 - CBOR UnMarshalled CBOR Web Token (CWT) Successfully...
    CWT CBOR UnMarshalled ProtectedHeader Successfully...
    CWT Read UnProtectedHeader Successfully...
    CWT CBOR UnMarshalled Payload Successfully...
    CWT Read COSE Signature (single signer) Successfully...
Successfully Decoded EU Covid Certificate

**** EU Covid Certificate Details **** 
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

**** EU Covid Certificate Summary **** 
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
