package contracts

import _ "embed"

//go:embed connector.schema.json
var ConnectorSchema string

//go:embed github-public/parameters.schema.json
var GitHubParameters string
