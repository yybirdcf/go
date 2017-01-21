package utils

import "errors"

func Str2Err(s string) error {
	if s == "" {
		return nil
	}
	return errors.New(s)
}

func Err2Str(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
