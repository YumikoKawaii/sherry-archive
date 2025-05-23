// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: tracking/tracking.proto

package api

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
	_ = sort.Sort
)

// Validate checks the field values on LogEntryRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *LogEntryRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on LogEntryRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// LogEntryRequestMultiError, or nil if none found.
func (m *LogEntryRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *LogEntryRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if all {
		switch v := interface{}(m.GetLogEntry()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, LogEntryRequestValidationError{
					field:  "LogEntry",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, LogEntryRequestValidationError{
					field:  "LogEntry",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetLogEntry()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return LogEntryRequestValidationError{
				field:  "LogEntry",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(errors) > 0 {
		return LogEntryRequestMultiError(errors)
	}

	return nil
}

// LogEntryRequestMultiError is an error wrapping multiple validation errors
// returned by LogEntryRequest.ValidateAll() if the designated constraints
// aren't met.
type LogEntryRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m LogEntryRequestMultiError) Error() string {
	msgs := make([]string, 0, len(m))
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m LogEntryRequestMultiError) AllErrors() []error { return m }

// LogEntryRequestValidationError is the validation error returned by
// LogEntryRequest.Validate if the designated constraints aren't met.
type LogEntryRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e LogEntryRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e LogEntryRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e LogEntryRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e LogEntryRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e LogEntryRequestValidationError) ErrorName() string { return "LogEntryRequestValidationError" }

// Error satisfies the builtin error interface
func (e LogEntryRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sLogEntryRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = LogEntryRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = LogEntryRequestValidationError{}

// Validate checks the field values on LogEntryResponse with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *LogEntryResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on LogEntryResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// LogEntryResponseMultiError, or nil if none found.
func (m *LogEntryResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *LogEntryResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Code

	// no validation rules for Message

	if len(errors) > 0 {
		return LogEntryResponseMultiError(errors)
	}

	return nil
}

// LogEntryResponseMultiError is an error wrapping multiple validation errors
// returned by LogEntryResponse.ValidateAll() if the designated constraints
// aren't met.
type LogEntryResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m LogEntryResponseMultiError) Error() string {
	msgs := make([]string, 0, len(m))
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m LogEntryResponseMultiError) AllErrors() []error { return m }

// LogEntryResponseValidationError is the validation error returned by
// LogEntryResponse.Validate if the designated constraints aren't met.
type LogEntryResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e LogEntryResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e LogEntryResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e LogEntryResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e LogEntryResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e LogEntryResponseValidationError) ErrorName() string { return "LogEntryResponseValidationError" }

// Error satisfies the builtin error interface
func (e LogEntryResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sLogEntryResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = LogEntryResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = LogEntryResponseValidationError{}

// Validate checks the field values on LogEntriesRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *LogEntriesRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on LogEntriesRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// LogEntriesRequestMultiError, or nil if none found.
func (m *LogEntriesRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *LogEntriesRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	for idx, item := range m.GetLogEntries() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, LogEntriesRequestValidationError{
						field:  fmt.Sprintf("LogEntries[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, LogEntriesRequestValidationError{
						field:  fmt.Sprintf("LogEntries[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return LogEntriesRequestValidationError{
					field:  fmt.Sprintf("LogEntries[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if len(errors) > 0 {
		return LogEntriesRequestMultiError(errors)
	}

	return nil
}

// LogEntriesRequestMultiError is an error wrapping multiple validation errors
// returned by LogEntriesRequest.ValidateAll() if the designated constraints
// aren't met.
type LogEntriesRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m LogEntriesRequestMultiError) Error() string {
	msgs := make([]string, 0, len(m))
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m LogEntriesRequestMultiError) AllErrors() []error { return m }

// LogEntriesRequestValidationError is the validation error returned by
// LogEntriesRequest.Validate if the designated constraints aren't met.
type LogEntriesRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e LogEntriesRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e LogEntriesRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e LogEntriesRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e LogEntriesRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e LogEntriesRequestValidationError) ErrorName() string {
	return "LogEntriesRequestValidationError"
}

// Error satisfies the builtin error interface
func (e LogEntriesRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sLogEntriesRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = LogEntriesRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = LogEntriesRequestValidationError{}

// Validate checks the field values on LogEntriesResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *LogEntriesResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on LogEntriesResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// LogEntriesResponseMultiError, or nil if none found.
func (m *LogEntriesResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *LogEntriesResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Code

	// no validation rules for Message

	if len(errors) > 0 {
		return LogEntriesResponseMultiError(errors)
	}

	return nil
}

// LogEntriesResponseMultiError is an error wrapping multiple validation errors
// returned by LogEntriesResponse.ValidateAll() if the designated constraints
// aren't met.
type LogEntriesResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m LogEntriesResponseMultiError) Error() string {
	msgs := make([]string, 0, len(m))
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m LogEntriesResponseMultiError) AllErrors() []error { return m }

// LogEntriesResponseValidationError is the validation error returned by
// LogEntriesResponse.Validate if the designated constraints aren't met.
type LogEntriesResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e LogEntriesResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e LogEntriesResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e LogEntriesResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e LogEntriesResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e LogEntriesResponseValidationError) ErrorName() string {
	return "LogEntriesResponseValidationError"
}

// Error satisfies the builtin error interface
func (e LogEntriesResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sLogEntriesResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = LogEntriesResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = LogEntriesResponseValidationError{}
