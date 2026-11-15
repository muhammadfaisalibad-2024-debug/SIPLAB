package helper

import (
	"fmt"
	"net/mail"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)

// ValidateStruct memvalidasi struct berdasarkan tag `validate` dan mengembalikan map kesalahan per field JSON.
// Mengembalikan nil apabila seluruh aturan validasi terpenuhi.
func ValidateStruct(s any) map[string]string {
	if s == nil {
		return nil
	}

	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	fields := make(map[string]string)

	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		fv := v.Field(i)

		tag := sf.Tag.Get("validate")
		if tag == "" || tag == "-" {
			continue
		}

		fieldName := jsonFieldName(sf)
		rules := strings.Split(tag, ",")

		// Tangani pointer & omitnil
		if fv.Kind() == reflect.Ptr {
			if fv.IsNil() {
				if containsRule(rules, "omitnil") || !containsRule(rules, "required") {
					continue
				}
				fields[fieldName] = "wajib diisi"
				continue
			}
			fv = fv.Elem()
		}

		for _, rule := range rules {
			rule = strings.TrimSpace(rule)
			if rule == "" || rule == "omitnil" {
				continue
			}

			if msg := checkRule(fv, rule); msg != "" {
				fields[fieldName] = msg
				break
			}
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return fields
}

func jsonFieldName(sf reflect.StructField) string {
	jsonTag := sf.Tag.Get("json")
	if jsonTag != "" {
		name := strings.Split(jsonTag, ",")[0]
		if name != "" && name != "-" {
			return name
		}
	}
	return strings.ToLower(sf.Name)
}

func containsRule(rules []string, target string) bool {
	for _, r := range rules {
		if strings.TrimSpace(r) == target {
			return true
		}
	}
	return false
}

func checkRule(fv reflect.Value, rule string) string {
	name, param, hasParam := strings.Cut(rule, "=")

	switch name {
	case "required":
		if isZeroValue(fv) {
			return "wajib diisi"
		}
	case "min":
		if !hasParam {
			return ""
		}
		minVal, err := strconv.Atoi(param)
		if err != nil {
			return ""
		}
		switch fv.Kind() {
		case reflect.String:
			if len([]rune(fv.String())) < minVal {
				return fmt.Sprintf("minimal %d karakter", minVal)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if fv.Int() < int64(minVal) {
				return fmt.Sprintf("minimal bernilai %d", minVal)
			}
		case reflect.Slice, reflect.Array:
			if fv.Len() < minVal {
				return fmt.Sprintf("minimal %d item", minVal)
			}
		}
	case "max":
		if !hasParam {
			return ""
		}
		maxVal, err := strconv.Atoi(param)
		if err != nil {
			return ""
		}
		switch fv.Kind() {
		case reflect.String:
			if len(fv.String()) > maxVal {
				return fmt.Sprintf("maksimal %d karakter", maxVal)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if fv.Int() > int64(maxVal) {
				return fmt.Sprintf("maksimal bernilai %d", maxVal)
			}
		case reflect.Slice, reflect.Array:
			if fv.Len() > maxVal {
				return fmt.Sprintf("maksimal %d item", maxVal)
			}
		}
	case "email":
		if fv.Kind() == reflect.String {
			str := fv.String()
			addr, err := mail.ParseAddress(str)
			if err != nil || addr.Address != str || !strings.Contains(strings.Split(str, "@")[1], ".") {
				return "format email tidak valid"
			}
		}
	case "username":
		if fv.Kind() == reflect.String {
			if !usernameRegex.MatchString(fv.String()) {
				return "hanya boleh berisi huruf, angka, titik, garis bawah, atau tanda hubung"
			}
		}
	case "nospace":
		if fv.Kind() == reflect.String {
			for _, r := range fv.String() {
				if unicode.IsSpace(r) {
					return "tidak boleh mengandung spasi"
				}
			}
		}
	case "strongpassword":
		if fv.Kind() == reflect.String {
			s := fv.String()
			if len(s) < 8 {
				return "password minimal 8 karakter"
			}
			var hasLetter, hasDigit bool
			for _, r := range s {
				if unicode.IsLetter(r) {
					hasLetter = true
				}
				if unicode.IsDigit(r) {
					hasDigit = true
				}
			}
			if !hasLetter || !hasDigit {
				return "password harus mengandung kombinasi huruf dan angka"
			}
		}
	case "oneof":
		if hasParam && fv.Kind() == reflect.String {
			options := strings.Fields(param)
			val := fv.String()
			matched := false
			for _, opt := range options {
				if val == opt {
					matched = true
					break
				}
			}
			if !matched {
				return fmt.Sprintf("harus salah satu dari: %s", strings.Join(options, ", "))
			}
		}
	}
	return ""
}

func isZeroValue(fv reflect.Value) bool {
	switch fv.Kind() {
	case reflect.String:
		return strings.TrimSpace(fv.String()) == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return fv.Float() == 0
	case reflect.Slice, reflect.Map, reflect.Array:
		return fv.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return fv.IsNil()
	default:
		return fv.IsZero()
	}
}