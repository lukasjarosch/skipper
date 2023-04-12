module github.com/lukasjarosch/skipper/examples/dev

go 1.20

replace github.com/lukasjarosch/skipper => ../../

require github.com/lukasjarosch/skipper v0.0.0-00010101000000-000000000000

require (
	github.com/spf13/afero v1.9.5 // direct
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
