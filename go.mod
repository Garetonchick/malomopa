module malomopa

go 1.23

require (
	github.com/gocql/gocql v1.7.0
	golang.org/x/sync v0.8.0
)

require go.uber.org/multierr v1.10.0 // indirect

require (
	github.com/go-chi/chi/v5 v5.1.0
	github.com/golang/snappy v0.0.3 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	go.uber.org/zap v1.27.0
	gopkg.in/inf.v0 v0.9.1 // indirect
)

replace github.com/gocql/gocql => github.com/scylladb/gocql v1.14.4
