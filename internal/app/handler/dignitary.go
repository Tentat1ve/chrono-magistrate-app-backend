package handler

import (
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"awesomeProject/internal/app/ds"
	"awesomeProject/internal/app/repository"
	"awesomeProject/internal/app/serializer"

	"github.com/gin-gonic/gin"
)

const (
	maxImageSize = 10 << 20 // 10 МБ
	maxVideoSize = 50 << 20 // 50 МБ
)

// GetDignitaries — GET /api/dignitaries?office_year=1570
// Список опубликованных сановников с признаком is_creator.
func (h *Handler) GetDignitaries(ctx *gin.Context) {
	user, err := h.CurrentUser()
	if err != nil {
		h.internalError(ctx, err)
		return
	}

	var year *int
	if yearQuery := ctx.Query("office_year"); yearQuery != "" {
		y, err := strconv.Atoi(yearQuery)
		if err != nil {
			h.errorJSON(ctx, http.StatusBadRequest, "office_year должен быть целым числом (до н.э. — отрицательным)")
			return
		}
		year = &y
	}

	dignitaries, err := h.Repository.GetDignitaries(year)
	if err != nil {
		h.internalError(ctx, err)
		return
	}

	result := make([]serializer.DignitaryListItemJSON, 0, len(dignitaries))
	for _, d := range dignitaries {
		result = append(result, serializer.DignitaryToListItem(d, user.ID))
	}
	ctx.JSON(http.StatusOK, result)
}

// GetDignitaryFeed — GET /api/dignitaries/feed, /api/dignitaries/feed/:id, /api/dignitaries/feed/:id?next=true
// Один опубликованный сановник с признаком is_liked.
func (h *Handler) GetDignitaryFeed(ctx *gin.Context) {
	user, err := h.CurrentUser()
	if err != nil {
		h.internalError(ctx, err)
		return
	}

	var dignitary ds.Dignitary
	if idStr := ctx.Param("id"); idStr == "" {
		dignitary, err = h.Repository.GetNextDignitary(0)
	} else {
		id, ok := h.parseID(ctx)
		if !ok {
			return
		}
		if ctx.Query("next") == "true" {
			dignitary, err = h.Repository.GetNextDignitary(id)
		} else {
			dignitary, err = h.Repository.GetDignitary(id)
		}
	}

	if errors.Is(err, repository.ErrNotFound) {
		h.errorJSON(ctx, http.StatusNotFound, "сановник не найден")
		return
	}
	if err != nil {
		h.internalError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, serializer.DignitaryToFeed(dignitary, user.ID))
}

// GetDignitaryDraft — GET /api/dignitaries/draft — черновик текущего пользователя
func (h *Handler) GetDignitaryDraft(ctx *gin.Context) {
	user, err := h.CurrentUser()
	if err != nil {
		h.internalError(ctx, err)
		return
	}

	draft, err := h.Repository.GetDraftDignitary(user.ID)
	if errors.Is(err, repository.ErrNotFound) {
		h.errorJSON(ctx, http.StatusNotFound, "у пользователя нет черновика")
		return
	}
	if err != nil {
		h.internalError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, serializer.DignitaryToJSON(draft))
}

// CreateDignitary — POST /api/dignitaries (multipart/form-data: name, image, video)
// Создаёт черновик; файлы сохраняются в Minio под сгенерированными латинскими именами.
func (h *Handler) CreateDignitary(ctx *gin.Context) {
	user, err := h.CurrentUser()
	if err != nil {
		h.internalError(ctx, err)
		return
	}

	name := strings.TrimSpace(ctx.PostForm("name"))
	if name == "" || !utf8.ValidString(name) || utf8.RuneCountInString(name) > 255 {
		h.errorJSON(ctx, http.StatusBadRequest, "укажите name в UTF-8 (до 255 символов)")
		return
	}
	image, ok := h.formFile(ctx, "image", "image/", maxImageSize)
	if !ok {
		return
	}
	video, ok := h.formFile(ctx, "video", "video/", maxVideoSize)
	if !ok {
		return
	}

	if _, err := h.Repository.GetDraftDignitary(user.ID); err == nil {
		h.errorJSON(ctx, http.StatusConflict, "у пользователя уже есть черновик")
		return
	}

	imageURL, err := h.upload(ctx, "dignitary_image", image)
	if err != nil {
		h.internalError(ctx, err)
		return
	}
	videoURL, err := h.upload(ctx, "dignitary_video", video)
	if err != nil {
		h.Repository.RemoveFile(ctx, imageURL)
		h.internalError(ctx, err)
		return
	}

	dignitary := ds.Dignitary{Name: name, ImageURL: imageURL, VideoURL: videoURL, CreatorID: user.ID}
	if err := h.Repository.CreateDraftDignitary(&dignitary); err != nil {
		h.Repository.RemoveFile(ctx, imageURL)
		h.Repository.RemoveFile(ctx, videoURL)
		if errors.Is(err, repository.ErrAlreadyExists) {
			h.errorJSON(ctx, http.StatusConflict, "у пользователя уже есть черновик")
			return
		}
		h.internalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, serializer.DignitaryToJSON(dignitary))
}

