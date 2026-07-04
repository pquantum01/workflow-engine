package dto

type UploadAttachmentRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Data        string `json:"data"` // base64-encoded bytes
}

type AttachmentMetaResponse struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SizeBytes   int    `json:"size_bytes"`
	UploadedAt  string `json:"uploaded_at"`
	UpdatedAt   string `json:"updated_at"`
}
