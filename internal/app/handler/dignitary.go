package handler

import (
	"errors"
	"net/http"
	"path"
	"strconv"
	"strings"

	"awesomeProject/internal/app/ds"
	"awesomeProject/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// DignitaryView — данные сановника для шаблона: проверенные ссылки на медиа и число лайков
type DignitaryView struct {
	ds.Dignitary
	ImageSrc   string
	VideoSrc   string
	LikesCount int
}

func (h *Handler) toView(d ds.Dignitary) DignitaryView {
	return DignitaryView{
		Dignitary:  d,
		ImageSrc:   h.mediaURL(d.ImageURL, defaultImageURL),
		VideoSrc:   h.mediaURL(d.VideoURL, defaultVideoURL),
		LikesCount: len(d.Likes),
	}
}

// GetDignitaries — страница «плитка»: GET /dignitaries?office_year=1570
func (h *Handler) GetDignitaries(ctx *gin.Context) {
	yearQuery := ctx.Query("office_year")

	var year *int
	if yearQuery != "" {
		y, err := strconv.Atoi(yearQuery)
		if err != nil {
			logrus.Error(err)
		} else {
			year = &y
		}
	}

	dignitaries, err := h.Repository.GetDignitaries(year)
	if err != nil {
		logrus.Error(err)
		h.errorPage(ctx, http.StatusInternalServerError, "Не удалось получить список сановников")
		return
	}

	views := make([]DignitaryView, 0, len(dignitaries))
	for _, d := range dignitaries {
		views = append(views, h.toView(d))
	}

	ctx.HTML(http.StatusOK, "dignitaries.html", gin.H{
		"dignitaries": views,
		"officeYear":  yearQuery,
	})
}

// GetDignitaryFeed — страница «лента»: GET /dignitary_feed, /dignitary_feed/:id, /dignitary_feed/:id?next=true
func (h *Handler) GetDignitaryFeed(ctx *gin.Context) {
	var (
		dignitary ds.Dignitary
		err       error
	)

	idStr := ctx.Param("id")
	if idStr == "" {
		dignitary, err = h.Repository.GetNextDignitary(0)
	} else {
		id, convErr := strconv.ParseUint(idStr, 10, 64)
		if convErr != nil {
			h.errorPage(ctx, http.StatusBadRequest, "Некорректный id сановника")
			return
		}
		if ctx.Query("next") == "true" {
			dignitary, err = h.Repository.GetNextDignitary(uint(id))
		} else {
			dignitary, err = h.Repository.GetDignitary(uint(id))
		}
	}

	if errors.Is(err, repository.ErrNotFound) {
		h.errorPage(ctx, http.StatusNotFound, "Сановник не найден или удалён")
		return
	}
	if err != nil {
		logrus.Error(err)
		h.errorPage(ctx, http.StatusInternalServerError, "Не удалось получить сановника")
		return
	}

	ctx.HTML(http.StatusOK, "dignitary_feed.html", gin.H{
		"dignitary": h.toView(dignitary),
	})
}

// GetDignitaryDraft — страница «добавление»: GET /dignitary_draft.
// Если черновика нет — форма создания (название, фото, видео, «Далее»),
// если есть — заполненная форма публикации.
func (h *Handler) GetDignitaryDraft(ctx *gin.Context) {
	h.renderDraft(ctx, http.StatusOK, "")
}

func (h *Handler) renderDraft(ctx *gin.Context, status int, formError string) {
	draft, err := h.Repository.GetDraftDignitary(currentUserID)
	if errors.Is(err, repository.ErrNotFound) {
		ctx.HTML(status, "dignitary_draft.html", gin.H{"error": formError})
		return
	}
	if err != nil {
		logrus.Error(err)
		h.errorPage(ctx, http.StatusInternalServerError, "Не удалось получить черновик")
		return
	}

	ctx.HTML(status, "dignitary_draft.html", gin.H{
		"draft": h.toView(draft),
		"error": formError,
	})
}

// CreateDignitaryDraft — кнопка «Далее»: POST /dignitary_draft.
// Файлы на сервер не передаются: форма отправляет только имена файлов,
// из которых формируются (неверные) url в Minio.
func (h *Handler) CreateDignitaryDraft(ctx *gin.Context) {
	if _, err := h.Repository.GetDraftDignitary(currentUserID); err == nil {
		ctx.Redirect(http.StatusSeeOther, "/dignitary_draft")
		return
	}

	name := strings.TrimSpace(ctx.PostForm("dignitary_name"))
	if name == "" {
		h.renderDraft(ctx, http.StatusBadRequest, "Укажите имя сановника")
		return
	}

	imageURL := h.uploadedFileURL(ctx.PostForm("dignitary_image"))
	videoURL := h.uploadedFileURL(ctx.PostForm("dignitary_video"))

	if err := h.Repository.CreateDraftDignitary(currentUserID, name, imageURL, videoURL); err != nil {
		logrus.Error(err)
		h.errorPage(ctx, http.StatusInternalServerError, "Не удалось создать черновик")
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/dignitary_draft")
}

func (h *Handler) uploadedFileURL(fileName string) string {
	fileName = path.Base(strings.ReplaceAll(fileName, "\\", "/"))
	if fileName == "" || fileName == "." || fileName == "/" {
		return ""
	}
	return h.minioURL + "/" + fileName
}

// PublishDignitaryDraft — кнопка «Опубликовать»: POST /dignitary_draft/publish
func (h *Handler) PublishDignitaryDraft(ctx *gin.Context) {
	office := strings.TrimSpace(ctx.PostForm("dignitary_office"))
	description := strings.TrimSpace(ctx.PostForm("dignitary_description"))
	officeStart, errStart := strconv.Atoi(ctx.PostForm("office_start"))
	officeEnd, errEnd := strconv.Atoi(ctx.PostForm("office_end"))

	switch {
	case office == "" || description == "":
		h.renderDraft(ctx, http.StatusBadRequest, "Заполните должность и краткое описание")
		return
	case errStart != nil || errEnd != nil:
		h.renderDraft(ctx, http.StatusBadRequest, "Укажите годы начала и окончания пребывания в должности")
		return
	case officeStart > officeEnd:
		h.renderDraft(ctx, http.StatusBadRequest, "Год начала не может быть позже года окончания")
		return
	}

	err := h.Repository.PublishDignitary(currentUserID, office, description, officeStart, officeEnd)
	if errors.Is(err, repository.ErrNotFound) {
		ctx.Redirect(http.StatusSeeOther, "/dignitary_draft")
		return
	}
	if err != nil {
		logrus.Error(err)
		h.errorPage(ctx, http.StatusInternalServerError, "Не удалось опубликовать сановника")
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/dignitaries")
}

// DeleteDignitary — логическое удаление: POST /dignitaries/:id/delete
func (h *Handler) DeleteDignitary(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.errorPage(ctx, http.StatusBadRequest, "Некорректный id сановника")
		return
	}

	err = h.Repository.DeleteDignitary(uint(id))
	if errors.Is(err, repository.ErrNotFound) {
		h.errorPage(ctx, http.StatusNotFound, "Сановник не найден или уже удалён")
		return
	}
	if err != nil {
		logrus.Error(err)
		h.errorPage(ctx, http.StatusInternalServerError, "Не удалось удалить сановника")
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/dignitaries")
}
