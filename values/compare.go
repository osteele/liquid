package values

import (
	"reflect"
	"time"

	"github.com/autopilot3/ap3-types-go/types/date"
)

var (
	int64Type   = reflect.TypeOf(int64(0))
	float64Type = reflect.TypeOf(float64(0))
)

// Equal returns a bool indicating whether a == b after conversion.
func Equal(a, b interface{}) bool { // nolint: gocyclo
	a, b = ToLiquid(a), ToLiquid(b)
	if a == nil || b == nil {
		return a == b
	}
	ra, rb := reflect.ValueOf(a), reflect.ValueOf(b)
	if ra.Kind() == reflect.Struct && ra.Type() == reflect.TypeOf(time.Time{}) {
		// we have a time comparison, try to convert b to time.Time
		// there should be only two cases: b is a user input string or a time.Time which is our variabeles from crm
		if rb.Kind() == reflect.String {
			db, err := ParseDate(rb.String())
			if err == nil {
				return ra.Interface().(time.Time).Equal(db)
			} else {
				return false
			}
		} else if rb.Kind() == reflect.Struct && rb.Type() == reflect.TypeOf(time.Time{}) {
			return ra.Interface().(time.Time).Equal(rb.Interface().(time.Time))
		}
	}
	switch joinKind(ra.Kind(), rb.Kind()) {
	case reflect.Array, reflect.Slice:
		if ra.Len() != rb.Len() {
			return false
		}
		for i := 0; i < ra.Len(); i++ {
			if !Equal(ra.Index(i).Interface(), rb.Index(i).Interface()) {
				return false
			}
		}
		return true
	case reflect.Bool:
		return ra.Bool() == rb.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return ra.Convert(int64Type).Int() == rb.Convert(int64Type).Int()
	case reflect.Float32, reflect.Float64:
		return ra.Convert(float64Type).Float() == rb.Convert(float64Type).Float()
	case reflect.String:
		return ra.String() == rb.String()
	case reflect.Ptr:
		if rb.Kind() == reflect.Ptr && (ra.IsNil() || rb.IsNil()) {
			return ra.IsNil() == rb.IsNil()
		}
		return a == b
	default:
		return a == b
	}
}

// Less returns a bool indicating whether a < b.
func Less(a, b interface{}) bool {
	a, b = ToLiquid(a), ToLiquid(b)
	if a == nil || b == nil {
		return false
	}
	ra, rb := reflect.ValueOf(a), reflect.ValueOf(b)
	// time comparison
	if ra.Kind() == reflect.Struct {
		if ra.Type() == reflect.TypeOf(time.Time{}) {
			// we have a time comparison, try to convert b to time.time
			// there should be only two cases: b is a user input string or a time.Time which is our variabeles from crm
			if rb.Kind() == reflect.String {
				db, err := ParseDate(rb.String())
				if err == nil {
					return ra.Interface().(time.Time).Before(db)
				}
			} else if rb.Kind() == reflect.Struct && rb.Type() == reflect.TypeOf(time.Time{}) {
				return ra.Interface().(time.Time).Before(rb.Interface().(time.Time))
			}
		}
	} else if rb.Kind() == reflect.Struct {
		if rb.Type() == reflect.TypeOf(time.Time{}) {
			// we have a time comparison, try to convert a to time.time
			// there should be only two cases: a is a user input string or a time.Time which is our variabeles from crm
			if ra.Kind() == reflect.String {
				da, err := ParseDate(ra.String())
				if err == nil {
					return da.Before(rb.Interface().(time.Time))
				}
			} else if ra.Kind() == reflect.Struct && ra.Type() == reflect.TypeOf(time.Time{}) {
				return ra.Interface().(time.Time).Before(rb.Interface().(time.Time))
			}
		}
	}
	// date comparison only for date.Date vs string case, since date.Date is of kind int so naturally two date.Date can be compared
	dVar := date.Date(1)
	if reflect.TypeOf(a) == reflect.TypeOf(dVar) {
		if rb.Kind() == reflect.String {
			db, err := ParseDate(rb.String())
			if err == nil {
				d := date.NewFromUTCTime(db)
				return a.(date.Date) < d
			}
		}
	} else if reflect.TypeOf(b) == reflect.TypeOf(dVar) {
		if ra.Kind() == reflect.String {
			da, err := ParseDate(ra.String())
			if err == nil {
				d := date.NewFromUTCTime(da)
				return d < b.(date.Date)
			}
		}
	}

	switch joinKind(ra.Kind(), rb.Kind()) {
	case reflect.Bool:
		return !ra.Bool() && rb.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return ra.Convert(int64Type).Int() < rb.Convert(int64Type).Int()
	case reflect.Float32, reflect.Float64:
		return ra.Convert(float64Type).Float() < rb.Convert(float64Type).Float()
	case reflect.String:
		return ra.String() < rb.String()
	default:
		return false
	}
}

func joinKind(a, b reflect.Kind) reflect.Kind { // nolint: gocyclo
	if a == b {
		return a
	}
	switch a {
	case reflect.Array, reflect.Slice:
		if b == reflect.Array || b == reflect.Slice {
			return reflect.Slice
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if isIntKind(b) {
			return reflect.Int64
		}
		if isFloatKind(b) {
			return reflect.Float64
		}
	case reflect.Float32, reflect.Float64:
		if isIntKind(b) || isFloatKind(b) {
			return reflect.Float64
		}
	}
	return reflect.Invalid
}

func isIntKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	default:
		return false
	}
}

func isFloatKind(k reflect.Kind) bool {
	switch k {
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}
