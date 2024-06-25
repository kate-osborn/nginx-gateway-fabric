package clientsettings

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/http"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/policies"
)

var (
	tmpl                 = template.Must(template.New("client settings policy").Parse(clientSettingsTemplate))
	tmplExternalRedirect = template.Must(
		template.New("client settings policy ext redirect").Parse(externalRedirectTemplate),
	)
)

const clientSettingsTemplate = `
{{- if .Body }}
	{{- if .Body.MaxSize }}
client_max_body_size {{ .Body.MaxSize }};
	{{- end }}
	{{- if .Body.Timeout }}
client_body_timeout {{ .Body.Timeout }};
	{{- end }}
{{- end }}
{{- if .KeepAlive }}
	{{- if .KeepAlive.Requests }}
keepalive_requests {{ .KeepAlive.Requests }};
	{{- end }}
	{{- if .KeepAlive.Time }}
keepalive_time {{ .KeepAlive.Time }};
	{{- end }}
    {{- if .KeepAlive.Timeout }}
        {{- if and .KeepAlive.Timeout.Server .KeepAlive.Timeout.Header }}
keepalive_timeout {{ .KeepAlive.Timeout.Server }} {{ .KeepAlive.Timeout.Header }};
        {{- else if .KeepAlive.Timeout.Server }}
keepalive_timeout {{ .KeepAlive.Timeout.Server }};
        {{- end }}
    {{- end }}
{{- end }}
`

const externalRedirectTemplate = `
client_max_body_size {{ . }};
`

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

// TODO: do I need the server here?
func (g Generator) GenerateForServer(pols []policies.Policy, _ http.Server) policies.GenerateResult {
	files := make([]policies.File, 0, len(pols))

	for _, pol := range pols {
		csp, ok := pol.(*ngfAPI.ClientSettingsPolicy)
		if !ok {
			continue
		}

		content := helpers.MustExecuteTemplate(tmpl, csp.Spec)
		// TODO: this check doesn't work
		// Find a way to eliminate empty files
		if len(content) == 0 {
			continue
		}

		files = append(files, policies.File{
			Name:    fmt.Sprintf("ClientSettingsPolicy_%s_%s_server.conf", csp.Namespace, csp.Name),
			Content: content,
		})
	}

	return policies.GenerateResult{Files: files}
}

func (g Generator) GenerateForLocation(pols []policies.Policy, location http.Location) policies.GenerateResult {
	if location.Type == http.ExternalLocationType {
		files := make([]policies.File, 0, len(pols))

		for _, pol := range pols {
			csp, ok := pol.(*ngfAPI.ClientSettingsPolicy)
			if !ok {
				continue
			}

			files = append(files, policies.File{
				Name:    fmt.Sprintf("ClientSettingsPolicy_%s_%s_ext.conf", csp.Namespace, csp.Name),
				Content: helpers.MustExecuteTemplate(tmpl, csp.Spec),
			})
		}

		return policies.GenerateResult{Files: files}
	}

	var maxBodySize ngfAPI.Size

	for _, pol := range pols {
		csp, ok := pol.(*ngfAPI.ClientSettingsPolicy)
		if !ok {
			continue
		}

		if csp.Spec.Body != nil && csp.Spec.Body.MaxSize != nil {
			maxBodySize = getMaxSize(maxBodySize, *csp.Spec.Body.MaxSize)
		}
	}

	if maxBodySize == "" {
		return policies.GenerateResult{}
	}

	return policies.GenerateResult{
		Files: []policies.File{
			{
				Name:    fmt.Sprintf("ClientSettingsPolicy_%s_redirect.conf", location.HTTPMatchKey),
				Content: helpers.MustExecuteTemplate(tmplExternalRedirect, maxBodySize),
			},
		},
	}
}

func (g Generator) GenerateForInternalLocation(
	pols []policies.Policy,
	_ http.Location,
) policies.GenerateResult {
	files := make([]policies.File, 0, len(pols))

	for _, pol := range pols {
		csp, ok := pol.(*ngfAPI.ClientSettingsPolicy)
		if !ok {
			continue
		}

		files = append(files, policies.File{
			Name:    fmt.Sprintf("ClientSettingsPolicy_%s_%s_int.conf", csp.Namespace, csp.Name),
			Content: helpers.MustExecuteTemplate(tmpl, csp.Spec),
		})
	}

	return policies.GenerateResult{Files: files}
}

func getMaxSize(s1 ngfAPI.Size, s2 ngfAPI.Size) ngfAPI.Size {
	if s1 == "" {
		return s2
	}

	if s2 == "" {
		return s1
	}

	s1Bytes, err := parseSizeToBytes(s1)
	if err != nil {
		panic(err)
	}

	s2Bytes, err := parseSizeToBytes(s2)
	if err != nil {
		panic(err)
	}

	if s1Bytes > s2Bytes {
		return s1
	}

	return s2
}

// sizeMultipliers defines the conversion rates for each unit to bytes.
var sizeMultipliers = map[string]int64{
	"b": 1,                  // bytes
	"k": 1024,               // kilobytes
	"m": 1024 * 1024,        // megabytes
	"g": 1024 * 1024 * 1024, // gigabytes
}

// parseSizeToBytes parses the size string and returns the size in bytes.
func parseSizeToBytes(s ngfAPI.Size) (int64, error) {
	re := regexp.MustCompile(`^(\d{1,4})(k|m|g)?$`)
	matches := re.FindStringSubmatch(string(s))
	if len(matches) < 3 {
		return 0, fmt.Errorf("invalid size format, could not find submatches: %s", s)
	}

	value, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size format, could not parse int: %s", s)
	}

	unit := strings.ToLower(matches[2])
	if unit == "" {
		unit = "b" // Default to bytes if no unit is specified
	}

	return value * sizeMultipliers[unit], nil
}
