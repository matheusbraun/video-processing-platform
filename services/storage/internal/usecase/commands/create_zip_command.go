package commands

type CreateZipCommand struct {
	VideoID   string
	S3Prefix  string
	OutputKey string
}
