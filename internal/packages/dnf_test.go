package packages

import "testing"

func TestDnf_Name(t *testing.T) {
	d := NewDnf()
	if got := d.Name(); got != "dnf" {
		t.Errorf("Name() = %q, want %q", got, "dnf")
	}
}

func TestDnf_Install_Validation(t *testing.T) {
	d := NewDnf()

	err := d.Install("")
	if err == nil {
		t.Error("Install(\"\") should return error for empty package name")
	}
}

func TestDnf_Uninstall_Validation(t *testing.T) {
	d := NewDnf()

	err := d.Uninstall("")
	if err == nil {
		t.Error("Uninstall(\"\") should return error for empty package name")
	}
}

func TestDnf_IsInstalled_Validation(t *testing.T) {
	d := NewDnf()

	_, err := d.IsInstalled("")
	if err == nil {
		t.Error("IsInstalled(\"\") should return error for empty package name")
	}
}

func TestDnf_IsAvailable(t *testing.T) {
	d := NewDnf()
	_ = d.IsAvailable()
}
