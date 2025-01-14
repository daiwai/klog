package app

import (
	"github.com/jotaen/klog/klog"
	"github.com/jotaen/klog/klog/service"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createMockConfigFromEnv(vs map[string]string) FromEnvVars {
	return FromEnvVars{GetVar: func(n string) string {
		return vs[n]
	}}
}

func TestCreatesNewDefaultConfig(t *testing.T) {
	c := NewDefaultConfig()
	assert.Equal(t, c.IsDebug.Value(), false)
	assert.Equal(t, c.Editor.Value(), "")
	assert.Equal(t, c.NoColour.Value(), false)
	assert.Equal(t, c.CpuKernels.Value(), 1)

	isRoundingSet := false
	c.DefaultRounding.Map(func(_ service.Rounding) {
		isRoundingSet = true
	})
	assert.False(t, isRoundingSet)

	isShouldTotalSet := false
	c.DefaultShouldTotal.Map(func(_ klog.ShouldTotal) {
		isShouldTotalSet = true
	})
	assert.False(t, isShouldTotalSet)
}

func TestSetsParamsMetadataIsHandledCorrectly(t *testing.T) {
	{
		c := NewDefaultConfig()
		assert.Equal(t, c.NoColour.Value(), false)
	}
	{
		c, _ := NewConfig(
			FromStaticValues{NumCpus: 1},
			createMockConfigFromEnv(map[string]string{
				"NO_COLOR": "1",
			}),
			FromConfigFile{""},
		)
		assert.Equal(t, c.NoColour.Value(), true)
	}
}

func TestSetsParamsFromEnv(t *testing.T) {
	// Read plain environment variables.
	{
		c, _ := NewConfig(
			FromStaticValues{NumCpus: 1},
			createMockConfigFromEnv(map[string]string{
				"EDITOR":     "subl",
				"KLOG_DEBUG": "1",
				"NO_COLOR":   "1",
			}),
			FromConfigFile{""},
		)
		assert.Equal(t, c.IsDebug.Value(), true)
		assert.Equal(t, c.NoColour.Value(), true)
		assert.Equal(t, c.Editor.Value(), "subl")
	}

	// `editor` from file would trump `$EDITOR` env variable.
	{
		c, _ := NewConfig(
			FromStaticValues{NumCpus: 1},
			createMockConfigFromEnv(map[string]string{
				"EDITOR": "subl",
			}),
			FromConfigFile{"editor = vi"},
		)
		assert.Equal(t, c.Editor.Value(), "vi")
	}
}

func TestSetsDefaultRoundingParamFromConfigFile(t *testing.T) {
	for _, x := range []struct {
		cfg string
		exp int
	}{
		{`default_rounding = 5m`, 5},
		{`default_rounding = 10m`, 10},
		{`default_rounding = 15m`, 15},
		{`default_rounding = 30m`, 30},
		{`default_rounding = 60m`, 60},
	} {
		c, _ := NewConfig(
			FromStaticValues{NumCpus: 1},
			createMockConfigFromEnv(map[string]string{}),
			FromConfigFile{x.cfg},
		)
		var value int
		c.DefaultRounding.Map(func(r service.Rounding) {
			value = r.ToInt()
		})
		assert.Equal(t, x.exp, value)
	}
}

func TestSetsDefaultShouldTotalParamFromConfigFile(t *testing.T) {
	for _, x := range []struct {
		cfg string
		exp string
	}{
		{`default_should_total = 8h30m!`, "8h30m!"},
	} {
		c, _ := NewConfig(
			FromStaticValues{NumCpus: 1},
			createMockConfigFromEnv(map[string]string{}),
			FromConfigFile{x.cfg},
		)
		var value string
		c.DefaultShouldTotal.Map(func(s klog.ShouldTotal) {
			value = s.ToString()
		})
		assert.Equal(t, x.exp, value)
	}
}

func TestSetsDateFormatParamFromConfigFile(t *testing.T) {
	for _, x := range []struct {
		cfg string
		exp bool
	}{
		{`date_format = YYYY-MM-DD`, true},
		{`date_format = YYYY/MM/DD`, false},
	} {
		c, _ := NewConfig(
			FromStaticValues{NumCpus: 1},
			createMockConfigFromEnv(map[string]string{}),
			FromConfigFile{x.cfg},
		)
		var value bool
		c.DateUseDashes.Map(func(s bool) {
			value = s
		})
		assert.Equal(t, x.exp, value)
	}
}

func TestSetTimeFormatParamFromConfigFile(t *testing.T) {
	for _, x := range []struct {
		cfg string
		exp bool
	}{
		{`time_convention = 24h`, true},
		{`time_convention = 12h`, false},
	} {
		c, _ := NewConfig(
			FromStaticValues{NumCpus: 1},
			createMockConfigFromEnv(map[string]string{}),
			FromConfigFile{x.cfg},
		)
		var value bool
		c.TimeUse24HourClock.Map(func(s bool) {
			value = s
		})
		assert.Equal(t, x.exp, value)
	}
}

func TestIgnoresUnknownPropertiesInConfigFile(t *testing.T) {
	for _, tml := range []string{`
unknown_property = 1
what_is_this = true
`,
	} {
		_, err := NewConfig(
			FromStaticValues{NumCpus: 1},
			createMockConfigFromEnv(map[string]string{}),
			FromConfigFile{tml},
		)
		assert.Nil(t, err)
	}
}

func TestIgnoresEmptyConfigFileOrEmptyParameters(t *testing.T) {
	for _, tml := range []string{
		``,
		`editor = `,
		`default_rounding =`,
		`default_should_total = `,
		`date_format = `,
		`time_convention = `,
	} {
		_, err := NewConfig(
			FromStaticValues{NumCpus: 1},
			createMockConfigFromEnv(map[string]string{}),
			FromConfigFile{tml},
		)
		assert.Nil(t, err)
	}
}

func TestRejectsInvalidConfigFile(t *testing.T) {
	for _, tml := range []string{
		`default_rounding = true`,              // Wrong type
		`default_rounding = 20m`,               // Invalid value
		`default_should_total = [true, false]`, // Wrong type
		`default_should_total = 15`,            // Invalid value
		`date_format = [true, false]`,          // Wrong type
		`date_format = YYYY.MM.DD`,             // Invalid value
		`time_convention = [true, false]`,      // Wrong type
		`time_convention = 2h`,                 // Invalid value
	} {
		_, err := NewConfig(
			FromStaticValues{NumCpus: 1},
			createMockConfigFromEnv(map[string]string{}),
			FromConfigFile{tml},
		)
		assert.Error(t, err)
	}
}
