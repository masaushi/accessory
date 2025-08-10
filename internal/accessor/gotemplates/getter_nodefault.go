package templates

var GetterNoDefault = `
func ({{.Receiver}} {{.Struct}}) {{.GetterMethod}}() {{.Type}} {
  {{- if .Lock }}
  {{- if eq .LockType "rwmutex" }}
  {{.Receiver}}.{{.Lock}}.RLock()
	defer {{.Receiver}}.{{.Lock}}.RUnlock()
  {{- else }}
  {{.Receiver}}.{{.Lock}}.Lock()
	defer {{.Receiver}}.{{.Lock}}.Unlock()
  {{- end }}
  {{- end }}
  return {{.Receiver}}.{{.Field}}
}`
