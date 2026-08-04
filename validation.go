package response

import (
	"errors"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Field codes produced by Validation. They are stable machine-readable
// identifiers that clients use for localization. Every built-in validator
// tag maps to one of these codes.
const (
	// Presence.
	fieldCodeRequired = "ERR_REQUIRED"

	// Length and value constraints.
	fieldCodeLength         = "ERR_LENGTH"
	fieldCodeMinLength      = "ERR_MIN_LENGTH"
	fieldCodeMaxLength      = "ERR_MAX_LENGTH"
	fieldCodeMinValue       = "ERR_MIN_VALUE"
	fieldCodeMaxValue       = "ERR_MAX_VALUE"
	fieldCodeOutOfRange     = "ERR_OUT_OF_RANGE"
	fieldCodeInvalidValue   = "ERR_INVALID_VALUE"
	fieldCodeMismatch       = "ERR_MISMATCH"
	fieldCodeInvalidEnum    = "ERR_INVALID_ENUM"
	fieldCodeDuplicate      = "ERR_DUPLICATE"
	fieldCodeInvalidContent = "ERR_INVALID_CONTENT"

	// Formats.
	fieldCodeInvalidEmail           = "ERR_INVALID_EMAIL"
	fieldCodeInvalidURL             = "ERR_INVALID_URL"
	fieldCodeInvalidURI             = "ERR_INVALID_URI"
	fieldCodeInvalidOrigin          = "ERR_INVALID_ORIGIN"
	fieldCodeInvalidURN             = "ERR_INVALID_URN"
	fieldCodeInvalidUUID            = "ERR_INVALID_UUID"
	fieldCodeInvalidULID            = "ERR_INVALID_ULID"
	fieldCodeInvalidIP              = "ERR_INVALID_IP"
	fieldCodeInvalidMAC             = "ERR_INVALID_MAC"
	fieldCodeInvalidNumber          = "ERR_INVALID_NUMBER"
	fieldCodeInvalidBoolean         = "ERR_INVALID_BOOLEAN"
	fieldCodeInvalidDatetime        = "ERR_INVALID_DATETIME"
	fieldCodeInvalidTimezone        = "ERR_INVALID_TIMEZONE"
	fieldCodeInvalidE164            = "ERR_INVALID_E164"
	fieldCodeInvalidAlpha           = "ERR_INVALID_ALPHA"
	fieldCodeInvalidAlphanum        = "ERR_INVALID_ALPHANUM"
	fieldCodeInvalidAlphaUnicode    = "ERR_INVALID_ALPHA_UNICODE"
	fieldCodeInvalidAlphanumUnicode = "ERR_INVALID_ALPHANUM_UNICODE"
	fieldCodeInvalidASCII           = "ERR_INVALID_ASCII"
	fieldCodeInvalidPrintableASCII  = "ERR_INVALID_PRINTABLE_ASCII"
	fieldCodeInvalidMultibyte       = "ERR_INVALID_MULTIBYTE"
	fieldCodeInvalidLowercase       = "ERR_INVALID_LOWERCASE"
	fieldCodeInvalidUppercase       = "ERR_INVALID_UPPERCASE"
	fieldCodeInvalidHexadecimal     = "ERR_INVALID_HEXADECIMAL"
	fieldCodeInvalidHexColor        = "ERR_INVALID_HEX_COLOR"
	fieldCodeInvalidRGB             = "ERR_INVALID_RGB"
	fieldCodeInvalidRGBA            = "ERR_INVALID_RGBA"
	fieldCodeInvalidHSL             = "ERR_INVALID_HSL"
	fieldCodeInvalidHSLA            = "ERR_INVALID_HSLA"
	fieldCodeInvalidCMYK            = "ERR_INVALID_CMYK"
	fieldCodeInvalidColor           = "ERR_INVALID_COLOR"
	fieldCodeInvalidBase32          = "ERR_INVALID_BASE32"
	fieldCodeInvalidBase64          = "ERR_INVALID_BASE64"
	fieldCodeInvalidBase64URL       = "ERR_INVALID_BASE64_URL"
	fieldCodeInvalidBase64RawURL    = "ERR_INVALID_BASE64_RAW_URL"
	fieldCodeInvalidDataURI         = "ERR_INVALID_DATA_URI"
	fieldCodeInvalidJSON            = "ERR_INVALID_JSON"
	fieldCodeInvalidJWT             = "ERR_INVALID_JWT"
	fieldCodeInvalidHTML            = "ERR_INVALID_HTML"
	fieldCodeInvalidHTMLEncoded     = "ERR_INVALID_HTML_ENCODED"
	fieldCodeInvalidURLEncoded      = "ERR_INVALID_URL_ENCODED"
	fieldCodeInvalidFile            = "ERR_INVALID_FILE"
	fieldCodeInvalidFilePath        = "ERR_INVALID_FILE_PATH"
	fieldCodeInvalidDir             = "ERR_INVALID_DIR"
	fieldCodeInvalidDirPath         = "ERR_INVALID_DIR_PATH"
	fieldCodeInvalidImage           = "ERR_INVALID_IMAGE"
	fieldCodeInvalidMimeType        = "ERR_INVALID_MIME_TYPE"
	fieldCodeInvalidISBN            = "ERR_INVALID_ISBN"
	fieldCodeInvalidISSN            = "ERR_INVALID_ISSN"
	fieldCodeInvalidCreditCard      = "ERR_INVALID_CREDIT_CARD"
	fieldCodeInvalidLuhn            = "ERR_INVALID_LUHN"
	fieldCodeInvalidCVE             = "ERR_INVALID_CVE"
	fieldCodeInvalidSemver          = "ERR_INVALID_SEMVER"
	fieldCodeInvalidHostname        = "ERR_INVALID_HOSTNAME"
	fieldCodeInvalidFQDN            = "ERR_INVALID_FQDN"
	fieldCodeInvalidHostnamePort    = "ERR_INVALID_HOSTNAME_PORT"
	fieldCodeInvalidPort            = "ERR_INVALID_PORT"
	fieldCodeInvalidDNSLabel        = "ERR_INVALID_DNS_LABEL"
	fieldCodeInvalidLatitude        = "ERR_INVALID_LATITUDE"
	fieldCodeInvalidLongitude       = "ERR_INVALID_LONGITUDE"
	fieldCodeInvalidSSN             = "ERR_INVALID_SSN"
	fieldCodeInvalidEthAddr         = "ERR_INVALID_ETH_ADDR"
	fieldCodeInvalidBtcAddr         = "ERR_INVALID_BTC_ADDR"
	fieldCodeInvalidMD4             = "ERR_INVALID_MD4"
	fieldCodeInvalidMD5             = "ERR_INVALID_MD5"
	fieldCodeInvalidSHA256          = "ERR_INVALID_SHA256"
	fieldCodeInvalidSHA384          = "ERR_INVALID_SHA384"
	fieldCodeInvalidSHA512          = "ERR_INVALID_SHA512"
	fieldCodeInvalidRIPEMD128       = "ERR_INVALID_RIPEMD128"
	fieldCodeInvalidRIPEMD160       = "ERR_INVALID_RIPEMD160"
	fieldCodeInvalidTiger128        = "ERR_INVALID_TIGER128"
	fieldCodeInvalidTiger160        = "ERR_INVALID_TIGER160"
	fieldCodeInvalidTiger192        = "ERR_INVALID_TIGER192"
	fieldCodeInvalidISO3166Alpha2   = "ERR_INVALID_ISO3166_ALPHA2"
	fieldCodeInvalidISO3166Alpha3   = "ERR_INVALID_ISO3166_ALPHA3"
	fieldCodeInvalidISO3166AlphaNum = "ERR_INVALID_ISO3166_ALPHA_NUMERIC"
	fieldCodeInvalidISO3166Subdiv   = "ERR_INVALID_ISO3166_2"
	fieldCodeInvalidCountryCode     = "ERR_INVALID_COUNTRY_CODE"
	fieldCodeInvalidISO4217         = "ERR_INVALID_ISO4217"
	fieldCodeInvalidLanguageTag     = "ERR_INVALID_LANGUAGE_TAG"
	fieldCodeInvalidPostalCode      = "ERR_INVALID_POSTAL_CODE"
	fieldCodeInvalidBIC             = "ERR_INVALID_BIC"
	fieldCodeInvalidMongoDBID       = "ERR_INVALID_MONGODB_ID"
	fieldCodeInvalidMongoDBConn     = "ERR_INVALID_MONGODB_CONNECTION"
	fieldCodeInvalidCron            = "ERR_INVALID_CRON"
	fieldCodeInvalidSpiceDB         = "ERR_INVALID_SPICEDB"
	fieldCodeInvalidEIN             = "ERR_INVALID_EIN"

	// Fallback for custom tags registered by the application.
	fieldCodeInvalid = "ERR_INVALID"
)

// FieldError describes a single failed field validation.
type FieldError struct {
	Field  string         `json:"field"`
	Code   string         `json:"code"`
	Params map[string]any `json:"params,omitempty"`
}

// validationBody is the JSON payload for validation error responses.
type validationBody struct {
	Success bool         `json:"success"`
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Errors  []FieldError `json:"errors,omitempty"`
}

// Validation writes a 422 Unprocessable Entity response for a failed request
// binding.
//
// req MUST be a non-nil pointer to the request struct; Validation panics
// otherwise. Field names are resolved from struct tags in this priority order:
//
//	json, form, uri, query, header, struct field name
//
// When err is not a validator error (for example malformed JSON), the response
// is returned without an errors list. Validation panics when called with a nil
// error.
func Validation(c *gin.Context, req any, err error) {
	if err == nil {
		panic("response: Validation called with a nil error")
	}
	if reflect.TypeOf(req) == nil {
		panic("response: Validation requires a non-nil pointer to the request struct")
	}
	rv := reflect.ValueOf(req)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		panic("response: Validation requires a non-nil pointer to the request struct")
	}

	body := validationBody{
		Success: false,
		Code:    CodeValidationError,
		Message: messageValidationFailed,
	}

	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		body.Errors = make([]FieldError, 0, len(verrs))
		for _, fe := range verrs {
			body.Errors = append(body.Errors, mapFieldError(fe, req))
		}
	}

	c.JSON(http.StatusUnprocessableEntity, body)
}

