package config

const CensorUsernameKey = "censor_username"

var defaultCensorWords = []string{
	"admin",
	"administrator",
	"root",
	"system",
	"moderator",
	"support",
	"official",
	"notakto",
	"null",
	"undefined",
	"asshole",
	"bastard",
	"bitch",
	"blowjob",
	"bullshit",
	"cock",
	"crap",
	"cunt",
	"dick",
	"douche",
	"fag",
	"faggot",
	"fuck",
	"motherfucker",
	"nazi",
	"nigger",
	"nigga",
	"penis",
	"piss",
	"pussy",
	"shit",
	"slut",
	"twat",
	"vagina",
	"whore",
}

func DefaultCensorWords() []string {
	return append([]string(nil), defaultCensorWords...)
}
