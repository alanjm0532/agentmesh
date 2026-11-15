package client

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/alanjm0532/agentmesh/internal/protocol"
)

func (c *Client) SendMessage(ctx context.Context, input protocol.SendMessageRequest) (protocol.Message, error) {
	return request[protocol.Message](ctx, c, http.MethodPost, protocol.MessagesPath, input)
}

func (c *Client) UnreadMessages(ctx context.Context, from string, limit int) (protocol.UnreadMessagesResponse, error) {
	query := url.Values{"limit": {strconv.Itoa(limit)}}
	if from != "" {
		query.Set("from", from)
	}
	return doRequest[protocol.UnreadMessagesResponse](ctx, c, http.MethodGet, protocol.UnreadMessagesPath+"?"+query.Encode())
}

func (c *Client) MessageHistory(ctx context.Context, peer, before string, limit int) (protocol.MessageHistoryResponse, error) {
	query := url.Values{"peer": {peer}, "limit": {strconv.Itoa(limit)}}
	if before != "" {
		query.Set("before", before)
	}
	return doRequest[protocol.MessageHistoryResponse](ctx, c, http.MethodGet, protocol.MessageHistoryPath+"?"+query.Encode())
}

func (c *Client) ReadMessages(ctx context.Context, positions []protocol.ReadPosition) error {
	_, err := request[map[string]string](ctx, c, http.MethodPost, protocol.ReadMessagesPath, protocol.ReadMessagesRequest{Reads: positions})
	return err
}
