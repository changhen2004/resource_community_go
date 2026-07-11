package media

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type responseEnvelope struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type errorData struct {
	ErrorCode string `json:"errorCode"`
}

func writeSuccess(ctx *gin.Context, status int, data interface{}) {
	ctx.JSON(status, responseEnvelope{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func writeError(ctx *gin.Context, status int, code int, message, errorCode string) {
	ctx.JSON(status, responseEnvelope{
		Code:    code,
		Message: message,
		Data: errorData{
			ErrorCode: errorCode,
		},
	})
}

func (h *Handler) UploadCover(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		writeError(ctx, http.StatusBadRequest, 10001, ErrFileRequired.Error(), "INVALID_REQUEST")
		return
	}

	url, err := h.service.SaveCover(fileHeader)
	if err != nil {
		h.writeUploadError(ctx, err)
		return
	}

	writeSuccess(ctx, http.StatusCreated, UploadSingleResponse{URL: url})
}

func (h *Handler) UploadContentImages(ctx *gin.Context) {
	form, err := ctx.MultipartForm()
	if err != nil {
		writeError(ctx, http.StatusBadRequest, 10001, ErrFileRequired.Error(), "INVALID_REQUEST")
		return
	}

	files := form.File["files"]
	urls, err := h.service.SaveContentImages(files)
	if err != nil {
		h.writeUploadError(ctx, err)
		return
	}

	writeSuccess(ctx, http.StatusCreated, UploadMultipleResponse{URLs: urls})
}

func (h *Handler) writeUploadError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrFileRequired):
		writeError(ctx, http.StatusBadRequest, 10001, err.Error(), "INVALID_REQUEST")
	case errors.Is(err, ErrTooManyFiles):
		writeError(ctx, http.StatusBadRequest, 10013, err.Error(), "TOO_MANY_FILES")
	case errors.Is(err, ErrFileTooLarge):
		writeError(ctx, http.StatusBadRequest, 10014, err.Error(), "FILE_TOO_LARGE")
	case errors.Is(err, ErrUnsupportedType):
		writeError(ctx, http.StatusBadRequest, 10015, err.Error(), "UNSUPPORTED_FILE_TYPE")
	default:
		writeError(ctx, http.StatusInternalServerError, 10005, err.Error(), "INTERNAL_ERROR")
	}
}
