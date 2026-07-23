package common

import (
	"bytes"
	"encoding/json"
)

// -------------------
// Source: https://medium.com/@0xfurai/the-patch-request-trap-why-your-go-api-probably-handles-updates-wrong-87a75dc16d05
// -------------------

type Optional[T any] struct {
	set   bool // Was this field present in the JSON?
	null  bool // If present, was it null?
	value T    // If present and not null, what's the value?
}

func (o Optional[T]) IsSet() bool {
	return o.set
}

func (o Optional[T]) IsNull() bool {
	return o.set && o.null
}

func (o Optional[T]) Value() (T, bool) {
	if o.set && !o.null {
		return o.value, true
	}
	var zero T
	return zero, false
}

func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.set = true // If this method is called, the field was in the JSON

	// Check for explicit null
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		o.null = true
		return nil
	}

	// Otherwise unmarshal the actual value
	o.null = false
	return json.Unmarshal(data, &o.value)
}

func SetUpdate[T any](updates map[string]interface{}, column string, opt Optional[T]) {
	if !opt.IsSet() {
		return
	}

	if opt.IsNull() {
		updates[column] = nil
		return
	}

	if value, ok := opt.Value(); ok {
		updates[column] = value
	}
}
