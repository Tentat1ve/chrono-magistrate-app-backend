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

// GetDignitarieLists godoc
// @Summary      Список услуг
// @Description  Опубликованные услуги (плитка). Доступен гостю; для авторизованного пользователя признак is_creator = 1 у его услуг. Необязательный фильтр date — сановники, находившиеся в должности в указанную дату.
// @Tags         Услуги
// @Produce      json
// @Param        date  query     string  false  "Дата в формате ГГГГ-ММ-ДД"  example(1570-01-01)
// @Success      200   {array}   serializer.DignitarieListItemJSON
// @Failure      400   {object}  serializer.ErrorJSON
// @Router       /dignitarie_lists [get]
func (h *Handler) GetDignitarieLists(ctx *gin.Context) {
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

	userID := currentUserID(ctx)
	result := make([]serializer.DignitarieListItemJSON, 0, len(dignitarieLists))
	for _, d := range dignitarieLists {
		result = append(result, serializer.DignitarieListToItem(d, userID))
	}
	ctx.JSON(http.StatusOK, result)
}

// GetDignitarieListFeed godoc
// @Summary      Лента
// @Description  Одна опубликованная услуга. Без параметров (переход с панели вкладок) — первая услуга, id — выбранная, id и next=true — следующая после неё. Доступна гостю; для авторизованного пользователя признаки is_liked и is_creator.
// @Tags         Услуги
// @Produce      json
// @Param        id    query     int   false  "ID услуги"
// @Param        next  query     bool  false  "true — вернуть следующую услугу после id"
// @Success      200   {object}  serializer.DignitarieListFeedJSON
// @Failure      400   {object}  serializer.ErrorJSON
// @Failure      404   {object}  serializer.ErrorJSON
// @Router       /dignitarie_lists/feed [get]
func (h *Handler) GetDignitarieListFeed(ctx *gin.Context) {
	var (
		dignitarieList ds.DignitarieList
		err            error
	)
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
	ctx.JSON(http.StatusOK, serializer.DignitarieListToFeed(dignitarieList, currentUserID(ctx)))
}

// GetDignitarieListDraft godoc
// @Summary      Черновик
// @Description  Черновик текущего пользователя (не более одного), id не указывается. Требует авторизации.
// @Tags         Услуги
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  serializer.DignitarieListJSON
// @Failure      401  {object}  serializer.ErrorJSON
// @Failure      404  {object}  serializer.ErrorJSON
// @Router       /dignitarie_lists/draft [get]
func (h *Handler) GetDignitarieListDraft(ctx *gin.Context) {
	draft, err := h.Repository.GetDraftDignitarieList(currentUserID(ctx))
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

// CreateDignitarieList godoc
// @Summary      Добавление услуги
// @Description  Создаёт черновик. Файлы изображения и видео сохраняются в MinIO под сгенерированными латинскими именами, в БД записываются их url. Создатель подставляется из токена. Требует авторизации.
// @Tags         Услуги
// @Accept       mpfd
// @Produce      json
// @Security     ApiKeyAuth
// @Param        name   formData  string  true  "Наименование"
// @Param        image  formData  file    true  "Изображение (image/*, до 10 МБ)"
// @Param        video  formData  file    true  "Короткое видео (video/*, до 50 МБ)"
// @Success      201    {object}  serializer.DignitarieListJSON
// @Failure      400    {object}  serializer.ErrorJSON
// @Failure      401    {object}  serializer.ErrorJSON
// @Failure      409    {object}  serializer.ErrorJSON
// @Router       /dignitarie_lists [post]
func (h *Handler) CreateDignitarieList(ctx *gin.Context) {
	userID := currentUserID(ctx)

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

	if _, err := h.Repository.GetDraftDignitarieList(userID); err == nil {
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

	dignitarieList := ds.DignitarieList{Name: name, ImageURL: imageURL, VideoURL: videoURL, CreatorID: userID}
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

// PublishDignitarieListDraft godoc
// @Summary      Публикация услуги
// @Description  Заполняет необязательные поля черновика текущего пользователя и меняет статус на «опубликован». Для гостя — 401. Требует авторизации.
// @Tags         Услуги
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        body  body      serializer.PublishDignitarieListRequest  true  "Краткое описание и две даты по теме"
// @Success      200   {object}  serializer.DignitarieListJSON
// @Failure      400   {object}  serializer.ErrorJSON
// @Failure      401   {object}  serializer.ErrorJSON
// @Failure      404   {object}  serializer.ErrorJSON
// @Router       /dignitarie_lists/draft/publish [put]
func (h *Handler) PublishDignitarieListDraft(ctx *gin.Context) {
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

	dignitarieList, err := h.Repository.PublishDignitarieList(currentUserID(ctx), strings.TrimSpace(req.Description), startDate, endDate)
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

// DeleteDignitarieList godoc
// @Summary      Удаление услуги
// @Description  Логическое удаление (статус «удален») своей услуги. Требует авторизации.
// @Tags         Услуги
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      int  true  "ID услуги"
// @Success      200  {object}  serializer.DeleteResponse
// @Failure      400  {object}  serializer.ErrorJSON
// @Failure      401  {object}  serializer.ErrorJSON
// @Failure      403  {object}  serializer.ErrorJSON
// @Failure      404  {object}  serializer.ErrorJSON
// @Router       /dignitarie_lists/{id} [delete]
func (h *Handler) DeleteDignitarieList(ctx *gin.Context) {
	id, ok := h.parseID(ctx)
	if !ok {
		return
	}

	err := h.Repository.DeleteDignitarieList(id, currentUserID(ctx))
	switch {
	case errors.Is(err, repository.ErrNotFound):
		h.errorJSON(ctx, http.StatusNotFound, "услуга не найдена")
	case errors.Is(err, repository.ErrForbidden):
		h.errorJSON(ctx, http.StatusForbidden, "удалять можно только свои услуги")
	case err != nil:
		h.internalError(ctx, err)
	default:
		ctx.JSON(http.StatusOK, serializer.DeleteResponse{Status: "ok", DeletedID: id})
	}
}

// LikeDignitarieList godoc
// @Summary      Лайк
// @Description  Ставит (liked = 1) или отменяет (liked = 0) лайк текущего пользователя. Требует авторизации.
// @Tags         Услуги
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id    path      int                      true  "ID услуги"
// @Param        body  body      serializer.LikeRequest   true  "liked: 1 — поставить, 0 — отменить"
// @Success      200   {object}  serializer.LikeResponse
// @Failure      400   {object}  serializer.ErrorJSON
// @Failure      401   {object}  serializer.ErrorJSON
// @Failure      404   {object}  serializer.ErrorJSON
// @Router       /dignitarie_lists/{id}/like [post]
func (h *Handler) LikeDignitarieList(ctx *gin.Context) {
	id, ok := h.parseID(ctx)
	if !ok {
		return
	}

	var req serializer.LikeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorJSON(ctx, http.StatusBadRequest, "поле liked должно быть 0 или 1")
		return
	}

	dignitarieList, err := h.Repository.SetLike(id, currentUserID(ctx), *req.Liked == 1)
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
