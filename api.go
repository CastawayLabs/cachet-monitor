package cachet

type CachetAPI struct {
	Url      string `json:"api_url"`
	Token    string `json:"api_token"`
	Insecure bool   `json:"insecure"`
}
