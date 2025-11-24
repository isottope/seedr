package seedr

const (
	BaseAPIURL       = "https://www.seedr.cc/api"
	OAuthURL         = "https://www.seedr.cc/oauth_test"
	ResourceURL      = OAuthURL + "/resource.php"
	TokenURL         = OAuthURL + "/token.php"
	DeviceCodeURL    = BaseAPIURL + "/device/code"
	DeviceAuthorizeURL = BaseAPIURL + "/device/authorize"

	DeviceClientID = "seedr_xbmc"
	PswrdClientID  = "seedr_chrome"
)
