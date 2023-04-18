module github.com/lukasjarosch/skipper/examples/dev

go 1.20

replace github.com/lukasjarosch/skipper => ../../

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/dominikbraun/graph v0.17.0
	github.com/lukasjarosch/skipper v0.0.0-00010101000000-000000000000
)

require golang.org/x/exp v0.0.0-20230321023759-10a507213a29 // indirect

require (
	github.com/spf13/afero v1.9.5 // direct
	golang.org/x/text v0.8.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
