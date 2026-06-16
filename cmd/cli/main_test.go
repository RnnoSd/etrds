package main

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	//"github.com/stretchr/testify/require"
)

func TestHelpConfigurations(t *testing.T) {
	testedCmds := []*cobra.Command{
		rootCmd,
		statusCmd,
	}
	for _, testedCmd := range testedCmds {
		assert.NotRegexpf(t, "still not done", testedCmd.Long, "Long descripction Command %s is still not fully implemented", testedCmd.Name())
		assert.NotRegexpf(t, "still not done", testedCmd.Short, "Long descripction Command %s is still not fully implemented", testedCmd.Name())
		assert.NotRegexpf(t, "still not done", testedCmd.Use, "Long descripction Command %s is still not fully implemented", testedCmd.Name())
	}
}
