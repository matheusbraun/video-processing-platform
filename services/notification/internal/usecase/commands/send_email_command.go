package commands

type SendEmailCommand struct {
	UserID       int64
	VideoID      string
	UserEmail    string
	Status       string
	FrameCount   int
	ErrorMessage string
}
