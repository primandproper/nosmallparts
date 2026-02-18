package textsearch

type IndexRequest struct {
	RequestID string `json:"id"`
	RowID     string `json:"rowID"`
	IndexType string `json:"type"`
	Delete    bool   `json:"delete"`
}
