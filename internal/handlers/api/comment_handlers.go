package api

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/req"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/locales"
	"bbs-go/internal/spam"
	"strconv"

	"github.com/gin-gonic/gin"

	"bbs-go/internal/pkg/ginx"
	"bbs-go/internal/pkg/params"

	"bbs-go/internal/handlers/render"
	"bbs-go/internal/services"
	"github.com/mlogclub/simple/web"
)

func CommentComments(ctx *gin.Context) {
	var (
		cursor, _     = params.GetInt64(ctx, "cursor")
		entityType, _ = params.Get(ctx, "entityType")
		entityId      = common.GetID(ctx, "entityId")
		currentUser   = common.GetCurrentUser(ctx)
	)
	if entityType == constants.EntityTopic {
		topic := services.TopicService.Get(entityId)
		if !services.CategoryService.CanViewTopic(currentUser, topic) {
			ginx.WriteJSON(ctx, ginx.ErrorCode(403, locales.Get("topic.no_permission")))
			return
		}
	}
	comments, cursor, hasMore := services.CommentService.GetComments(entityType, entityId, cursor)
	ginx.WriteJSON(ctx, ginx.CursorData(render.BuildComments(comments, currentUser, true, false), strconv.FormatInt(cursor, 10), hasMore))

}

func CommentReplies(ctx *gin.Context) {
	var (
		cursor, _    = params.GetInt64(ctx, "cursor")
		commentId, _ = params.GetInt64(ctx, "commentId")
	)
	currentUser := common.GetCurrentUser(ctx)
	if !canViewCommentThread(currentUser, commentId) {
		ginx.WriteJSON(ctx, ginx.ErrorCode(403, locales.Get("topic.no_permission")))
		return
	}
	comments, cursor, hasMore := services.CommentService.GetReplies(commentId, cursor, 10)
	ginx.WriteJSON(ctx, ginx.CursorData(render.BuildComments(comments, currentUser, false, true), strconv.FormatInt(cursor, 10), hasMore))

}

func CommentCreate(ctx *gin.Context) {
	user := common.GetCurrentUser(ctx)
	if err := services.UserService.CheckPostStatus(user); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	var body req.CreateCommentReq
	if err := ginx.Bind(ctx, &body); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	body.UserAgent = web.GetUserAgent(ctx.Request)
	body.Ip = web.GetRequestIP(ctx.Request)
	if err := spam.CheckComment(user, body); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if !canCommentOnEntity(user, body) {
		ginx.WriteJSON(ctx, ginx.ErrorCode(403, locales.Get("topic.no_permission")))
		return
	}

	comment, err := services.CommentService.Publish(user.Id, body)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}

	ginx.WriteJSON(ctx, render.BuildComment(comment))

}

func canCommentOnEntity(user *models.User, body req.CreateCommentReq) bool {
	entityId := body.DecodedEntityId()
	switch body.EntityType {
	case constants.EntityTopic:
		return services.CategoryService.CanViewTopic(user, services.TopicService.Get(entityId))
	case constants.EntityComment:
		return canViewCommentThread(user, entityId)
	default:
		return true
	}
}

func canViewCommentThread(user *models.User, commentId int64) bool {
	for i := 0; i < 20 && commentId > 0; i++ {
		comment := services.CommentService.Get(commentId)
		if comment == nil || comment.Status != constants.StatusOk {
			return false
		}
		switch comment.EntityType {
		case constants.EntityTopic:
			return services.CategoryService.CanViewTopic(user, services.TopicService.Get(comment.EntityId))
		case constants.EntityComment:
			commentId = comment.EntityId
		default:
			return true
		}
	}
	return false
}

func CommentRemove(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	user := common.GetCurrentUser(ctx)
	if err := services.CommentService.DeleteByUser(user, id); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, nil)

}
