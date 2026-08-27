# go-response

A lightweight HTTP response library built specifically for [Gin](https://github.com/gin-gonic/gin).

go-response standardizes JSON responses across your Go services so every
endpoint produces the same predictable shape:

- Success (`200 OK`, `201 Created`, `204 No Content`)
- Validation errors (`422`)
- Application errors (`400`–`500`)
- Pagination

## Installation

```sh
go get github.com/helloirfanaditya/go-response
```

Requires Go 1.25+.

## Quick Start

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/helloirfanaditya/go-response"
)

type createUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func main() {
	r := gin.Default()

	r.POST("/users", func(c *gin.Context) {
		var req createUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Validation(c, &req, err)
			return
		}
		response.Created(c, gin.H{"id": 1})
	})

	r.GET("/users/:id", func(c *gin.Context) {
		response.Error(c, response.NotFound("user not found"))
	})

	r.Run(":8080")
}
```

## Response Formats

Every response follows the same envelope.

### Success

```json
{
    "success": true,
    "code": "SUCCESS",
    "message": "Success",
    "data": {}
}
```

### Error

```json
{
    "success": false,
    "code": "NOT_FOUND",
    "message": "User not found"
}
```

### Validation

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

### Pagination

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

Unused fields are never included in the payload.

## Functions

| Function | HTTP | Code |
|---|---|---|
| `response.Success(c, data)` | 200 | `SUCCESS` |
| `response.Created(c, data)` | 201 | `CREATED` |
| `response.NoContent(c)` | 204 | — (no body) |
| `response.Validation(c, &req, err)` | 422 | `VALIDATION_ERROR` |
| `response.NewValidationError(msg, fields...)` | 422 | `VALIDATION_ERROR` |
| `response.Error(c, err)` | mapped | mapped |
| `response.Paginate(c, data, meta)` | 200 | `SUCCESS` |

## Errors

`response.Error` converts application errors into an HTTP status, response
code, and message. Errors that implement `ErrorCoder` are mapped to their own
status and code, which keeps the mapping extensible.

Built-in constructors:

| Constructor | HTTP | Code |
|---|---|---|
| `response.BadRequest(msg)` | 400 | `BAD_REQUEST` |
| `response.Unauthorized(msg)` | 401 | `UNAUTHORIZED` |
| `response.Forbidden(msg)` | 403 | `FORBIDDEN` |
| `response.NotFound(msg)` | 404 | `NOT_FOUND` |
| `response.Conflict(msg)` | 409 | `CONFLICT` |
| `response.UnprocessableEntity(msg)` | 422 | `VALIDATION_ERROR` |
| `response.InternalServerError(msg)` | 500 | `INTERNAL_SERVER_ERROR` |

Any other error falls back to `500 INTERNAL_SERVER_ERROR`.

Custom error types only need to implement `ErrorCoder`:

```go
type rateLimitedError struct{}

func (rateLimitedError) Error() string        { return "too many requests" }
func (rateLimitedError) StatusCode() int      { return 429 }
func (rateLimitedError) ResponseCode() string { return "RATE_LIMITED" }

// in a handler:
response.Error(c, rateLimitedError{})
```

## Domain-Specific Error Codes

The generic constructors (`BadRequest`, `NotFound`, ...) are convenient, but
for a production API you usually want **stable, machine-readable codes per
domain** — clients use them for localization and UI logic, so they must never
change between releases.

Use `response.NewError(status, code, message)` to build errors with your own
code:

```go
import "net/http"

