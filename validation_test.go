package response

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-playground/validator/v10"
)

type createUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type minLengthRequest struct {
	Password string `json:"password" validate:"min=8"`
}

type minValueRequest struct {
	Age int `json:"age" validate:"min=18"`
}

type enumRequest struct {
	Role string `json:"role" validate:"oneof=admin user"`
}

type address struct {
	City string `json:"city" validate:"required"`
}

type nestedRequest struct {
	Address address `json:"address"`
}

type item struct {
	SKU string `json:"sku" validate:"required"`
}

type diveRequest struct {
	Items []item `json:"items" validate:"dive"`
}

type priorityRequest struct {
	JSONName string `json:"json_name" form:"form_name" query:"query_name" validate:"required"`
	FormName string `form:"form_name2" query:"query_name2" validate:"required"`
	Plain    string `validate:"required"`
}

// validate runs the default validator against req.
func validate(t *testing.T, req any) error {
	t.Helper()
	return validator.New().Struct(req)
}

func TestValidation(t *testing.T) {
	req := &createUserRequest{}
	err := validate(t, req)
	if err == nil {
		t.Fatal("expected validation errors")
	}

	c, w := newContext()
	Validation(c, req, err)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}
	body := decode(t, w)
	if body["success"] != false {
		t.Errorf("success = %v, want false", body["success"])
	}
	if body["code"] != CodeValidationError {
		t.Errorf("code = %v, want %s", body["code"], CodeValidationError)
	}
	if body["message"] != "Validation failed" {
		t.Errorf("message = %v, want %q", body["message"], "Validation failed")
	}

	list, ok := body["errors"].([]any)
	if !ok {
		t.Fatalf("errors = %v, want a list", body["errors"])
	}
	if len(list) != 2 {
		t.Fatalf("got %d errors, want 2", len(list))
	}

	byField := map[string]map[string]any{}
	for _, e := range list {
		m := e.(map[string]any)
		byField[m["field"].(string)] = m
	}

	if byField["email"]["code"] != fieldCodeRequired {
		t.Errorf("email code = %v, want %s", byField["email"]["code"], fieldCodeRequired)
	}
	if byField["password"]["code"] != fieldCodeRequired {
		t.Errorf("password code = %v, want %s", byField["password"]["code"], fieldCodeRequired)
	}
	if _, ok := byField["email"]["params"]; ok {
		t.Errorf("email must not include params: %v", byField["email"])
	}
	if _, ok := byField["Email"]; ok {
		t.Error("field name must use the json tag, not the Go field name")
	}
}

func TestValidationMinLength(t *testing.T) {
	req := &minLengthRequest{Password: "short"}
	err := validate(t, req)

	c, w := newContext()
	Validation(c, req, err)

	body := decode(t, w)
	list := body["errors"].([]any)
	entry := list[0].(map[string]any)
	if entry["field"] != "password" {
		t.Errorf("field = %v, want password", entry["field"])
	}
	if entry["code"] != fieldCodeMinLength {
		t.Errorf("code = %v, want %s", entry["code"], fieldCodeMinLength)
	}
	params := entry["params"].(map[string]any)
	if params["min"] != float64(8) {
		t.Errorf("params.min = %v, want 8", params["min"])
	}
}

func TestValidationMinValue(t *testing.T) {
	req := &minValueRequest{Age: 10}
	err := validate(t, req)

	c, w := newContext()
	Validation(c, req, err)

	body := decode(t, w)
	entry := body["errors"].([]any)[0].(map[string]any)
	if entry["field"] != "age" {
		t.Errorf("field = %v, want age", entry["field"])
	}
	if entry["code"] != fieldCodeMinValue {
		t.Errorf("code = %v, want %s", entry["code"], fieldCodeMinValue)
	}
}

func TestValidationInvalidEnum(t *testing.T) {
	req := &enumRequest{Role: "root"}
	err := validate(t, req)

	c, w := newContext()
	Validation(c, req, err)

	body := decode(t, w)
	entry := body["errors"].([]any)[0].(map[string]any)
	if entry["code"] != fieldCodeInvalidEnum {
		t.Errorf("code = %v, want %s", entry["code"], fieldCodeInvalidEnum)
	}
	params := entry["params"].(map[string]any)
	if params["oneof"] != "admin user" {
		t.Errorf("params.oneof = %v, want %q", params["oneof"], "admin user")
	}
}

