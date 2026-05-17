package handler

import (
	"video-recommend/internal/domain"
	"video-recommend/internal/middleware"
	"video-recommend/internal/service"
	"video-recommend/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userSvc *service.UserService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		userSvc: service.NewUserService(),
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req domain.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.userSvc.Register(&req)
	if err != nil {
		if err == service.ErrUserAlreadyExists {
			response.BadRequest(c, "username already exists")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}

	response.Created(c, user)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req domain.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.userSvc.Login(&req)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			response.Unauthorized(c, "invalid username or password")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, resp)
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	user, err := h.userSvc.GetByID(userID)
	if err != nil {
		if err == service.ErrUserNotFound {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, user)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req domain.UserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.userSvc.UpdateProfile(userID, &req)
	if err != nil {
		if err == service.ErrUserNotFound {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, user)
}

func (h *UserHandler) List(c *gin.Context) {
	var req struct {
		Page     int `form:"page" binding:"min=1"`
		PageSize int `form:"page_size" binding:"min=1,max=100"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}

	users, total, err := h.userSvc.List(req.Page, req.PageSize)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.PageSuccess(c, total, req.Page, req.PageSize, users)
}