// mapFieldError converts a validator error into a FieldError, resolving the
// wire field name from the request struct.
func mapFieldError(fe validator.FieldError, req any) FieldError {
	code, params := mapTag(fe)
	return FieldError{
		Field:  resolveFieldName(req, fe.StructNamespace()),
		Code:   code,
		Params: params,
	}
}

// mapTag converts a validator tag into a standardized validation code and
// optional params.
func mapTag(fe validator.FieldError) (string, map[string]any) {
	tag := fe.Tag()
	param := fe.Param()

	switch tag {
	// Presence: the field is (or is not) allowed to have a value.
	case "required", "required_if", "required_unless", "skip_unless",
		"required_with", "required_with_all", "required_without", "required_without_all",
		"excluded_if", "excluded_unless", "excluded_with", "excluded_with_all",
		"excluded_without", "excluded_without_all":
		return fieldCodeRequired, nil

	// Length and value constraints.
	case "min":
		if isLengthTag(fe) {
			return fieldCodeMinLength, map[string]any{tag: asParam(param)}
		}
		return fieldCodeMinValue, map[string]any{tag: asParam(param)}
	case "max":
		if isLengthTag(fe) {
			return fieldCodeMaxLength, map[string]any{tag: asParam(param)}
		}
		return fieldCodeMaxValue, map[string]any{tag: asParam(param)}
	case "len":
		return fieldCodeLength, map[string]any{tag: asParam(param)}
	case "gt", "gte", "lt", "lte":
		return fieldCodeOutOfRange, map[string]any{tag: asParam(param)}
	case "eq", "eq_ignore_case", "ne", "ne_ignore_case":
		return fieldCodeInvalidValue, map[string]any{tag: asParam(param)}
	case "isdefault":
		return fieldCodeInvalidValue, nil

	// Cross-field constraints.
	case "eqfield", "nefield", "gtfield", "gtefield", "ltfield", "ltefield",
		"eqcsfield", "necsfield", "gtcsfield", "gtecsfield", "ltcsfield", "ltecsfield",
		"fieldcontains", "fieldexcludes":
		return fieldCodeMismatch, map[string]any{tag: param}

	// Enums and duplicates.
	case "oneof", "oneofci", "noneof", "noneofci":
		return fieldCodeInvalidEnum, map[string]any{tag: param}
	case "unique":
		return fieldCodeDuplicate, nil

	// String content.
	case "contains", "containsany", "containsrune", "excludes", "excludesall",
		"excludesrune", "startswith", "endswith", "startsnotwith", "endsnotwith":
		return fieldCodeInvalidContent, map[string]any{tag: param}

	// Formats.
	case "email":
		return fieldCodeInvalidEmail, nil
	case "url", "http_url", "https_url":
		return fieldCodeInvalidURL, nil
	case "uri":
		return fieldCodeInvalidURI, nil
	case "origin":
		return fieldCodeInvalidOrigin, nil
	case "urn_rfc2141":
		return fieldCodeInvalidURN, nil
	case "uuid", "uuid3", "uuid4", "uuid5", "uuid_rfc4122",
		"uuid3_rfc4122", "uuid4_rfc4122", "uuid5_rfc4122":
		return fieldCodeInvalidUUID, nil
	case "ulid":
		return fieldCodeInvalidULID, nil
	case "ip", "ipv4", "ipv6", "cidr", "cidrv4", "cidrv6",
		"ip_addr", "ip4_addr", "ip6_addr", "tcp_addr", "tcp4_addr", "tcp6_addr",
		"udp_addr", "udp4_addr", "udp6_addr", "unix_addr", "uds_exists":
		return fieldCodeInvalidIP, nil
	case "mac":
		return fieldCodeInvalidMAC, nil
	case "numeric", "number":
		return fieldCodeInvalidNumber, nil
	case "boolean":
		return fieldCodeInvalidBoolean, nil
	case "datetime":
		return fieldCodeInvalidDatetime, map[string]any{tag: param}
	case "timezone":
		return fieldCodeInvalidTimezone, nil
	case "e164":
		return fieldCodeInvalidE164, nil
	case "alpha", "alphaspace":
		return fieldCodeInvalidAlpha, nil
	case "alphanum", "alphanumspace":
		return fieldCodeInvalidAlphanum, nil
	case "alphaunicode":
		return fieldCodeInvalidAlphaUnicode, nil
	case "alphanumunicode":
		return fieldCodeInvalidAlphanumUnicode, nil
	case "ascii":
		return fieldCodeInvalidASCII, nil
	case "printascii":
		return fieldCodeInvalidPrintableASCII, nil
	case "multibyte":
		return fieldCodeInvalidMultibyte, nil
	case "lowercase":
		return fieldCodeInvalidLowercase, nil
	case "uppercase":
		return fieldCodeInvalidUppercase, nil
	case "hexadecimal":
		return fieldCodeInvalidHexadecimal, nil
	case "hexcolor":
		return fieldCodeInvalidHexColor, nil
	case "rgb":
		return fieldCodeInvalidRGB, nil
	case "rgba":
		return fieldCodeInvalidRGBA, nil
	case "hsl":
		return fieldCodeInvalidHSL, nil
	case "hsla":
		return fieldCodeInvalidHSLA, nil
	case "cmyk":
		return fieldCodeInvalidCMYK, nil
	case "iscolor":
		return fieldCodeInvalidColor, nil
	case "base32":
		return fieldCodeInvalidBase32, nil
	case "base64":
		return fieldCodeInvalidBase64, nil
	case "base64url":
		return fieldCodeInvalidBase64URL, nil
	case "base64rawurl":
		return fieldCodeInvalidBase64RawURL, nil
	case "datauri":
		return fieldCodeInvalidDataURI, nil
	case "json":
		return fieldCodeInvalidJSON, nil
	case "jwt":
		return fieldCodeInvalidJWT, nil
	case "html":
		return fieldCodeInvalidHTML, nil
	case "html_encoded":
		return fieldCodeInvalidHTMLEncoded, nil
	case "url_encoded":
		return fieldCodeInvalidURLEncoded, nil
	case "file":
		return fieldCodeInvalidFile, nil
	case "filepath":
		return fieldCodeInvalidFilePath, nil
	case "dir":
		return fieldCodeInvalidDir, nil
	case "dirpath":
		return fieldCodeInvalidDirPath, nil
	case "image":
		return fieldCodeInvalidImage, nil
	case "mimetype":
		return fieldCodeInvalidMimeType, nil
	case "isbn", "isbn10", "isbn13":
		return fieldCodeInvalidISBN, nil
	case "issn":
		return fieldCodeInvalidISSN, nil
	case "credit_card":
		return fieldCodeInvalidCreditCard, nil
	case "luhn_checksum":
		return fieldCodeInvalidLuhn, nil
	case "cve":
		return fieldCodeInvalidCVE, nil
	case "semver":
		return fieldCodeInvalidSemver, nil
	case "hostname", "hostname_rfc1123":
		return fieldCodeInvalidHostname, nil
	case "fqdn":
		return fieldCodeInvalidFQDN, nil
	case "hostname_port":
		return fieldCodeInvalidHostnamePort, nil
	case "port":
		return fieldCodeInvalidPort, nil
	case "dns_rfc1035_label":
		return fieldCodeInvalidDNSLabel, nil
	case "latitude":
		return fieldCodeInvalidLatitude, nil
	case "longitude":
		return fieldCodeInvalidLongitude, nil
	case "ssn":
		return fieldCodeInvalidSSN, nil
	case "eth_addr", "eth_addr_checksum":
		return fieldCodeInvalidEthAddr, nil
	case "btc_addr", "btc_addr_bech32":
		return fieldCodeInvalidBtcAddr, nil
	case "md4":
		return fieldCodeInvalidMD4, nil
	case "md5":
		return fieldCodeInvalidMD5, nil
	case "sha256":
		return fieldCodeInvalidSHA256, nil
	case "sha384":
		return fieldCodeInvalidSHA384, nil
	case "sha512":
		return fieldCodeInvalidSHA512, nil
	case "ripemd128":
		return fieldCodeInvalidRIPEMD128, nil
	case "ripemd160":
		return fieldCodeInvalidRIPEMD160, nil
	case "tiger128":
		return fieldCodeInvalidTiger128, nil
	case "tiger160":
		return fieldCodeInvalidTiger160, nil
	case "tiger192":
		return fieldCodeInvalidTiger192, nil
	case "iso3166_1_alpha2", "iso3166_1_alpha2_eu":
		return fieldCodeInvalidISO3166Alpha2, nil
	case "iso3166_1_alpha3", "iso3166_1_alpha3_eu":
		return fieldCodeInvalidISO3166Alpha3, nil
	case "iso3166_1_alpha_numeric", "iso3166_1_alpha_numeric_eu":
		return fieldCodeInvalidISO3166AlphaNum, nil
	case "iso3166_2":
		return fieldCodeInvalidISO3166Subdiv, nil
	case "country_code", "eu_country_code":
		return fieldCodeInvalidCountryCode, nil
	case "iso4217", "iso4217_numeric":
		return fieldCodeInvalidISO4217, nil
	case "bcp47_language_tag", "bcp47_strict_language_tag":
		return fieldCodeInvalidLanguageTag, nil
	case "postcode_iso3166_alpha2", "postcode_iso3166_alpha2_field":
		return fieldCodeInvalidPostalCode, nil
	case "bic", "bic_iso_9362_2014":
		return fieldCodeInvalidBIC, nil
	case "mongodb":
		return fieldCodeInvalidMongoDBID, nil
	case "mongodb_connection_string":
		return fieldCodeInvalidMongoDBConn, nil
	case "cron":
		return fieldCodeInvalidCron, nil
	case "spicedb":
		return fieldCodeInvalidSpiceDB, nil
	case "ein":
		return fieldCodeInvalidEIN, nil

	// Custom tags registered by the application.
	default:
		return fieldCodeInvalid, nil
	}
}

