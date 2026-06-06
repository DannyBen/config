package tomldoc

import "testing"

func TestValidateRejectsDuplicateKeys(t *testing.T) {
	err := validateTOML("font = \"arial\"\nfont = \"arial\"\n")

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateRejectsScalarThenArrayTableRedefinition(t *testing.T) {
	err := validateTOML("servers = \"asd\"\n\n[[servers]]\nname = \"api\"\n")

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateRejectsScalarThenTableRedefinition(t *testing.T) {
	err := validateTOML("server = \"localhost\"\n\n[server]\nport = 3000\n")

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestOperationsRejectInvalidTOMLBeforePlanning(t *testing.T) {
	tests := []struct {
		name string
		run  func(string) error
	}{
		{
			name: "get",
			run: func(source string) error {
				_, err := Get(source, "font")
				return err
			},
		},
		{
			name: "list",
			run: func(source string) error {
				_, err := List(source, "")
				return err
			},
		},
		{
			name: "set",
			run: func(source string) error {
				_, err := Set(source, "font", "mono")
				return err
			},
		},
		{
			name: "unset",
			run: func(source string) error {
				_, err := Unset(source, "font")
				return err
			},
		},
		{
			name: "delete",
			run: func(source string) error {
				_, err := Delete(source, "style", nil)
				return err
			},
		},
	}

	source := "font = \"arial\"\nfont = \"arial\"\n"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.run(source); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
