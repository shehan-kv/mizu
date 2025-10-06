package service

import (
	"context"
	"encoding/json"
	"errors"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
	"mizu/internal/dto/common"
	dto "mizu/internal/dto/message"
	"mizu/internal/event"
	"mizu/internal/logger"
	"mizu/internal/middleware"
)

// Handles message related operations.
type MessageService struct {
	lg      logger.Logger
	evtSndr *event.EventSender
	msgSt   store.MessageStore
}

// Creates a new instance of MessageService.
// It takes a logger, a MessageStore for message-related data operations.
//
// Parameters:
//   - lg: logger that implements the logger.Logger interface
//   - msgSt: project store that implements the MessageStore interface
//
// Returns:
//   - a pointer to a new MessageService
func NewMessageService(lg logger.Logger, evtSndr *event.EventSender, msgSt store.MessageStore) *MessageService {
	return &MessageService{
		lg:      lg,
		evtSndr: evtSndr,
		msgSt:   msgSt,
	}
}

// CreateMessage creates a new message in a specified channel and
// returns a pointer to a dto.MessageResponse struct.
// Checks if the sending user has access to the specified channel.
// Sends message to all other users in the specified channel
// using event.EventSender.
//
// If the sending user doesn't have access to the channel, it returns service.ErrUnauthorized.
// If a required field is missing, it returns service.ErrBadRequest.
// if a foreign key violation occurs, it returns service.ErrBadRequest.
// If any internal errors occur, it returns service.ErrInternalError.
func (msgSrv *MessageService) CreateMessage(
	ctx context.Context,
	channelId int64,
	request *dto.MessageCreateRequest) (*dto.MessageResponse, error) {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		msgSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"channel_id", channelId,
			"err", err)
		return nil, ErrInternalError
	}

	channels, err := msgSrv.msgSt.GetChannelIdsByUserId(ctx, actor.Id)
	if err != nil {
		msgSrv.lg.Error("could not get channels",
			"event", event.EventGetFailed,
			"correlation_id", correlationId,
			"scope", "user_service",
			"channel_id", channelId,
			"actor_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	users, err := msgSrv.msgSt.GetUsersByChannelId(ctx, channelId)
	if err != nil {
		msgSrv.lg.Error("could not get users by channel",
			"event", event.EventGetFailed,
			"correlation_id", correlationId,
			"scope", "user_service",
			"channel_id", channelId,
			"actor_id", actor.Id,
			"err", err)
		return nil, ErrInternalError
	}

	// removing sender id from users list
	// to send message created event to
	for i, id := range users {
		if id == actor.Id {
			users = append(users[:i], users[i+1:]...)
			break
		}
	}

	isChannelValid := false
	for _, channel := range channels {
		if channel == channelId {
			isChannelValid = true
		}
	}

	if !isChannelValid {
		msgSrv.lg.Error("actor does not have access to channel",
			"event", event.EventUserUnauthorized,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"channel_id", channelId,
			"err", err)
		return nil, ErrUnauthorized
	}

	message, err := msgSrv.msgSt.CreateOne(ctx, &params.MessageCreate{
		ChannelId: channelId,
		UserId:    actor.Id,
		Type:      params.MessageTypeUser,
		Message:   request.Message,
	})
	if err != nil {
		if errors.Is(err, store.ErrForeignKeyViolation) {
			msgSrv.lg.Error("message foreign key constraint violated",
				"event", event.EventCreateFailed,
				"correlation_id", correlationId,
				"scope", "user_service",
				"actor_id", actor.Id,
				"channel_id", channelId,
				"err", err)
			return nil, ErrBadRequest
		}

		if errors.Is(err, store.ErrNotNullViolation) {
			msgSrv.lg.Error("message required field missing",
				"event", event.EventCreateFailed,
				"correlation_id", correlationId,
				"scope", "user_service",
				"actor_id", actor.Id,
				"channel_id", channelId,
				"err", err)
			return nil, ErrBadRequest
		}

		msgSrv.lg.Error("could not create message",
			"event", event.EventCreateFailed,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"channel_id", channelId,
			"err", err)
		return nil, ErrInternalError
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		msgSrv.lg.Error("could not marshal message into json",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"channel_id", channelId,
			"err", err)
	}

	msgSrv.evtSndr.SendTo(string(event.EventMessage), msgBytes, users)

	resp := dto.MessageResponse{
		UserId:    message.UserId,
		MessageId: message.MessageId,
		Message:   message.Message,
		FirstName: message.FirstName,
		LastName:  message.LastName,
		Role:      message.Role,
		Title:     message.Title,
		Image:     message.Image,
		CreatedAt: message.CreatedAt,
	}

	return &resp, nil
}

// GetMessages returns a paginated list of MessageResponse DTOs.
// Checks if the sending user has access to the specified channel.
//
// If the requesting user doesn't have access to the channel, it returns service.ErrUnauthorized.
// If any internal errors occur, it returns service.ErrInternalError.
func (msgSrv *MessageService) GetMessages(
	ctx context.Context,
	channelId int64,
	query *dto.MessageSearchQuery) (*common.Page[[]dto.MessageResponse], error) {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		msgSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "message_service",
			"channel_id", channelId,
			"page", query.Page,
			"limit", query.Limit,
			"err", err)
		return nil, ErrInternalError
	}

	channels, err := msgSrv.msgSt.GetChannelIdsByUserId(ctx, actor.Id)
	if err != nil {
		msgSrv.lg.Error("could not get channels",
			"event", event.EventGetFailed,
			"correlation_id", correlationId,
			"scope", "user_service",
			"channel_id", channelId,
			"actor_id", actor.Id,
			"page", query.Page,
			"limit", query.Limit,
			"err", err)
		return nil, ErrInternalError
	}

	isChannelValid := false
	for _, channel := range channels {
		if channel == channelId {
			isChannelValid = true
		}
	}

	if !isChannelValid {
		msgSrv.lg.Error("actor does not have access to channel",
			"event", event.EventUserUnauthorized,
			"correlation_id", correlationId,
			"scope", "user_service",
			"actor_id", actor.Id,
			"channel_id", channelId,
			"page", query.Page,
			"limit", query.Limit,
			"err", err)
		return nil, ErrUnauthorized
	}

	msgList, err := msgSrv.msgSt.GetByChannelId(ctx, channelId, &params.MessageSearch{
		Offset: (query.Page - 1) * query.Limit,
		Limit:  query.Limit,
	})
	if err != nil {
		msgSrv.lg.Error("could not get messages",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "message_service",
			"channel_id", channelId,
			"actor_id", actor.Id,
			"page", query.Page,
			"limit", query.Limit,
			"err", err)
		return nil, ErrInternalError
	}

	count, err := msgSrv.msgSt.CountByChannelId(ctx, channelId)
	if err != nil {
		msgSrv.lg.Error("could not get messages count",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "message_service",
			"channel_id", channelId,
			"actor_id", actor.Id,
			"page", query.Page,
			"limit", query.Limit,
			"err", err)
		return nil, ErrInternalError
	}

	msgResponses := make([]dto.MessageResponse, len(msgList))
	for i, msg := range msgList {
		resp := dto.MessageResponse{
			UserId:    msg.UserId,
			MessageId: msg.MessageId,
			FirstName: msg.FirstName,
			LastName:  msg.LastName,
			Role:      msg.Role,
			Type:      msg.Type,
			Title:     msg.Title,
			Image:     msg.Image,
			CreatedAt: msg.CreatedAt,
			Message:   msg.Message,
		}
		msgResponses[i] = resp
	}

	resp := common.Page[[]dto.MessageResponse]{
		Count: count,
		Limit: query.Limit,
		Page:  query.Page,
		Data:  msgResponses,
	}

	return &resp, nil
}

func (msgSrv *MessageService) GetChannels(ctx context.Context) ([]dto.ChannelResponse, error) {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		msgSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "message_service",
			"err", err)
		return nil, ErrInternalError
	}

	channels, err := msgSrv.msgSt.GetChannelsByUserId(ctx, actor.Id)
	if err != nil {
		return nil, ErrInternalError
	}

	resp := make([]dto.ChannelResponse, len(channels))

	for i, channel := range channels {
		resp[i] = dto.ChannelResponse{
			Id:        channel.Id,
			ProjectId: channel.ProjectId,
			Name:      channel.Name,
			CreatedAt: channel.CreatedAt,
		}
	}

	return resp, nil
}