// formFile достаёт файл из формы и проверяет тип и размер
func (h *Handler) formFile(ctx *gin.Context, field, typePrefix string, maxSize int64) (*multipart.FileHeader, bool) {
	file, err := ctx.FormFile(field)
	if err != nil {
		h.errorJSON(ctx, http.StatusBadRequest, "прикрепите файл "+field)
		return nil, false
	}
	if !strings.HasPrefix(file.Header.Get("Content-Type"), typePrefix) {
		h.errorJSON(ctx, http.StatusBadRequest, "файл "+field+" должен иметь тип "+typePrefix+"*")
		return nil, false
	}
	if file.Size > maxSize {
		h.errorJSON(ctx, http.StatusRequestEntityTooLarge, "файл "+field+" слишком большой")
		return nil, false
	}
	return file, true
}

func (h *Handler) upload(ctx *gin.Context, prefix string, file *multipart.FileHeader) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()
	return h.Repository.UploadFile(ctx, prefix, file.Filename, file.Header.Get("Content-Type"), f, file.Size)
}

// PublishDignitaryDraft — PUT /api/dignitaries/draft/publish
// Заполняет поля черновика и меняет статус на «опубликован».
func (h *Handler) PublishDignitaryDraft(ctx *gin.Context) {
	user, err := h.CurrentUser()
	if err != nil {
		h.internalError(ctx, err)
		return
	}

	var req serializer.PublishDignitaryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorJSON(ctx, http.StatusBadRequest, "заполните office, description, office_start и office_end")
		return
	}
	if *req.OfficeStart > *req.OfficeEnd {
		h.errorJSON(ctx, http.StatusBadRequest, "office_start не может быть больше office_end")
		return
	}

	dignitary, err := h.Repository.PublishDignitary(user.ID, strings.TrimSpace(req.Office),
		strings.TrimSpace(req.Description), *req.OfficeStart, *req.OfficeEnd)
	if errors.Is(err, repository.ErrNotFound) {
		h.errorJSON(ctx, http.StatusNotFound, "у пользователя нет черновика")
		return
	}
	if err != nil {
		h.internalError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, serializer.DignitaryToJSON(dignitary))
}

// DeleteDignitary — DELETE /api/dignitaries/:id — логическое удаление своего сановника
func (h *Handler) DeleteDignitary(ctx *gin.Context) {
	user, err := h.CurrentUser()
	if err != nil {
		h.internalError(ctx, err)
		return
	}
	id, ok := h.parseID(ctx)
	if !ok {
		return
	}

	err = h.Repository.DeleteDignitary(id, user.ID)
	switch {
	case errors.Is(err, repository.ErrNotFound):
		h.errorJSON(ctx, http.StatusNotFound, "сановник не найден")
	case errors.Is(err, repository.ErrForbidden):
		h.errorJSON(ctx, http.StatusForbidden, "удалять можно только своих сановников")
	case err != nil:
		h.internalError(ctx, err)
	default:
		ctx.JSON(http.StatusOK, gin.H{"status": "ok", "deleted_id": id})
	}
}

// LikeDignitary — POST /api/dignitaries/:id/like, тело {"liked": 1} или {"liked": 0}
func (h *Handler) LikeDignitary(ctx *gin.Context) {
	user, err := h.CurrentUser()
	if err != nil {
		h.internalError(ctx, err)
		return
	}
	id, ok := h.parseID(ctx)
	if !ok {
		return
	}

	var req serializer.LikeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorJSON(ctx, http.StatusBadRequest, "поле liked должно быть 0 или 1")
		return
	}

	dignitary, err := h.Repository.SetLike(id, user.ID, *req.Liked == 1)
	if errors.Is(err, repository.ErrNotFound) {
		h.errorJSON(ctx, http.StatusNotFound, "сановник не найден")
		return
	}
	if err != nil {
		h.internalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, serializer.LikeResponse{
		DignitaryID: dignitary.ID,
		IsLiked:     *req.Liked,
		LikesCount:  len(dignitary.Likes),
	})
}

func (h *Handler) parseID(ctx *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || id == 0 {
		h.errorJSON(ctx, http.StatusBadRequest, "некорректный id")
		return 0, false
	}
	return uint(id), true
}
