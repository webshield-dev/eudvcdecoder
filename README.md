# Overview
This CLI tool decodes an EU Digital COVID Certificate (also referred to as Digital Green Certificates), it does NOT
yet verify the Signature.

The CLI flags are
1. -qrc_file <value> file containing the qr code png
2. -verbose <level> where level is 0 -> 9, default is zero

Example running with no verbose
- `go run . -qrfile ./testfiles/vaccine/at_1.png`
- `go run . -qrfile ./testfiles/vaccine/ie_1_qr.png`

Example running with verbose

    `go run . -qrfile ./testfiles/vaccine/ie_1_qr.png -verbose 1`


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

Example output

```json

```

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

