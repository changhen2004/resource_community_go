package media

type UploadSingleResponse struct {
	URL string `json:"url"`
}

type UploadMultipleResponse struct {
	URLs []string `json:"urls"`
}
