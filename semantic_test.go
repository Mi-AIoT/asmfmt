package asmfmt

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestOptionalSemanticEquivalence(t *testing.T) {
	as := strings.TrimSpace(os.Getenv("ASMFMT_AS"))
	if as == "" {
		t.Skip("set ASMFMT_AS to run optional assembler equivalence checks")
	}
	objdump := strings.TrimSpace(os.Getenv("ASMFMT_OBJDUMP"))
	if objdump == "" {
		objdump = "objdump"
	}
	asFlags := splitEnvList("ASMFMT_ASFLAGS")
	objdumpFlags := splitEnvList("ASMFMT_OBJDUMPFLAGS")

	fixtures := []string{
		"testdata/riscv_pseudo.in",
		"testdata/riscv_relocations_labels.in",
		"testdata/riscv_csr.in",
		"testdata/riscv_insn.in",
	}
	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(filepath.Base(fixture), func(t *testing.T) {
			src, err := os.ReadFile(fixture)
			if err != nil {
				t.Fatal(err)
			}
			formatted, err := Format(strings.NewReader(string(src)))
			if err != nil {
				t.Fatal(err)
			}

			origObj := assembleFixture(t, as, asFlags, fixture, src)
			fmtObj := assembleFixture(t, as, asFlags, fixture, formatted)
			origDump := objdumpFixture(t, objdump, objdumpFlags, origObj)
			fmtDump := objdumpFixture(t, objdump, objdumpFlags, fmtObj)
			if !bytes.Equal(origDump, fmtDump) {
				t.Fatalf("semantic mismatch for %s\n--- original ---\n%s\n--- formatted ---\n%s", fixture, origDump, fmtDump)
			}
		})
	}
}

func assembleFixture(t *testing.T, as string, flags []string, fixture string, src []byte) string {
	t.Helper()
	dir := t.TempDir()
	in := filepath.Join(dir, filepath.Base(fixture))
	out := filepath.Join(dir, filepath.Base(fixture)+".o")
	if err := os.WriteFile(in, src, 0o666); err != nil {
		t.Fatal(err)
	}
	args := append([]string{}, flags...)
	args = append(args, "-o", out, in)
	cmd := exec.Command(as, args...)
	cmd.Env = os.Environ()
	if data, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("assemble %s: %v\n%s", fixture, err, data)
	}
	return out
}

func objdumpFixture(t *testing.T, objdump string, flags []string, object string) []byte {
	t.Helper()
	args := append([]string{}, flags...)
	args = append(args, "-dr", object)
	cmd := exec.Command(objdump, args...)
	cmd.Env = os.Environ()
	data, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("objdump %s: %v\n%s", object, err, data)
	}
	return normalizeObjdump(data)
}

func normalizeObjdump(data []byte) []byte {
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	var keep []string
	for _, line := range lines {
		if strings.HasSuffix(line, "file format elf64-littleriscv") || strings.HasSuffix(line, "file format elf32-littleriscv") {
			continue
		}
		if strings.HasPrefix(line, "/") {
			continue
		}
		keep = append(keep, line)
	}
	return []byte(strings.Join(keep, "\n"))
}

func splitEnvList(name string) []string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return nil
	}
	return strings.Fields(value)
}
