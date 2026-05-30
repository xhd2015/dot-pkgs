package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/dupstat"
)

func testdataDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(wd, "testdata")
}

func TestIntegrationCrossPkgDup(t *testing.T) {
	dir := filepath.Join(testdataDir(t), "cross-pkg-dup")
	var buf bytes.Buffer
	err := runWithOutput([]string{"--dir", dir, "--threshold", "0.3", "--ngram", "5"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	if !strings.Contains(output, "CROSS-PACKAGE") {
		t.Errorf("expected CROSS-PACKAGE group, got:\n%s", output)
	}
	if !strings.Contains(output, "ValidateEmail") {
		t.Errorf("expected ValidateEmail, got:\n%s", output)
	}
	if !strings.Contains(output, "CheckEmail") {
		t.Errorf("expected CheckEmail, got:\n%s", output)
	}
	if !strings.Contains(output, "HashPassword") {
		t.Errorf("expected HashPassword, got:\n%s", output)
	}
	if !strings.Contains(output, "EncryptPassword") {
		t.Errorf("expected EncryptPassword, got:\n%s", output)
	}
}

func TestIntegrationSamePkgDup(t *testing.T) {
	dir := filepath.Join(testdataDir(t), "same-pkg-dup")
	var buf bytes.Buffer
	err := runWithOutput([]string{"--dir", dir, "--threshold", "0.3", "--ngram", "5"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	if !strings.Contains(output, "SAME-PACKAGE") {
		t.Errorf("expected SAME-PACKAGE group, got:\n%s", output)
	}
	if !strings.Contains(output, "ReadConfig") {
		t.Errorf("expected ReadConfig, got:\n%s", output)
	}
	if !strings.Contains(output, "WriteConfig") {
		t.Errorf("expected WriteConfig, got:\n%s", output)
	}
}

func TestIntegrationStructuralDup(t *testing.T) {
	dir := filepath.Join(testdataDir(t), "structural-dup")
	var buf bytes.Buffer
	err := runWithOutput([]string{"--dir", dir, "--threshold", "0.3", "--ngram", "5"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	if !strings.Contains(output, "ProcessOrder") {
		t.Errorf("expected ProcessOrder, got:\n%s", output)
	}
	if !strings.Contains(output, "ProcessPayment") {
		t.Errorf("expected ProcessPayment, got:\n%s", output)
	}
}

func TestIntegrationNoDup(t *testing.T) {
	dir := filepath.Join(testdataDir(t), "no-dup")
	var buf bytes.Buffer
	err := runWithOutput([]string{"--dir", dir, "--threshold", "0.3", "--ngram", "5"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	if strings.Contains(output, "similarity") {
		t.Errorf("expected no similar pairs, got:\n%s", output)
	}
}

func TestIntegrationSubsetDup(t *testing.T) {
	dir := filepath.Join(testdataDir(t), "subset-dup")
	var buf bytes.Buffer
	err := runWithOutput([]string{"--dir", dir, "--threshold", "0.3", "--ngram", "5"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	if !strings.Contains(output, "DoQuick") {
		t.Errorf("expected DoQuick, got:\n%s", output)
	}
	if !strings.Contains(output, "DoFull") {
		t.Errorf("expected DoFull, got:\n%s", output)
	}
}

func TestIntegrationWordStatDup(t *testing.T) {
	dir := filepath.Join(testdataDir(t), "wordstat-dup")
	var buf bytes.Buffer
	err := runWithOutput([]string{"--dir", dir, "--threshold", "0.3", "--algorithm", "wordstat"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	if !strings.Contains(output, "ProcessUser") {
		t.Errorf("expected ProcessUser, got:\n%s", output)
	}
	if !strings.Contains(output, "HandleRequest") {
		t.Errorf("expected HandleRequest, got:\n%s", output)
	}
	if !strings.Contains(output, "wordstat") {
		t.Errorf("expected wordstat scores, got:\n%s", output)
	}
	if strings.Contains(output, "similarity") {
		if !strings.Contains(output, "wordstat") {
			t.Errorf("wordstat should use wordstat score format")
		}
	}
}

func TestFindModuleRoot(t *testing.T) {
	dir := filepath.Join(testdataDir(t), "cross-pkg-dup")
	root, err := dupstat.FindModuleRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	absDir, _ := filepath.Abs(dir)
	absRoot, _ := filepath.Abs(root)
	if absRoot != absDir {
		t.Errorf("expected module root %s, got %s", absDir, absRoot)
	}
}
