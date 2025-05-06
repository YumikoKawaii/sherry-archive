package constants

const (
	MultimediaCompressionTopic = "sherry.archive.multimedia.compression"
	MultimediaTopic            = "sherry.archive.multimedia"
	FileNameDelimiter          = "."
)

var SupportedMultimediaTypes = map[string]bool{
	"jpg": true,
	"png": true,
}
