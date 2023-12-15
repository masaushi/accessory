package templates

var GetterNoDefault = `
func ({{.Receiver}} {{.Struct}}) {{.GetterMethod}}() {{.Type}} {
  {{- if .Lock }}
  {{.Receiver}}.{{.Lock}}.Lock()
	defer {{.Receiver}}.{{.Lock}}.Unlock()
  {{- end }}
  return {{.Receiver}}.{{.Field}}
}`
