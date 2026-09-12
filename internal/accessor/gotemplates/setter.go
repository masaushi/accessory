package templates

var Setter = `
func ({{.Receiver}} *{{.Struct}}) {{.SetterMethod}}(val {{.Type}}){{- if .ReturnReceiver -}} (* {{.Struct}}) {{- end -}}{
  if {{.Receiver}} == nil {
    return{{if .ReturnReceiver }} nil{{end}}
  }
  {{- if ne .Lock "" }}
  {{.Receiver}}.{{.Lock}}.Lock()
  defer {{.Receiver}}.{{.Lock}}.Unlock()
  {{- end }}
  {{.Receiver}}.{{.Field}} = val
  {{- if .ReturnReceiver }}
  return {{.Receiver}}
  {{- end}}
}
`
