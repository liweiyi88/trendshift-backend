package trending

type Tag struct {
	Id   int
	Name string
}

type CreateTagRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateTagRequest struct {
	Id   int    `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}
