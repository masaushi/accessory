package cmd_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/spf13/afero"

	"github.com/masaushi/accessory/cmd"
)

func TestExecute(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		cmd    string
		output string
	}{
		"Getter": {
			cmd:    "accessory -type Tester testdata/getter",
			output: "testdata/getter/tester_accessor.go",
		},
		"Setter": {
			cmd:    "accessory -type Tester testdata/setter",
			output: "testdata/setter/tester_accessor.go",
		},
		"GetterAndSetter": {
			cmd:    "accessory -type Tester testdata/getter_and_setter",
			output: "testdata/getter_and_setter/tester_accessor.go",
		},
		"CamelCaseNodeName": {
			cmd:    "accessory -type Tester testdata/camel_case_node_name",
			output: "testdata/camel_case_node_name/tester_accessor.go",
		},
		"IgnoreFields": {
			cmd:    "accessory -type Tester testdata/ignore_fields",
			output: "testdata/ignore_fields/tester_accessor.go",
		},
		"NoDefaultValue": {
			cmd:    "accessory -type Tester testdata/no_default_value",
			output: "testdata/no_default_value/tester_accessor.go",
		},
		"ImportPackages": {
			cmd:    "accessory -type Tester testdata/import_packages",
			output: "testdata/import_packages/tester_accessor.go",
		},
		"WithOutput": {
			cmd:    "accessory -type Tester -output my_accessor.go testdata/with_output",
			output: "testdata/with_output/my_accessor.go",
		},
		"WithReceiver": {
			cmd:    "accessory -type Tester -receiver tester testdata/with_receiver",
			output: "testdata/with_receiver/tester_accessor.go",
		},
		"WithLock": {
			cmd:    "accessory -type Tester -lock lock testdata/with_lock",
			output: "testdata/with_lock/tester_accessor.go",
		},
		"TimeField": {
			cmd:    "accessory -type Tester testdata/time_field",
			output: "testdata/time_field/tester_accessor.go",
		},
	}

	fs := afero.NewMemMapFs()
	snapshot := cupaloy.New(
		cupaloy.SnapshotSubdirectory("testdata/.snapshots"),
		cupaloy.SnapshotFileExtension(".go"),
	)

	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			args := strings.Split(tt.cmd, " ")
			cmd.Execute(fs, args)

			output, _ := filepath.Abs(tt.output)

			exists, err := afero.Exists(fs, output)
			if err != nil {
				t.Fatal(err)
			}
			if !exists {
				t.Fatalf("file %s not exists", output)
			}

			file, err := afero.ReadFile(fs, output)
			if err != nil {
				t.Fatal(err)
			}

			snapshot.SnapshotT(t, file)
		})
	}
}
