package password_test

import (
	"strings"
	"testing"

	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/auth/internal/password"
)

func TestHashAndVerify(t *testing.T) {
	t.Parallel()

	h, err := password.Hash("s3cret!")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if !strings.HasPrefix(h, "$2a$") && !strings.HasPrefix(h, "$2b$") {
		t.Fatalf("unexpected bcrypt prefix in %q", h)
	}

	ok, err := password.Verify(h, "s3cret!")
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !ok {
		t.Fatalf("Verify returned false for the correct password")
	}

	ok, err = password.Verify(h, "wrong")
	if err != nil {
		t.Fatalf("Verify wrong: %v", err)
	}
	if ok {
		t.Fatalf("Verify returned true for the wrong password")
	}
}

func TestHashRejectsEmpty(t *testing.T) {
	t.Parallel()
	if _, err := password.Hash(""); err == nil {
		t.Fatal("Hash(\"\") = nil, want error")
	}
}

func TestVerifyMalformed(t *testing.T) {
	t.Parallel()
	if _, err := password.Verify("not-bcrypt", "x"); err == nil {
		t.Fatal("Verify(malformed) = nil, want error")
	}
}

func TestIsBcrypt(t *testing.T) {
	t.Parallel()
	cases := map[string]bool{
		"$2a$10$abcdefghijklmnopqrstuvwxyz": true,
		"$2b$10$abcdefghijklmnopqrstuvwxyz": true,
		"$2y$10$abcdefghijklmnopqrstuvwxyz": true,
		"plaintext":                         false,
		"":                                  false,
		"$3a$10$abc":                        false,
	}
	for in, want := range cases {
		if got := password.IsBcrypt(in); got != want {
			t.Errorf("IsBcrypt(%q) = %v, want %v", in, got, want)
		}
	}
}
