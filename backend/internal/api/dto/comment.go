package dto

type AddCommentRequest struct {
	Author  string `json:"author"`
	Content string `json:"content"`
}

type CommentResponse struct {
	ID             string `json:"id"`
	TaskInstanceID string `json:"task_instance_id"`
	Author         string `json:"author"`
	Content        string `json:"content"`
	CreatedAt      string `json:"created_at"`
}