func TestValidationNestedField(t *testing.T) {
	req := &nestedRequest{}
	err := validate(t, req)

	c, w := newContext()
	Validation(c, req, err)

	body := decode(t, w)
	entry := body["errors"].([]any)[0].(map[string]any)
	if entry["field"] != "city" {
		t.Errorf("field = %v, want city", entry["field"])
	}
	if entry["code"] != fieldCodeRequired {
		t.Errorf("code = %v, want %s", entry["code"], fieldCodeRequired)
	}
}

func TestValidationDiveField(t *testing.T) {
	req := &diveRequest{Items: []item{{}}}
	err := validate(t, req)

	c, w := newContext()
	Validation(c, req, err)

	body := decode(t, w)
	entry := body["errors"].([]any)[0].(map[string]any)
	if entry["field"] != "sku" {
		t.Errorf("field = %v, want sku", entry["field"])
	}
}

func TestValidationTagPriority(t *testing.T) {
	req := &priorityRequest{}
	err := validate(t, req)

	c, w := newContext()
	Validation(c, req, err)

	body := decode(t, w)

	// The three required failures must resolve to json_name, form_name2 and
	// Plain (json > form > field name).
	seen := map[string]bool{}
	for _, e := range body["errors"].([]any) {
		m := e.(map[string]any)
		seen[m["field"].(string)] = true
	}
	if !seen["json_name"] {
		t.Error("expected field json_name from json tag")
	}
	if !seen["form_name2"] {
		t.Error("expected field form_name2 from form tag")
	}
	if !seen["Plain"] {
		t.Error("expected field Plain from struct field name fallback")
	}
}

func TestValidationWrappedError(t *testing.T) {
	req := &createUserRequest{}
	err := fmt.Errorf("binding failed: %w", validate(t, req))

	c, w := newContext()
	Validation(c, req, err)

	body := decode(t, w)
	if _, ok := body["errors"]; !ok {
		t.Fatal("wrapped validation errors must be detected")
	}
}

func TestValidationNonValidationError(t *testing.T) {
	c, w := newContext()
	Validation(c, &createUserRequest{}, errors.New("invalid character 'x' looking for beginning of value"))

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}
	body := decode(t, w)
	if body["code"] != CodeValidationError {
		t.Errorf("code = %v, want %s", body["code"], CodeValidationError)
	}
	if _, ok := body["errors"]; ok {
		t.Error("errors must be omitted when there are no field errors")
	}
}

func TestValidationPanicsOnNonPointer(t *testing.T) {
	req := createUserRequest{}
	err := validate(t, &req)

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on non-pointer request")
		}
	}()
	c, _ := newContext()
	Validation(c, req, err)
}

func TestValidationPanicsOnNilPointer(t *testing.T) {
	req := (*createUserRequest)(nil)
	err := validate(t, &createUserRequest{})

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on nil pointer request")
		}
	}()
	c, _ := newContext()
	Validation(c, req, err)
}

func TestValidationPanicsOnNilError(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on nil error")
		}
	}()
	c, _ := newContext()
	Validation(c, &createUserRequest{}, nil)
}

