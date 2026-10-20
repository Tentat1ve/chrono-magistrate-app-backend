package handler

import (
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"awesomeProject/internal/app/ds"
	"awesomeProject/internal/app/repository"
	"awesomeProject/internal/app/serializer"

	"github.com/gin-gonic/gin"
)

const (
	dateLayout   = "2006-01-02"
	maxImageSize = 10 << 20 // 10 МБ
	maxVideoSize = 50 << 20 // 50 МБ
)

// GetDignitarieLists — GET /api/dignitarie_lists?date=1570-01-01
// Список опубликованных услуг с признаком is_creator; необязательный фильтр —
// сановники, находившиеся в должности в указанную дату.
func (h *Handler) GetDignitarieLists(ctx *gin.Context) {
	user, err := h.CurrentUser()
	if err != nil {
		h.internalError(ctx, err)
		return
	}

	var onDate *time.Time
	if dateQuery := ctx.Query("date"); dateQuery != "" {
		parsed, err := time.Parse(dateLayout, dateQuery)
		if err != nil {
			h.errorJSON(ctx, http.StatusBadRequest, "date должна быть в формате ГГГГ-ММ-ДД")
			return
		}
		onDate = &parsed
	}

	dignitarieLists, err := h.Repository.GetDignitarieLists(onDate)
	if err != nil {
		h.internalError(ctx, err)
		return
	}

	result := make([]serializer.DignitarieListItemJSON, 0, len(dignitarieLists))
	for _, d := range dignitarieLists {
		result = append(result, serializer.DignitarieListToItem(d, user.ID))
	}
	ctx.JSON(http.StatusOK, result)
}

// GetDignitarieListFeed — GET /api/dignitarie_lists/feed[?id=3[&next=true]]
// Одна опубликованная услуга с признаками is_liked и is_creator.
// Без id (переход с панели вкладок) возвращается первая услуга, с id — выбранная,
// с id и next=true — следующая после неё.
func (h *Handler) GetDignitarieListFeed(ctx *gin.Context) {
	user, err := h.CurrentUser()
	if err != nil {
		h.internalError(ctx, err)
		return
	}

	var dignitarieList ds.DignitarieList
	if idQuery := ctx.Query("id"); idQuery == "" {
		dignitarieList, err = h.Repository.GetNextDignitarieList(0)
	} else {
		id, convErr := strconv.ParseUint(idQuery, 10, 64)
		if convErr != nil || id == 0 {
			h.errorJSON(ctx, http.StatusBadRequest, "некорректный id")
			return
		}
		if ctx.Query("next") == "true" {
			dignitarieList, err = h.Repository.GetNextDignitarieList(uint(id))
		} else {
			dignitarieList, err = h.Repository.GetDignitarieList(uint(id))
		}
	}

	if errors.Is(err, repository.ErrNotFound) {
		h.errorJSON(ctx, http.StatusNotFound, "услуга не найдена")
		return
	}
	if err != nil {
		h.internalError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, serializer.DignitarieListToFeed(dignitarieList, user.ID))
}

// GetDignitarieListDraft — GET /api/dignitarie_lists/draft — черновик текущего пользователя
func (h *Handler) GetDignitarieListDraft(ctx *gin.Context) {
	user, err := h.CurrentUser()
	if err != nil {
		h.internalError(ctx, err)
		return
	}

	draft, err := h.Repository.GetDraftDignitarieList(user.ID)
	if errors.Is(err, repository.ErrNotFound) {
		h.errorJSON(ctx, http.StatusNotFound, "у пользователя нет черновика")
		return
	}
	if err != nil {
		h.internalError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, serializer.DignitarieListToJSON(draft))
}

