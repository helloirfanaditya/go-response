# go-response Specification

Version: 0.1.0

---

# Overview

go-response is a lightweight HTTP response library built specifically for Gin.

The purpose of this library is to provide a consistent and reusable JSON response format across Go backend services.

The library focuses on one responsibility only:

**Standardizing HTTP responses.**

---

# Goals

The library must:

- Follow idiomatic Go.
- Minimize boilerplate.
- Be production-ready.
- Keep the public API small.
- Produce predictable JSON responses.
- Be easy to maintain.
- Remain backward compatible whenever possible.

---

# Non Goals

This library MUST NOT provide:

- JWT
- Authentication
- Authorization
- Database
- ORM
- Logging
- Configuration
- Dependency Injection
- Middleware
- HTTP Client

---

# Supported Framework

Current supported framework:

- Gin

---

# Public API

The public API should remain as small as possible.

```go
response.Success(c, data)

response.Created(c, data)

response.NoContent(c)

response.Validation(c, &request, err)

response.Error(c, err)

response.Paginate(c, data, meta)
```

Avoid adding new exported functions unless absolutely necessary.

---

# JSON Structure

Every response MUST follow the same structure.

## Success Response

```json
{
    "success": true,
    "code": "SUCCESS",
    "message": "Success",
    "data": {}
}
```

---

## Error Response

```json
{
    "success": false,
    "code": "NOT_FOUND",
    "message": "User not found"
}
```

---

## Validation Response

```json
{
    "success": false,
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "errors": [
        {
            "field": "email",
            "code": "ERR_REQUIRED"
        },
        {
            "field": "password",
            "code": "ERR_MIN_LENGTH",
            "params": {
                "min": 8
            }
        }
    ]
}
```

---

## Pagination Response

```json
{
    "success": true,
    "code": "SUCCESS",
    "message": "Success",
    "data": [],
    "meta": {
        "page": 1,
        "perPage": 10,
        "total": 100,
        "totalPage": 10
    }
}
```

---

# JSON Rules

The response object has the following fields.

| Field | Required | Description |
|--------|----------|-------------|
| success | Yes | Indicates whether the request succeeded. |
| code | Yes | Stable machine-readable response code. |
| message | Optional | Human-readable message. |
| data | Success only | Response payload. |
| errors | Validation only | Validation error list. |
| meta | Pagination only | Pagination metadata. |

Unused fields MUST NOT be included in the JSON response.

---

# Success Responses

Supported responses:

## Success

HTTP Status

200 OK

Code

```
SUCCESS
```

---

## Created

HTTP Status

201 Created

Code

```
CREATED
```

---

## No Content

HTTP Status

204 No Content

No response body.

---

# Validation

Validation uses:

```
github.com/go-playground/validator/v10
```

The Validation function MUST receive a pointer to the request struct.

Example:

```go
var req CreateUserRequest

if err := c.ShouldBindJSON(&req); err != nil {
    response.Validation(c, &req, err)
    return
}
```

Passing a non-pointer request is unsupported.

---

# Field Name Resolution

Validation errors must use request tags instead of Go struct names.

Priority:

1. json
2. form
3. uri
4. query
5. header
6. struct field name

Example:

```go
Email string `json:"email"`
```

↓

```json
{
    "field": "email",
    "code": "ERR_REQUIRED"
}
```

Never return:

```json
{
    "field": "Email"
}
```

---

# Validation Codes

Validation responses MUST use standardized error codes.

The full set of codes, covering every built-in validator tag:

```
ERR_REQUIRED

ERR_LENGTH

ERR_MIN_LENGTH

ERR_MAX_LENGTH

ERR_MIN_VALUE

ERR_MAX_VALUE

ERR_OUT_OF_RANGE

ERR_INVALID_VALUE

ERR_MISMATCH

ERR_INVALID_ENUM

ERR_DUPLICATE

ERR_INVALID_CONTENT

ERR_INVALID_EMAIL

ERR_INVALID_URL

ERR_INVALID_URI

ERR_INVALID_ORIGIN

ERR_INVALID_URN

ERR_INVALID_UUID

ERR_INVALID_ULID

ERR_INVALID_IP

ERR_INVALID_MAC

ERR_INVALID_NUMBER

ERR_INVALID_BOOLEAN

ERR_INVALID_DATETIME

ERR_INVALID_TIMEZONE

ERR_INVALID_E164

ERR_INVALID_ALPHA

ERR_INVALID_ALPHANUM

ERR_INVALID_ALPHA_UNICODE

ERR_INVALID_ALPHANUM_UNICODE

ERR_INVALID_ASCII

ERR_INVALID_PRINTABLE_ASCII

ERR_INVALID_MULTIBYTE

ERR_INVALID_LOWERCASE

ERR_INVALID_UPPERCASE

ERR_INVALID_HEXADECIMAL

ERR_INVALID_HEX_COLOR

ERR_INVALID_RGB

ERR_INVALID_RGBA

ERR_INVALID_HSL

ERR_INVALID_HSLA

ERR_INVALID_CMYK

ERR_INVALID_COLOR

ERR_INVALID_BASE32

ERR_INVALID_BASE64

ERR_INVALID_BASE64_URL

ERR_INVALID_BASE64_RAW_URL

ERR_INVALID_DATA_URI

ERR_INVALID_JSON

ERR_INVALID_JWT

ERR_INVALID_HTML

ERR_INVALID_HTML_ENCODED

ERR_INVALID_URL_ENCODED

ERR_INVALID_FILE

ERR_INVALID_FILE_PATH

ERR_INVALID_DIR

ERR_INVALID_DIR_PATH

ERR_INVALID_IMAGE

ERR_INVALID_MIME_TYPE

ERR_INVALID_ISBN

ERR_INVALID_ISSN

ERR_INVALID_CREDIT_CARD

ERR_INVALID_LUHN

ERR_INVALID_CVE

ERR_INVALID_SEMVER

ERR_INVALID_HOSTNAME

ERR_INVALID_FQDN

ERR_INVALID_HOSTNAME_PORT

ERR_INVALID_PORT

ERR_INVALID_DNS_LABEL

ERR_INVALID_LATITUDE

ERR_INVALID_LONGITUDE

ERR_INVALID_SSN

ERR_INVALID_ETH_ADDR

ERR_INVALID_BTC_ADDR

ERR_INVALID_MD4

ERR_INVALID_MD5

ERR_INVALID_SHA256

ERR_INVALID_SHA384

ERR_INVALID_SHA512

ERR_INVALID_RIPEMD128

ERR_INVALID_RIPEMD160

ERR_INVALID_TIGER128

ERR_INVALID_TIGER160

ERR_INVALID_TIGER192

ERR_INVALID_ISO3166_ALPHA2

ERR_INVALID_ISO3166_ALPHA3

ERR_INVALID_ISO3166_ALPHA_NUMERIC

ERR_INVALID_ISO3166_2

ERR_INVALID_COUNTRY_CODE

ERR_INVALID_ISO4217

ERR_INVALID_LANGUAGE_TAG

ERR_INVALID_POSTAL_CODE

ERR_INVALID_BIC

ERR_INVALID_MONGODB_ID

ERR_INVALID_MONGODB_CONNECTION

ERR_INVALID_CRON

ERR_INVALID_SPICEDB

ERR_INVALID_EIN

ERR_INVALID
```

The mapping from validator tags to codes is documented in the README.

Validation responses MUST NOT contain translated messages.

Localization is the responsibility of the client.

---

# Error Responses

response.Error(c, err)

must automatically convert application errors into:

- HTTP Status
- Response Code
- Response Message

Example:

| Error | HTTP |
|--------|------|
| BadRequest | 400 |
| Unauthorized | 401 |
| Forbidden | 403 |
| NotFound | 404 |
| Conflict | 409 |
| Validation | 422 |
| Internal | 500 |

The mapping mechanism should be extensible.

---

# Pagination

Pagination metadata:

```go
type Meta struct {
    Page      int
    PerPage   int
    Total     int64
    TotalPage int
}
```

Example:

```go
response.Paginate(c, users, meta)
```

---

# Error Code Philosophy

There are two levels of error codes.

Response Code

Example:

```
SUCCESS

NOT_FOUND

VALIDATION_ERROR

CONFLICT

INTERNAL_SERVER_ERROR
```

Validation Code

Example:

```
ERR_REQUIRED

ERR_INVALID_EMAIL

ERR_MIN_LENGTH
```

Response codes describe the request outcome.

Validation codes describe individual field validation failures.

---

# Design Principles

The library should always follow:

- Idiomatic Go
- Small Public API
- Explicit Behavior
- Predictable Output
- Minimal Dependencies
- Production Ready

Avoid:

- Magic
- Reflection where unnecessary
- Hidden side effects
- Global mutable state

---

# Documentation

Every exported type must include GoDoc.

Every exported function must include GoDoc.

README examples must always match the implementation.

---

# Testing

Every exported function must include unit tests.

Public API changes require new tests.

---

# Versioning

Semantic Versioning.

Major versions may introduce breaking changes.

Minor versions must remain backward compatible.

---

# Philosophy

go-response solves one problem.

Standardized HTTP responses.

It should remain lightweight.

It should never become a framework.
