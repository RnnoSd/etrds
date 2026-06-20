package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"gopkg.in/yaml.v3"
)

type versionSetting struct {
	Number string `json:"version" yaml:"version"`
}

type Engine string

type Paths []string

func (p *Paths) UnmarshalJSON(data []byte) error {
	if string(data[0]) == `[` {
		var out []string
		if err := json.Unmarshal(data, &out); err != nil {
			return nil
		}
		*p = Paths(out)
		return nil
	}
	var out string
	if err := json.Unmarshal(data, &out); err != nil {
		return nil
	}
	*p = Paths([]string{out})
	return nil
}

func (p *Paths) UnmarshalYAML(unmarshal func(any) error) error {
	out := []string{}
	if sliceErr := unmarshal(&out); sliceErr != nil {
		var ele string
		if strErr := unmarshal(&ele); strErr != nil {
			return strErr
		}
		out = []string{ele}
	}

	*p = Paths(out)
	return nil
}

const (
	EngineMySQL Engine = "mysql"
)

type Config struct {
	Version string               `json:"version" yaml:"version"`
	Servers []Server             `json:"servers" yaml:"servers"`
	SQL     []SQL                `json:"sql" yaml:"sql"`
	Plugins []Plugin             `json:"plugins" yaml:"plugins"`
	Rules   []Rule               `json:"rules" yaml:"rules"`
	Options map[string]yaml.Node `json:"options" yaml:"options"`
}

type Server struct {
	Name   string `json:"name,omitempty" yaml:"name"`
	Engine Engine `json:"engine,omitempty" yaml:"engine"`
	URI    string `json:"uri" yaml:"uri"`
}

type Database struct {
	URI     string `json:"uri" yaml:"uri"`
	Managed bool   `json:"managed" yaml:"managed"`
}

type Cloud struct {
	Organization string `json:"organization" yaml:"organization"`
	Project      string `json:"project" yaml:"project"`
	Hostname     string `json:"hostname" yaml:"hostname"`
	AuthToken    string `json:"-" yaml:"-"`
}

type Plugin struct {
	Name    string   `json:"name" yaml:"name"`
	Env     []string `json:"env" yaml:"env"`
	Process *struct {
		Cmd    string `json:"cmd" yaml:"cmd"`
		Format string `json:"format" yaml:"format"`
	} `json:"process" yaml:"process"`
	WASM *struct {
		URL    string `json:"url" yaml:"url"`
		SHA256 string `json:"sha256" yaml:"sha256"`
	} `json:"wasm" yaml:"wasm"`
}

type Rule struct {
	Name string `json:"name" yaml:"name"`
	Rule string `json:"rule" yaml:"rule"`
	Msg  string `json:"message" yaml:"message"`
}

type SQL struct {
	Name                 string    `json:"name" yaml:"name"`
	Engine               Engine    `json:"engine,omitempty" yaml:"engine"`
	Schema               Paths     `json:"schema" yaml:"schema"`
	Queries              Paths     `json:"queries" yaml:"queries"`
	Database             *Database `json:"database" yaml:"database"`
	StrictFunctionChecks bool      `json:"strict_function_checks" yaml:"strict_function_checks"`
	StrictOrderBy        *bool     `json:"strict_order_by" yaml:"strict_order_by"`
	Codegen              []Codegen `json:"codegen" yaml:"codegen"`
	Rules                []string  `json:"rules" yaml:"rules"`
	Analyzer             Analyzer  `json:"analyzer" yaml:"analyzer"`
}

// AnalyzerDatabase represents the database analyzer setting.
// It can be a boolean (true/false) or the string "only" for database-only mode.
type AnalyzerDatabase struct {
	value  *bool // nil means not set, true/false for boolean values
	isOnly bool  // true when set to "only"
}

// IsEnabled returns true if the database analyzer should be used.
// Returns true for both `true` and `"only"` settings.
func (a AnalyzerDatabase) IsEnabled() bool {
	if a.isOnly {
		return true
	}
	return a.value == nil || *a.value
}

