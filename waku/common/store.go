package common

type StoreQueryRequest struct {
	RequestId         string        `json:"requestId,omitempty"`
	IncludeData       bool          `json:"includeData,omitempty"`
	PubsubTopic       *string       `json:"pubsubTopic,omitempty"`
	ContentTopics     []string      `json:"contentTopics,omitempty"`
	TimeStart         *int64        `json:"timeStart,omitempty"`
	TimeEnd           *int64        `json:"timeEnd,omitempty"`
	MessageHashes     []MessageHash `json:"messageHashes,omitempty"`
	PaginationCursor  []MessageHash `json:"paginationCursor,omitempty"`
	PaginationForward bool          `json:"paginationForward,omitempty"`
	PaginationLimit   *uint64       `json:"paginationLimit,omitempty"`
}

type StoreQueryResponse struct {
	RequestId        string        `json:"request_id,omitempty"`
	StatusCode       *uint32       `json:"status_code,omitempty"`
	StatusDesc       *string       `json:"status_desc,omitempty"`
	Messages         []*Envelope   `json:"messages,omitempty"`
	PaginationCursor []MessageHash `json:"pagination_cursor,omitempty"`
}
