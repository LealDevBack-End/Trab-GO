package main

import "testing"

func TestIsCPF_Valid(t *testing.T) {
	valid := []string{
		"529.982.247-25",
		"52998224725",
		"111.444.777-35",
	}
	for _, cpf := range valid {
		if !isCPF(cpf) {
			t.Errorf("isCPF(%q) = false, want true", cpf)
		}
	}
}

func TestIsCPF_Invalid(t *testing.T) {
	invalid := []string{
		"",
		"123",
		"12345678901",
		"11111111111",
		"00000000000",
		"529.982.247-26",
		"52998224726",
		"abc",
	}
	for _, cpf := range invalid {
		if isCPF(cpf) {
			t.Errorf("isCPF(%q) = true, want false", cpf)
		}
	}
}

func TestOnlyDigits(t *testing.T) {
	got := onlyDigits("529.982.247-25")
	want := "52998224725"
	if got != want {
		t.Errorf("onlyDigits = %q, want %q", got, want)
	}
}

func TestAllDigitsEqual(t *testing.T) {
	if !allDigitsEqual("11111111111") {
		t.Error("allDigitsEqual should be true for repeated digits")
	}
	if allDigitsEqual("52998224725") {
		t.Error("allDigitsEqual should be false for varied digits")
	}
}
