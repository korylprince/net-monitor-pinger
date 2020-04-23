package main

type config struct {
	DNSWorkers        int    `required:"true" default:"8"`
	DNSLookupInterval int    `required:"true" default:"30"` // in minutes
	PingWorkers       int    `required:"true" default:"16"`
	PingBufferSize    int    `required:"true" default:"1024"`
	PingInterval      int    `required:"true" default:"5"`    // in seconds
	PingTimeout       int    `required:"true" default:"1000"` // in milliseconds
	PurgeInterval     int    `required:"true" default:"60"`   // in minutes
	PurgeOlderThan    int    `required:"true" default:"1440"` // in minutes
	GraphQLEndpoint   string `required:"true"`
	GraphQLAPISecret  string `required:"true"`
}
