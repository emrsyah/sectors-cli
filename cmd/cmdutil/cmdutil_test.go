package cmdutil

import (
	"testing"

	"github.com/spf13/cobra"
)

// newTestCmd returns a command with the flag set but NOT marked changed.
func newTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "t", Run: func(*cobra.Command, []string) {}}
	cmd.Flags().String("name", "default", "")
	cmd.Flags().Int("num", 0, "")
	cmd.Flags().Bool("flag", false, "")
	cmd.Flags().Float64("ratio", 0, "")
	return cmd
}

func TestOptStr_UnsetIsNil(t *testing.T) {
	cmd := newTestCmd()
	if got := OptStr(cmd, "name", "default"); got != nil {
		t.Errorf("unset flag should be nil, got %v", *got)
	}
}

func TestOptStr_SetReturnsPointer(t *testing.T) {
	cmd := newTestCmd()
	_ = cmd.Flags().Set("name", "banks")
	got := OptStr(cmd, "name", "banks")
	if got == nil || *got != "banks" {
		t.Errorf("set flag = %v, want banks", got)
	}
}

func TestOptInt_ZeroButSetIsNotNil(t *testing.T) {
	cmd := newTestCmd()
	// Explicitly setting to the zero value must still produce a non-nil pointer
	// (the whole reason the Opt helpers exist).
	_ = cmd.Flags().Set("num", "0")
	got := OptInt(cmd, "num", 0)
	if got == nil || *got != 0 {
		t.Errorf("explicit zero = %v, want non-nil 0", got)
	}
}

func TestOptBoolAndFloat(t *testing.T) {
	cmd := newTestCmd()
	if OptBool(cmd, "flag", false) != nil {
		t.Error("unset bool should be nil")
	}
	if OptFloat(cmd, "ratio", 0) != nil {
		t.Error("unset float should be nil")
	}
	_ = cmd.Flags().Set("flag", "true")
	_ = cmd.Flags().Set("ratio", "1.5")
	if b := OptBool(cmd, "flag", true); b == nil || *b != true {
		t.Errorf("set bool = %v", b)
	}
	if f := OptFloat(cmd, "ratio", 1.5); f == nil || *f != 1.5 {
		t.Errorf("set float = %v", f)
	}
}

type myEnum string

func TestOptEnum(t *testing.T) {
	cmd := newTestCmd()
	if OptEnum[myEnum](cmd, "name", "x") != nil {
		t.Error("unset enum should be nil")
	}
	_ = cmd.Flags().Set("name", "retail")
	got := OptEnum[myEnum](cmd, "name", "retail")
	if got == nil || *got != myEnum("retail") {
		t.Errorf("set enum = %v", got)
	}
}

func TestNormalizers(t *testing.T) {
	if Sym("bbca") != "BBCA" {
		t.Error("Sym should uppercase")
	}
	if Code("mg") != "MG" {
		t.Error("Code should uppercase")
	}
	if Slug("Banks") != "banks" {
		t.Error("Slug should lowercase")
	}
}