// CreateDignitarieList — POST /api/dignitarie_lists (multipart/form-data: name, image, video)
// Создаёт черновик; файлы сохраняются в Minio под сгенерированными латинскими именами.
func (h *Handler) CreateDignitarieList(ctx *gin.Context) {
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

	if _, err := h.Repository.GetDraftDignitarieList(user.ID); err == nil {
		h.errorJSON(ctx, http.StatusConflict, "у пользователя уже есть черновик")
		return
	}

	imageURL, err := h.upload(ctx, "dignitarie_list_image", image)
	if err != nil {
		h.internalError(ctx, err)
		return
	}
	videoURL, err := h.upload(ctx, "dignitarie_list_video", video)
	if err != nil {
		h.Repository.RemoveFile(ctx, imageURL)
		h.internalError(ctx, err)
		return
	}

	dignitarieList := ds.DignitarieList{Name: name, ImageURL: imageURL, VideoURL: videoURL, CreatorID: user.ID}
	if err := h.Repository.CreateDraftDignitarieList(&dignitarieList); err != nil {
		h.Repository.RemoveFile(ctx, imageURL)
		h.Repository.RemoveFile(ctx, videoURL)
		if errors.Is(err, repository.ErrAlreadyExists) {
			h.errorJSON(ctx, http.StatusConflict, "у пользователя уже есть черновик")
			return
		}
		h.internalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, serializer.DignitarieListToJSON(dignitarieList))
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

// PublishDignitarieListDraft — PUT /api/dignitarie_lists/draft/publish
// Заполняет необязательные поля черновика и меняет статус на «опубликован».
func (h *Handler) PublishDignitarieListDraft(ctx *gin.Context) {
	user, err := h.CurrentUser()
	if err != nil {
		h.internalError(ctx, err)
		return
	}

	var req serializer.PublishDignitarieListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorJSON(ctx, http.StatusBadRequest, "заполните description, start_date и end_date")
		return
	}
	startDate, errStart := time.Parse(dateLayout, req.StartDate)
	endDate, errEnd := time.Parse(dateLayout, req.EndDate)
	if errStart != nil || errEnd != nil {
		h.errorJSON(ctx, http.StatusBadRequest, "start_date и end_date должны быть в формате ГГГГ-ММ-ДД")
		return
	}
	if startDate.After(endDate) {
		h.errorJSON(ctx, http.StatusBadRequest, "start_date не может быть позже end_date")
		return
	}

	dignitarieList, err := h.Repository.PublishDignitarieList(user.ID, strings.TrimSpace(req.Description), startDate, endDate)
	if errors.Is(err, repository.ErrNotFound) {
		h.errorJSON(ctx, http.StatusNotFound, "у пользователя нет черновика")
		return
	}
	if err != nil {
		h.internalError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, serializer.DignitarieListToJSON(dignitarieList))
}

// DeleteDignitarieList — DELETE /api/dignitarie_lists/:id — логическое удаление своей услуги
func (h *Handler) DeleteDignitarieList(ctx *gin.Context) {
	user, err := h.CurrentUser()
	if err != nil {
		h.internalError(ctx, err)
		return
	}
	id, ok := h.parseID(ctx)
	if !ok {
		return
	}

	err = h.Repository.DeleteDignitarieList(id, user.ID)
	switch {
	case errors.Is(err, repository.ErrNotFound):
		h.errorJSON(ctx, http.StatusNotFound, "услуга не найдена")
	case errors.Is(err, repository.ErrForbidden):
		h.errorJSON(ctx, http.StatusForbidden, "удалять можно только свои услуги")
	case err != nil:
		h.internalError(ctx, err)
	default:
		ctx.JSON(http.StatusOK, gin.H{"status": "ok", "deleted_id": id})
	}
}

// LikeDignitarieList — POST /api/dignitarie_lists/:id/like, тело {"liked": 1} или {"liked": 0}
func (h *Handler) LikeDignitarieList(ctx *gin.Context) {
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

	dignitarieList, err := h.Repository.SetLike(id, user.ID, *req.Liked == 1)
	if errors.Is(err, repository.ErrNotFound) {
		h.errorJSON(ctx, http.StatusNotFound, "услуга не найдена")
		return
	}
	if err != nil {
		h.internalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, serializer.LikeResponse{
		DignitarieListID: dignitarieList.ID,
		IsLiked:          *req.Liked,
		LikesCount:       len(dignitarieList.Likes),
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