// isLengthTag reports whether min/max applies to the length of a string or
// collection rather than the numeric value of a field.
func isLengthTag(fe validator.FieldError) bool {
	switch fe.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
		return true
	default:
		return false
	}
}

// asParam converts a validator parameter into a JSON-friendly value, keeping
// numbers numeric.
func asParam(v string) any {
	if i, err := strconv.ParseInt(v, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(v, 64); err == nil {
		return f
	}
	return v
}

// resolveFieldName returns the wire name of the field identified by ns, a
// validator struct namespace such as "CreateUserRequest.Address.City" or
// "CreateUserRequest.Items[0].Name". The first segment is the root struct
// name and is skipped. When the field cannot be resolved, the last path
// segment is returned as a fallback.
func resolveFieldName(req any, ns string) string {
	current := derefType(reflect.TypeOf(req))
	segs := strings.Split(ns, ".")
	if len(segs) > 1 {
		segs = segs[1:]
	}

	for i, seg := range segs {
		name := seg
		indexed := false
		if j := strings.IndexByte(name, '['); j >= 0 {
			name = name[:j]
			indexed = true
		}

		f, ok := current.FieldByName(name)
		if !ok {
			return segs[len(segs)-1]
		}
		if i == len(segs)-1 {
			return tagName(f)
		}

		next := derefType(f.Type)
		switch {
		case indexed:
			switch next.Kind() {
			case reflect.Slice, reflect.Array:
				current = derefType(next.Elem())
			default:
				return segs[len(segs)-1]
			}
		case next.Kind() == reflect.Struct:
			current = next
		case next.Kind() == reflect.Slice || next.Kind() == reflect.Array:
			current = derefType(next.Elem())
		default:
			return segs[len(segs)-1]
		}
	}
	return segs[len(segs)-1]
}

// derefType unwraps pointer types.
func derefType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

// tagName returns the wire name of a struct field following the documented tag
// priority, falling back to the Go field name.
func tagName(f reflect.StructField) string {
	for _, tag := range []string{"json", "form", "uri", "query", "header"} {
		v, ok := f.Tag.Lookup(tag)
		if !ok {
			continue
		}
		name := strings.Split(v, ",")[0]
		if name != "" && name != "-" {
			return name
		}
	}
	return f.Name
}
