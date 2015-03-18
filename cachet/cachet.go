package cachet

import "os"

var ApiUrl = os.Getenv("CACHET_API")
var ApiToken = os.Getenv("CACHET_TOKEN")