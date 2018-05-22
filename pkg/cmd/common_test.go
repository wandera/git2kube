package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"reflect"
	"testing"
)

func TestExpandArgs(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		env    map[string]string
		result []string
	}{
		{
			name: "No Env",
			args: []string{
				"arg1",
			},
			result: []string{
				"arg1",
			},
		},
		{
			name: "Simple Env",
			args: []string{
				"$ENV",
			},
			env: map[string]string{
				"ENV": "test",
			},
			result: []string{
				"test",
			},
		},
		{
			name: "Simple Env Multiple",
			args: []string{
				"$ENV",
				"${ENV}",
				"$ENV",
			},
			env: map[string]string{
				"ENV": "test",
			},
			result: []string{
				"test",
				"test",
				"test",
			},
		},
		{
			name: "Interpolation",
			args: []string{
				"This is $ENV property",
				"This is ${ENV} property",
				"This is $ENV property",
			},
			env: map[string]string{
				"ENV": "test",
			},
			result: []string{
				"This is test property",
				"This is test property",
				"This is test property",
			},
		},
		{
			name: "Multiple Env",
			args: []string{
				"This is $ENV property $ENV2",
			},
			env: map[string]string{
				"ENV":  "test",
				"ENV2": "test2",
			},
			result: []string{
				"This is test property test2",
			},
		},
	}

	for _, c := range cases {
		setEnvFromMap(c.env)
		command := &cobra.Command{}
		ExpandArgs(command, c.args)
		res := command.Flags().Args()
		if !reflect.DeepEqual(res, c.result) {
			t.Errorf("%s case failed: result args mismatch expected %s but got %s instead", c.name, c.result, res)
		}
		unsetEnvFromMap(c.env)
	}
}

func setEnvFromMap(env map[string]string) {
	for k, v := range env {
		os.Setenv(k, v)
	}
}

func unsetEnvFromMap(env map[string]string) {
	for k := range env {
		os.Unsetenv(k)
	}
}
