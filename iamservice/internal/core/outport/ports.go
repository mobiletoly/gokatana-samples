package outport

//go:generate go tool gobetter -input $GOFILE

type Ports struct { //+gob:Constructor
	AuthUserPersist    AuthUserPersist
	UserProfilePersist UserProfilePersist
	Tx                 TxPort
	Mailer             Mailer
}