// IsOnly returns true if the analyzer is set to "only" mode.
func (a AnalyzerDatabase) IsOnly() bool {
	return a.isOnly
}

func (a *AnalyzerDatabase) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as boolean first
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		a.value = &b
		a.isOnly = false
		return nil
	}

	// Try to unmarshal as string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if s == "only" {
			a.isOnly = true
			a.value = nil
			return nil
		}
		return errors.New("analyzer.database must be true, false, or \"only\"")
	}

	return errors.New("analyzer.database must be true, false, or \"only\"")
}

func (a *AnalyzerDatabase) UnmarshalYAML(unmarshal func(any) error) error {
	// Try to unmarshal as boolean first
	var b bool
	if err := unmarshal(&b); err == nil {
		a.value = &b
		a.isOnly = false
		return nil
	}

	// Try to unmarshal as string
	var s string
	if err := unmarshal(&s); err == nil {
		if s == "only" {
			a.isOnly = true
			a.value = nil
			return nil
		}
		return errors.New("analyzer.database must be true, false, or \"only\"")
	}

	return errors.New("analyzer.database must be true, false, or \"only\"")
}

type Analyzer struct {
	Database AnalyzerDatabase `json:"database" yaml:"database"`
}

// TODO: Figure out a better name for this
type Codegen struct {
	Out     string    `json:"out" yaml:"out"`
	Plugin  string    `json:"plugin" yaml:"plugin"`
	Options yaml.Node `json:"options" yaml:"options"`
}

var (
	ErrMissingEngine  = errors.New("unknown engine")
	ErrMissingVersion = errors.New("no version number")
	ErrNoOutPath      = errors.New("no output path")
	ErrNoPackagePath  = errors.New("missing package path")
	ErrNoPackages     = errors.New("no packages")
	ErrNoQuerierType  = errors.New("no querier emit type enabled")
	ErrUnknownEngine  = errors.New("invalid engine")
	ErrUnknownVersion = errors.New("invalid version number")
)

var (
	ErrPluginBuiltin      = errors.New("a built-in plugin with that name already exists")
	ErrPluginNoName       = errors.New("missing plugin name")
	ErrPluginExists       = errors.New("a plugin with that name already exists")
	ErrPluginNotFound     = errors.New("no plugin found")
	ErrPluginNoType       = errors.New("plugin: field `process` or `wasm` required")
	ErrPluginBothTypes    = errors.New("plugin: `process` and `wasm` cannot both be defined")
	ErrPluginProcessNoCmd = errors.New("plugin: missing process command")
)

var (
	ErrInvalidDatabase          = errors.New("database must be managed or have a non-empty URI")
	ErrManagedDatabaseNoProject = errors.New(`managed databases require a cloud project

If you don't have a project, you can create one from the sqlc Cloud
dashboard at https://dashboard.sqlc.dev/. If you have a project, ensure
you've set its id as the value of the "project" field within the "cloud"
section of your sqlc configuration. The id will look similar to
"01HA8TWGMYPHK0V2GGMB3R2TP9".`)
)

var ErrManagedDatabaseNoAuthToken = errors.New(`managed databases require an auth token

If you don't have an auth token, you can create one from the sqlc Cloud
dashboard at https://dashboard.sqlc.dev/. If you have an auth token, ensure
you've set it as the value of the SQLC_AUTH_TOKEN environment variable.`)

func ParseConfig(rd io.Reader) (Config, error) {
	var buf bytes.Buffer
	var config Config
	var version versionSetting

	ver := io.TeeReader(rd, &buf)
	dec := yaml.NewDecoder(ver)
	if err := dec.Decode(&version); err != nil {
		return config, err
	}
	if version.Number == "" {
		return config, ErrMissingVersion
	}
	var err error
	switch version.Number {
	case "1":
		config, err = v1ParseConfig(&buf)
		if err != nil {
			return config, err
		}
	case "2":
		config, err = v2ParseConfig(&buf)
		if err != nil {
			return config, err
		}
	default:
		return config, ErrUnknownVersion
	}
	return config, nil
}
