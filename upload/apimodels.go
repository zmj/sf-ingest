package upload

type folder struct {
	Name string
	ID   string `json:"Id"`
}

type uploadSpecReq struct {
	Method   string
	Raw      bool
	FileName string
	FileSize uint64
}

type uploadSpec struct {
	ChunkURI string `json:"ChunkUri"`
}

type uploadResult struct {
	Error        bool   `json:"error"`
	ErrorMessage string `json:"errormessage"`
	Value        []file `json:"value"`
}

type file struct {
	ID  string `json:"id"`
	Md5 string `json:"md5"`
}
