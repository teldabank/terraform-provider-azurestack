package azurestack

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestObjectIdFromJWT(t *testing.T) {
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"oid":"11111111-2222-3333-4444-555555555555","appid":"app"}`))
	token := "header." + payload + ".signature"

	actual, err := objectIdFromJWT(token)
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	expected := "11111111-2222-3333-4444-555555555555"
	if actual != expected {
		t.Fatalf("expected %q but got %q", expected, actual)
	}
}

func TestObjectIdFromJWT_Invalid(t *testing.T) {
	if _, err := objectIdFromJWT("not-a-jwt"); err == nil {
		t.Fatal("expected an error for a non-JWT input")
	}

	if _, err := objectIdFromJWT("header." + base64.RawURLEncoding.EncodeToString([]byte("not json")) + ".sig"); err == nil {
		t.Fatal("expected an error for a payload that is not JSON")
	}

	// padded base64 (which JWT does not use) must not silently decode
	padded := base64.URLEncoding.EncodeToString([]byte(`{"oid":"x"}`))
	if strings.HasSuffix(padded, "=") {
		if _, err := objectIdFromJWT("header." + padded + ".sig"); err == nil {
			t.Fatal("expected an error for padded base64")
		}
	}
}
