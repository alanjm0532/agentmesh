package protocol

import (
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MessagesPath       = "/messages"
	UnreadMessagesPath = "/messages/unread"
	ReadMessagesPath   = "/messages/read"
	MessageHistoryPath = "/messages/history"
	MaxMessageBytes    = 128 * 1024
	MaxMessagePageSize = 200
)

type SendMessageRequest struct {
	// Recipient is the destination username; sending creates the private conversation if needed.
	Recipient string `json:"recipient" binding:"required,notblank,max=255,utf8text"`
	// ClientMessageID is a client-generated UUID reused across retries to prevent duplicate sends.
	ClientMessageID string `json:"client_message_id"`
	// EncryptionAlgorithm identifies the body encryption scheme; only none is currently accepted.
	EncryptionAlgorithm EncryptionAlgorithm `json:"encryption_algorithm"`
	// ContentType describes the body format; currently only text/plain is accepted.
	ContentType string `json:"content_type"`
	// Body is 1-131072 UTF-8 bytes without NUL; whitespace and newlines are preserved.
	Body string `json:"body"`
}

// ValidMessageBody uses bytes, not runes. PostgreSQL TEXT cannot store NUL.
func ValidMessageBody(body string) bool {
	return len(body) > 0 && len(body) <= MaxMessageBytes && utf8.ValidString(body) && !strings.ContainsRune(body, 0)
}

type Message struct {
	// ID is the server-generated UUID used to identify and deduplicate received messages.
	ID string `json:"id"`
	// ConversationID is the UUID of the private conversation containing this message.
	ConversationID string `json:"conversation_id"`
	// Seq is the positive, monotonically increasing message position within the conversation.
	Seq int64 `json:"seq"`
	// SenderID is the sender's stable user UUID, determined from the authenticated token.
	SenderID string `json:"sender_id"`
	// SenderUsername is the sender's account name for display.
	SenderUsername string `json:"sender_username"`
	// ClientMessageID is the original client-supplied send idempotency key.
	ClientMessageID string `json:"client_message_id"`
	// EncryptionAlgorithm identifies how the body is encrypted; none means plaintext.
	EncryptionAlgorithm EncryptionAlgorithm `json:"encryption_algorithm"`
	// ContentType describes the plaintext content format, currently text/plain.
	ContentType string `json:"content_type"`
	// Body contains the message content; with none it is unencrypted UTF-8 text.
	Body string `json:"body"`
	// CreatedAt is the server-assigned message creation time in UTC.
	CreatedAt time.Time `json:"created_at"`
}

type ReadPosition struct {
	// ConversationID is the UUID of the conversation whose read cursor is being advanced.
	ConversationID string `json:"conversation_id"`
	// LastReadSeq marks all positions through this sequence as read; zero means no progress.
	LastReadSeq int64 `json:"last_read_seq"`
}

type ReadMessagesRequest struct {
	// Reads contains 1-200 account read positions, applied atomically and never moved backwards.
	Reads []ReadPosition `json:"reads"`
}

type UnreadMessagesQuery struct {
	// From optionally restricts unread messages to one peer's username.
	From string `form:"from" binding:"omitempty,notblank,max=255,utf8text"`
	// Limit bounds the batch size to 1-200 messages, defaulting to 100.
	Limit int `form:"limit,default=100" binding:"min=1,max=200"`
}

type MessageHistoryQuery struct {
	// Peer is the username whose private conversation is being queried.
	Peer string `form:"peer" binding:"required,notblank,max=255,utf8text"`
	// Before is the optional opaque cursor for an older history page.
	Before string `form:"before"`
	// Limit bounds the page size to 1-200 messages, defaulting to 50.
	Limit int `form:"limit,default=50" binding:"min=1,max=200"`
}

type UnreadMessagesResponse struct {
	// Messages contains the unread batch, ordered by sequence within each conversation.
	Messages []Message `json:"messages"`
	// ReadPositions may be acknowledged only after this batch has been successfully output.
	ReadPositions []ReadPosition `json:"read_positions"`
	// HasMore indicates that additional unread messages remain beyond this batch.
	HasMore bool `json:"has_more"`
}

type MessageHistoryResponse struct {
	// Messages contains the history page in descending conversation-sequence order.
	Messages []Message `json:"messages"`
	// NextCursor is the opaque before cursor for the next older page; omitted on the last page.
	NextCursor string `json:"next_cursor,omitempty"`
	// HasMore indicates whether an older history page is available.
	HasMore bool `json:"has_more"`
}
