package cachet

import "os"

// apiUrl -> https://demo.cachethq.io/api
// apiToken -> qwertyuiop
var apiUrl = os.Getenv("CACHET_API")
var apiToken = os.Getenv("CACHET_TOKEN")