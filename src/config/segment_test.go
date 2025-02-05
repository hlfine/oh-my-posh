package config

import (
	"encoding/json"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments"

	"github.com/stretchr/testify/assert"
)

const (
	cwd = "Projects/oh-my-posh"
)

func TestMapSegmentWriterCanMap(t *testing.T) {
	sc := &Segment{
		Type: SESSION,
	}
	env := new(mock.Environment)
	err := sc.MapSegmentWithWriter(env)
	assert.NoError(t, err)
	assert.NotNil(t, sc.writer)
}

func TestMapSegmentWriterCannotMap(t *testing.T) {
	sc := &Segment{
		Type: "nilwriter",
	}
	env := new(mock.Environment)
	err := sc.MapSegmentWithWriter(env)
	assert.Error(t, err)
}

func TestParseTestConfig(t *testing.T) {
	segmentJSON :=
		`
		{
			"type": "path",
			"style": "powerline",
			"powerline_symbol": "\uE0B0",
			"foreground": "#ffffff",
			"background": "#61AFEF",
			"properties": {
				"style": "folder"
			},
			"exclude_folders": [
				"/super/secret/project"
			]
		}
		`
	segment := &Segment{}
	err := json.Unmarshal([]byte(segmentJSON), segment)
	assert.NoError(t, err)
}

func TestShouldIncludeFolder(t *testing.T) {
	cases := []struct {
		Case     string
		Included bool
		Excluded bool
		Expected bool
	}{
		{Case: "Include", Included: true, Excluded: false, Expected: true},
		{Case: "Exclude", Included: false, Excluded: true, Expected: false},
		{Case: "Include & Exclude", Included: true, Excluded: true, Expected: false},
		{Case: "!Include & !Exclude", Included: false, Excluded: false, Expected: false},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("GOOS").Return(runtime.LINUX)
		env.On("Home").Return("")
		env.On("Pwd").Return(cwd)
		env.On("DirMatchesOneOf", cwd, []string{"Projects/oh-my-posh"}).Return(tc.Included)
		env.On("DirMatchesOneOf", cwd, []string{"Projects/nope"}).Return(tc.Excluded)
		segment := &Segment{
			IncludeFolders: []string{"Projects/oh-my-posh"},
			ExcludeFolders: []string{"Projects/nope"},
			env:            env,
		}
		got := segment.shouldIncludeFolder()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestGetColors(t *testing.T) {
	cases := []struct {
		Case       string
		Expected   color.Ansi
		Default    color.Ansi
		Region     string
		Profile    string
		Templates  []string
		Background bool
	}{
		{Case: "No template - foreground", Expected: "color", Background: false, Default: "color"},
		{Case: "No template - background", Expected: "color", Background: true, Default: "color"},
		{Case: "Nil template", Expected: "color", Default: "color", Templates: nil},
		{
			Case:     "Template - default",
			Expected: "color",
			Default:  "color",
			Templates: []string{
				"{{if contains \"john\" .Profile}}color2{{end}}",
			},
			Profile: "doe",
		},
		{
			Case:     "Template - override",
			Expected: "color2",
			Default:  "color",
			Templates: []string{
				"{{if contains \"john\" .Profile}}color2{{end}}",
			},
			Profile: "john",
		},
		{
			Case:     "Template - override multiple",
			Expected: "color3",
			Default:  "color",
			Templates: []string{
				"{{if contains \"doe\" .Profile}}color2{{end}}",
				"{{if contains \"john\" .Profile}}color3{{end}}",
			},
			Profile: "john",
		},
		{
			Case:     "Template - override multiple no match",
			Expected: "color",
			Default:  "color",
			Templates: []string{
				"{{if contains \"doe\" .Profile}}color2{{end}}",
				"{{if contains \"philip\" .Profile}}color3{{end}}",
			},
			Profile: "john",
		},
	}
	for _, tc := range cases {
		segment := &Segment{
			writer: &segments.Aws{
				Profile: tc.Profile,
				Region:  tc.Region,
			},
		}

		if tc.Background {
			segment.Background = tc.Default
			segment.BackgroundTemplates = tc.Templates
			bgColor := segment.ResolveBackground()
			assert.Equal(t, tc.Expected, bgColor, tc.Case)
			continue
		}

		segment.Foreground = tc.Default
		segment.ForegroundTemplates = tc.Templates
		fgColor := segment.ResolveForeground()
		assert.Equal(t, tc.Expected, fgColor, tc.Case)
	}
}

func TestEvaluateNeeds(t *testing.T) {
	cases := []struct {
		Segment *Segment
		Case    string
		Needs   []string
	}{
		{
			Case: "No needs",
			Segment: &Segment{
				Template: "foo",
			},
		},
		{
			Case: "Template needs",
			Segment: &Segment{
				Template: "{{ .Segments.Git.URL }}",
			},
			Needs: []string{"Git"},
		},
		{
			Case: "Template & Foreground needs",
			Segment: &Segment{
				Template:            "{{ .Segments.Git.URL }}",
				ForegroundTemplates: []string{"foo", "{{ .Segments.Os.Icon }}"},
			},
			Needs: []string{"Git", "Os"},
		},
		{
			Case: "Template & Foreground & Background needs",
			Segment: &Segment{
				Template:            "{{ .Segments.Git.URL }}",
				ForegroundTemplates: []string{"foo", "{{ .Segments.Os.Icon }}"},
				BackgroundTemplates: []string{"bar", "{{ .Segments.Exit.Icon }}"},
			},
			Needs: []string{"Git", "Os", "Exit"},
		},
	}
	for _, tc := range cases {
		tc.Segment.evaluateNeeds()
		assert.Equal(t, tc.Needs, tc.Segment.Needs, tc.Case)
	}
}