// TestMapTag verifies that each validator tag family maps to the documented
// validation code.
func TestMapTag(t *testing.T) {
	cases := []struct {
		name  string
		value any
		tag   string
		code  string
	}{
		// Presence and constraints.
		{"required", "", "required", fieldCodeRequired},
		{"len", "12345", "len=6", fieldCodeLength},
		{"min-length", "abc", "min=5", fieldCodeMinLength},
		{"max-length", "abcdef", "max=3", fieldCodeMaxLength},
		{"min-value", 5, "min=18", fieldCodeMinValue},
		{"max-value", 20, "max=10", fieldCodeMaxValue},
		{"out-of-range", 1, "gt=5", fieldCodeOutOfRange},
		{"eq", "x", "eq=y", fieldCodeInvalidValue},
		{"ne", "y", "ne=y", fieldCodeInvalidValue},
		{"oneof", "c", "oneof=a b", fieldCodeInvalidEnum},
		{"unique", []string{"a", "a"}, "unique", fieldCodeDuplicate},
		{"contains", "abc", "contains=x", fieldCodeInvalidContent},

		// Formats.
		{"email", "not-an-email", "email", fieldCodeInvalidEmail},
		{"url", "not a url", "url", fieldCodeInvalidURL},
		{"origin", "not-an-origin", "origin", fieldCodeInvalidOrigin},
		{"urn", "not-a-urn", "urn_rfc2141", fieldCodeInvalidURN},
		{"uuid", "not-a-uuid", "uuid4", fieldCodeInvalidUUID},
		{"ulid", "not-ulid", "ulid", fieldCodeInvalidULID},
		{"ip", "999.999.999.999", "ip", fieldCodeInvalidIP},
		{"mac", "not-a-mac", "mac", fieldCodeInvalidMAC},
		{"numeric", "abc", "numeric", fieldCodeInvalidNumber},
		{"boolean", "not-bool", "boolean", fieldCodeInvalidBoolean},
		{"datetime", "2024-13-01", "datetime=2006-01-02", fieldCodeInvalidDatetime},
		{"timezone", "Not/AZone", "timezone", fieldCodeInvalidTimezone},
		{"e164", "+123", "e164", fieldCodeInvalidE164},
		{"alpha", "abc123", "alpha", fieldCodeInvalidAlpha},
		{"alphanum", "abc!", "alphanum", fieldCodeInvalidAlphanum},
		{"alpha-unicode", "abc123", "alphaunicode", fieldCodeInvalidAlphaUnicode},
		{"alphanum-unicode", "abc!", "alphanumunicode", fieldCodeInvalidAlphanumUnicode},
		{"ascii", "héllo", "ascii", fieldCodeInvalidASCII},
		{"printable-ascii", "abc\n", "printascii", fieldCodeInvalidPrintableASCII},
		{"multibyte", "abc", "multibyte", fieldCodeInvalidMultibyte},
		{"lowercase", "ABC", "lowercase", fieldCodeInvalidLowercase},
		{"uppercase", "abc", "uppercase", fieldCodeInvalidUppercase},
		{"hexadecimal", "xyz", "hexadecimal", fieldCodeInvalidHexadecimal},
		{"hex-color", "red", "hexcolor", fieldCodeInvalidHexColor},
		{"rgb", "red", "rgb", fieldCodeInvalidRGB},
		{"cmyk", "red", "cmyk", fieldCodeInvalidCMYK},
		{"iscolor", "red", "iscolor", fieldCodeInvalidColor},
		{"base32", "not base32!", "base32", fieldCodeInvalidBase32},
		{"base64", "not base64!!", "base64", fieldCodeInvalidBase64},
		{"base64url", "!!!", "base64url", fieldCodeInvalidBase64URL},
		{"base64rawurl", "a b", "base64rawurl", fieldCodeInvalidBase64RawURL},
		{"data-uri", "not", "datauri", fieldCodeInvalidDataURI},
		{"json", "not json", "json", fieldCodeInvalidJSON},
		{"jwt", "not a jwt", "jwt", fieldCodeInvalidJWT},
		{"html", "abc", "html", fieldCodeInvalidHTML},
		{"html-encoded", "abc", "html_encoded", fieldCodeInvalidHTMLEncoded},
		{"url-encoded", "abc%xyz", "url_encoded", fieldCodeInvalidURLEncoded},
		{"file", "nonexistent_file_xyz", "file", fieldCodeInvalidFile},
		{"dir", "not-a-dir", "dir", fieldCodeInvalidDir},
		{"image", "not", "image", fieldCodeInvalidImage},
		{"mimetype", "not-a-mime", "mimetype", fieldCodeInvalidMimeType},
		{"isbn", "123", "isbn", fieldCodeInvalidISBN},
		{"issn", "123", "issn", fieldCodeInvalidISSN},
		{"credit-card", "1234", "credit_card", fieldCodeInvalidCreditCard},
		{"luhn", "1234", "luhn_checksum", fieldCodeInvalidLuhn},
		{"cve", "not-cve", "cve", fieldCodeInvalidCVE},
		{"semver", "not-semver", "semver", fieldCodeInvalidSemver},
		{"hostname", "-invalid-", "hostname", fieldCodeInvalidHostname},
		{"fqdn", "not fqdn", "fqdn", fieldCodeInvalidFQDN},
		{"hostname-port", "example.com:99999", "hostname_port", fieldCodeInvalidHostnamePort},
		{"port", uint(99999), "port", fieldCodeInvalidPort},
		{"dns-label", "invalid label", "dns_rfc1035_label", fieldCodeInvalidDNSLabel},
		{"latitude", "99", "latitude", fieldCodeInvalidLatitude},
		{"longitude", "999", "longitude", fieldCodeInvalidLongitude},
		{"ssn", "123", "ssn", fieldCodeInvalidSSN},
		{"eth-addr", "0x123", "eth_addr", fieldCodeInvalidEthAddr},
		{"btc-addr", "1", "btc_addr", fieldCodeInvalidBtcAddr},
		{"md4", "short", "md4", fieldCodeInvalidMD4},
		{"md5", "short", "md5", fieldCodeInvalidMD5},
		{"sha256", "short", "sha256", fieldCodeInvalidSHA256},
		{"sha384", "short", "sha384", fieldCodeInvalidSHA384},
		{"sha512", "short", "sha512", fieldCodeInvalidSHA512},
		{"ripemd128", "short", "ripemd128", fieldCodeInvalidRIPEMD128},
		{"ripemd160", "short", "ripemd160", fieldCodeInvalidRIPEMD160},
		{"tiger128", "short", "tiger128", fieldCodeInvalidTiger128},
		{"tiger160", "short", "tiger160", fieldCodeInvalidTiger160},
		{"tiger192", "short", "tiger192", fieldCodeInvalidTiger192},
		{"iso3166-alpha2", "XX", "iso3166_1_alpha2", fieldCodeInvalidISO3166Alpha2},
		{"iso3166-alpha3", "ZZZ", "iso3166_1_alpha3", fieldCodeInvalidISO3166Alpha3},
		{"iso3166-alpha-numeric", "999", "iso3166_1_alpha_numeric", fieldCodeInvalidISO3166AlphaNum},
		{"country-code", "XX", "country_code", fieldCodeInvalidCountryCode},
		{"iso4217", "XYZ", "iso4217", fieldCodeInvalidISO4217},
		{"language-tag", "12345", "bcp47_language_tag", fieldCodeInvalidLanguageTag},
		{"postal-code", "not-a-zip", "postcode_iso3166_alpha2=US", fieldCodeInvalidPostalCode},
		{"bic", "short", "bic", fieldCodeInvalidBIC},
		{"mongodb", "invalid", "mongodb", fieldCodeInvalidMongoDBID},
		{"mongodb-connection", "not-mongodb", "mongodb_connection_string", fieldCodeInvalidMongoDBConn},
		{"cron", "not cron", "cron", fieldCodeInvalidCron},
		{"spicedb", "", "spicedb", fieldCodeInvalidSpiceDB},
		{"ein", "123", "ein", fieldCodeInvalidEIN},

		// Custom tags registered by the application.
		{"custom", "x", "custom", fieldCodeInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := validator.New()
			if tc.code == fieldCodeInvalid {
				if err := v.RegisterValidation("custom",
					func(fl validator.FieldLevel) bool { return false }); err != nil {
					t.Fatal(err)
				}
			}
			err := v.Var(tc.value, tc.tag)
			if err == nil {
				t.Fatalf("expected tag %q to fail", tc.tag)
			}
			var verrs validator.ValidationErrors
			if !errors.As(err, &verrs) {
				t.Fatalf("unexpected error type %T", err)
			}
			code, _ := mapTag(verrs[0])
			if code != tc.code {
				t.Errorf("tag %q: code = %s, want %s", tc.tag, code, tc.code)
			}
		})
	}
}

// TestMapTagParams verifies that comparison and cross-field tags carry params.
func TestMapTagParams(t *testing.T) {
	t.Run("len", func(t *testing.T) {
		var verrs validator.ValidationErrors
		if err := validator.New().Var("12345", "len=6"); !errors.As(err, &verrs) {
			t.Fatal("expected validation error")
		}
		_, params := mapTag(verrs[0])
		if params["len"] != int64(6) {
			t.Errorf("params.len = %v, want 6", params["len"])
		}
	})

	t.Run("eqfield", func(t *testing.T) {
		type req struct {
			A string `validate:"eqfield=B"`
			B string
		}
		var verrs validator.ValidationErrors
		if err := validator.New().Struct(&req{A: "x", B: "y"}); !errors.As(err, &verrs) {
			t.Fatal("expected validation error")
		}
		code, params := mapTag(verrs[0])
		if code != fieldCodeMismatch {
			t.Errorf("code = %s, want %s", code, fieldCodeMismatch)
		}
		if params["eqfield"] != "B" {
			t.Errorf("params.eqfield = %v, want B", params["eqfield"])
		}
	})
}