var (
	ErrInvalidAmount       = response.NewError(http.StatusBadRequest, "TRANSACTION_INVALID_AMOUNT", "amount must be greater than zero")
	ErrInsufficientBalance = response.NewError(http.StatusBadRequest, "TRANSACTION_INSUFFICIENT_BALANCE", "insufficient balance")
	ErrRecipientNotFound   = response.NewError(http.StatusNotFound, "TRANSACTION_RECIPIENT_NOT_FOUND", "recipient account not found")
	ErrStoreFailed         = response.NewError(http.StatusInternalServerError, "TRANSACTION_STORE_FAILED", "failed to store transaction")
)
```

```go
// in a handler:
resp, err := h.svc.Transfer(req)
if err != nil {
	response.Error(c, err)
	return
}
response.Success(c, gin.H{"transaction_id": resp})
```

Response:

```json
{
	"success": false,
	"code": "TRANSACTION_INSUFFICIENT_BALANCE",
	"message": "insufficient balance"
}
```

Rules of thumb:

- Codes are `UPPER_SNAKE_CASE` and prefixed by domain (`TRANSACTION_`, `USER_`, `ORDER_`, ...)
- `code` is the contract between server and client — never parse `message`
- Server-side failures (storage, third-party APIs) should map to `500`, not `400`

### Field-Level Errors (Business Rules)

Validator tags can't express every rule. When business logic rejects a request
with reasons tied to specific fields, build the error with `[]FieldError`
using `response.NewValidationError`:

```go
response.Error(c, response.NewValidationError("transfer rejected",
	response.FieldError{
		Field: "amount",
		Code:  "TRANSACTION_EXCEEDS_LIMIT",
		Params: map[string]any{"max": 5000000},
	},
	response.FieldError{
		Field: "recipient",
		Code:  "TRANSACTION_ACCOUNT_BLOCKED",
	},
))
```

Response:

```json
{
	"success": false,
	"code": "VALIDATION_ERROR",
	"message": "transfer rejected",
	"errors": [
		{
			"field": "amount",
			"code": "TRANSACTION_EXCEEDS_LIMIT",
			"params": {
				"max": 5000000
			}
		},
		{
			"field": "recipient",
			"code": "TRANSACTION_ACCOUNT_BLOCKED"
		}
	]
}
```

`FieldError` has the same shape as the validation errors produced by
`response.Validation`, so clients handle both identically.

Custom errors can also carry field errors by implementing `ErrorCoder` **and**
`FieldErrorProvider` (`FieldErrors() []FieldError`) — `response.Error` picks
both up automatically.

## Validation

`response.Validation` uses
[go-playground/validator/v10](https://github.com/go-playground/validator/v10).
It MUST receive a pointer to the request struct.

Field names are resolved from struct tags in this priority order:

1. `json`
2. `form`
3. `uri`
4. `query`
5. `header`
6. struct field name

```go
Email string `json:"email"` // → "email", never "Email"
```

Validation codes are stable and never localized — translation is the client's
responsibility. Every built-in validator tag maps to a code:

| Code | Validator tags |
|---|---|
| `ERR_REQUIRED` | `required`, `required_if`, `required_unless`, `skip_unless`, `required_with`, `required_with_all`, `required_without`, `required_without_all`, `excluded_if`, `excluded_unless`, `excluded_with`, `excluded_with_all`, `excluded_without`, `excluded_without_all` |
| `ERR_LENGTH` | `len` |
| `ERR_MIN_LENGTH` | `min` on strings and collections |
| `ERR_MAX_LENGTH` | `max` on strings and collections |
| `ERR_MIN_VALUE` | `min` on numbers |
| `ERR_MAX_VALUE` | `max` on numbers |
| `ERR_OUT_OF_RANGE` | `gt`, `gte`, `lt`, `lte` |
| `ERR_INVALID_VALUE` | `eq`, `eq_ignore_case`, `ne`, `ne_ignore_case`, `isdefault` |
| `ERR_MISMATCH` | `eqfield`, `nefield`, `gtfield`, `gtefield`, `ltfield`, `ltefield`, `eqcsfield`, `necsfield`, `gtcsfield`, `gtecsfield`, `ltcsfield`, `ltecsfield`, `fieldcontains`, `fieldexcludes` |
| `ERR_INVALID_ENUM` | `oneof`, `oneofci`, `noneof`, `noneofci` |
| `ERR_DUPLICATE` | `unique` |
| `ERR_INVALID_CONTENT` | `contains`, `containsany`, `containsrune`, `excludes`, `excludesall`, `excludesrune`, `startswith`, `endswith`, `startsnotwith`, `endsnotwith` |
| `ERR_INVALID_EMAIL` | `email` |
| `ERR_INVALID_URL` | `url`, `http_url`, `https_url` |
| `ERR_INVALID_URI` | `uri` |
| `ERR_INVALID_ORIGIN` | `origin` |
| `ERR_INVALID_URN` | `urn_rfc2141` |
| `ERR_INVALID_UUID` | `uuid`, `uuid3`, `uuid4`, `uuid5`, `uuid_rfc4122`, `uuid3_rfc4122`, `uuid4_rfc4122`, `uuid5_rfc4122` |
| `ERR_INVALID_ULID` | `ulid` |
| `ERR_INVALID_IP` | `ip`, `ipv4`, `ipv6`, `cidr`, `cidrv4`, `cidrv6`, `ip_addr`, `ip4_addr`, `ip6_addr`, `tcp_addr`, `tcp4_addr`, `tcp6_addr`, `udp_addr`, `udp4_addr`, `udp6_addr`, `unix_addr`, `uds_exists` |
| `ERR_INVALID_MAC` | `mac` |
| `ERR_INVALID_NUMBER` | `numeric`, `number` |
| `ERR_INVALID_BOOLEAN` | `boolean` |
| `ERR_INVALID_DATETIME` | `datetime` |
| `ERR_INVALID_TIMEZONE` | `timezone` |
| `ERR_INVALID_E164` | `e164` |
| `ERR_INVALID_ALPHA` | `alpha`, `alphaspace` |
| `ERR_INVALID_ALPHANUM` | `alphanum`, `alphanumspace` |
| `ERR_INVALID_ALPHA_UNICODE` | `alphaunicode` |
| `ERR_INVALID_ALPHANUM_UNICODE` | `alphanumunicode` |
| `ERR_INVALID_ASCII` | `ascii` |
| `ERR_INVALID_PRINTABLE_ASCII` | `printascii` |
| `ERR_INVALID_MULTIBYTE` | `multibyte` |
| `ERR_INVALID_LOWERCASE` | `lowercase` |
| `ERR_INVALID_UPPERCASE` | `uppercase` |
| `ERR_INVALID_HEXADECIMAL` | `hexadecimal` |
| `ERR_INVALID_HEX_COLOR` | `hexcolor` |
| `ERR_INVALID_RGB` | `rgb` |
| `ERR_INVALID_RGBA` | `rgba` |
| `ERR_INVALID_HSL` | `hsl` |
| `ERR_INVALID_HSLA` | `hsla` |
| `ERR_INVALID_CMYK` | `cmyk` |
| `ERR_INVALID_COLOR` | `iscolor` |
| `ERR_INVALID_BASE32` | `base32` |
| `ERR_INVALID_BASE64` | `base64` |
| `ERR_INVALID_BASE64_URL` | `base64url` |
| `ERR_INVALID_BASE64_RAW_URL` | `base64rawurl` |
| `ERR_INVALID_DATA_URI` | `datauri` |
| `ERR_INVALID_JSON` | `json` |
| `ERR_INVALID_JWT` | `jwt` |
| `ERR_INVALID_HTML` | `html` |
| `ERR_INVALID_HTML_ENCODED` | `html_encoded` |
| `ERR_INVALID_URL_ENCODED` | `url_encoded` |
| `ERR_INVALID_FILE` | `file` |
| `ERR_INVALID_FILE_PATH` | `filepath` |
| `ERR_INVALID_DIR` | `dir` |
| `ERR_INVALID_DIR_PATH` | `dirpath` |
| `ERR_INVALID_IMAGE` | `image` |
| `ERR_INVALID_MIME_TYPE` | `mimetype` |
| `ERR_INVALID_ISBN` | `isbn`, `isbn10`, `isbn13` |
| `ERR_INVALID_ISSN` | `issn` |
| `ERR_INVALID_CREDIT_CARD` | `credit_card` |
| `ERR_INVALID_LUHN` | `luhn_checksum` |
| `ERR_INVALID_CVE` | `cve` |
| `ERR_INVALID_SEMVER` | `semver` |
| `ERR_INVALID_HOSTNAME` | `hostname`, `hostname_rfc1123` |
| `ERR_INVALID_FQDN` | `fqdn` |
| `ERR_INVALID_HOSTNAME_PORT` | `hostname_port` |
| `ERR_INVALID_PORT` | `port` |
| `ERR_INVALID_DNS_LABEL` | `dns_rfc1035_label` |
| `ERR_INVALID_LATITUDE` | `latitude` |
| `ERR_INVALID_LONGITUDE` | `longitude` |
| `ERR_INVALID_SSN` | `ssn` |
| `ERR_INVALID_ETH_ADDR` | `eth_addr`, `eth_addr_checksum` |
| `ERR_INVALID_BTC_ADDR` | `btc_addr`, `btc_addr_bech32` |
| `ERR_INVALID_MD4` | `md4` |
| `ERR_INVALID_MD5` | `md5` |
| `ERR_INVALID_SHA256` | `sha256` |
| `ERR_INVALID_SHA384` | `sha384` |
| `ERR_INVALID_SHA512` | `sha512` |
| `ERR_INVALID_RIPEMD128` | `ripemd128` |
| `ERR_INVALID_RIPEMD160` | `ripemd160` |
| `ERR_INVALID_TIGER128` | `tiger128` |
| `ERR_INVALID_TIGER160` | `tiger160` |
| `ERR_INVALID_TIGER192` | `tiger192` |
| `ERR_INVALID_ISO3166_ALPHA2` | `iso3166_1_alpha2`, `iso3166_1_alpha2_eu` |
| `ERR_INVALID_ISO3166_ALPHA3` | `iso3166_1_alpha3`, `iso3166_1_alpha3_eu` |
| `ERR_INVALID_ISO3166_ALPHA_NUMERIC` | `iso3166_1_alpha_numeric`, `iso3166_1_alpha_numeric_eu` |
| `ERR_INVALID_ISO3166_2` | `iso3166_2` |
| `ERR_INVALID_COUNTRY_CODE` | `country_code`, `eu_country_code` |
| `ERR_INVALID_ISO4217` | `iso4217`, `iso4217_numeric` |
| `ERR_INVALID_LANGUAGE_TAG` | `bcp47_language_tag`, `bcp47_strict_language_tag` |
| `ERR_INVALID_POSTAL_CODE` | `postcode_iso3166_alpha2`, `postcode_iso3166_alpha2_field` |
| `ERR_INVALID_BIC` | `bic`, `bic_iso_9362_2014` |
| `ERR_INVALID_MONGODB_ID` | `mongodb` |
| `ERR_INVALID_MONGODB_CONNECTION` | `mongodb_connection_string` |
| `ERR_INVALID_CRON` | `cron` |
| `ERR_INVALID_SPICEDB` | `spicedb` |
| `ERR_INVALID_EIN` | `ein` |
| `ERR_INVALID` | custom tags registered with `RegisterValidation` |

Comparison tags (`len`, `min`, `max`, `eq`, `ne`, `lt`, `lte`, `gt`, `gte`),
cross-field tags, `oneof`, `datetime`, and content tags carry their bound or
reference in `params`.

### Conditional and Cross-Field Validation

Rules can depend on other fields in the same request. Everything below stays
machine-readable and maps to the same code table.

**Conditional presence** — a field is only required when another field has a
certain value:

```go
type paymentRequest struct {
	Method string `json:"method" validate:"required,oneof=transfer credit_card"`
	Card   string `json:"card" validate:"required_if=Method credit_card"`
}
```

Request `{"method": "credit_card"}`:

```json
{
	"success": false,
	"code": "VALIDATION_ERROR",
	"message": "Validation failed",
	"errors": [
		{
			"field": "card",
			"code": "ERR_REQUIRED"
		}
	]
}
```

**Cross-field comparison** — a field must match (or differ from) another field:

```go
type registerRequest struct {
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}
```

Request `{"password": "supersecret", "confirm_password": "different"}`:

```json
{
	"success": false,
	"code": "VALIDATION_ERROR",
	"message": "Validation failed",
	"errors": [
		{
			"field": "confirm_password",
			"code": "ERR_MISMATCH",
			"params": {
				"eqfield": "Password"
			}
		}
	]
}
```

**Custom tags** — validators registered with `RegisterValidation` map to
`ERR_INVALID`:

```go
// in init():
v := binding.Validator.Engine().(*validator.Validate)
_ = v.RegisterValidation("no_spaces", func(fl validator.FieldLevel) bool {
	return !strings.Contains(fl.Field().String(), " ")
})

type usernameRequest struct {
	Username string `json:"username" validate:"required,no_spaces"`
}
```

Request `{"username": "has space"}`:

```json
{
	"success": false,
	"code": "VALIDATION_ERROR",
	"message": "Validation failed",
	"errors": [
		{
			"field": "username",
			"code": "ERR_INVALID"
		}
	]
}
```

> Both `binding:"required"` (Gin) and `validate:"required"` (validator) produce
> `validator.ValidationErrors`, so `response.Validation` handles them the same
> way. Prefer `validate` tags so the mapping table applies uniformly.

## Pagination

```go
response.Paginate(c, users, response.Meta{
    Page:    1,
    PerPage: 10,
    Total:   100,
})
```

When `TotalPage` is zero it is computed as `ceil(total / perPage)`.

## Development

```sh
go test ./...
```
