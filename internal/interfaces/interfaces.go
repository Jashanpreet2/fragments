package interfaces

type IFragment interface {
	GetJson() (string, bool)
	GetData() ([]byte, bool)
	SetData(data []byte) bool
	Save() bool
	MimeType() string
	ConvertMimetype(ext string) ([]byte, string, error)
	Formats() []string
}
