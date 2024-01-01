package drive

type Credentials struct {
	Username        string
	Password        string
	TwoFA           string
	MailboxPassword string
}

type Tokens struct {
	UID           string
	AccessToken   string
	RefreshToken  string
	SaltedKeyPass string
}
