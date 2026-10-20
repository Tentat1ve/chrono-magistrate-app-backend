package handler

import (
	"errors"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"awesomeProject/internal/app/ds"
	"awesomeProject/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const dateLayout = "2006-01-02" // формат поля input type="date"

// DignitarieListView — данные услуги для шаблона: проверенные ссылки на медиа,
// число лайков и даты в виде текста
type DignitarieListView struct {
	ds.DignitarieList
	ImageSrc        string
	VideoSrc        string
	LikesCount      int
	DescriptionText string
	StartText       string // дата начала для отображения
	EndText         string // дата конца для отображения
	StartInput      string // дата начала для поля input type="date"
	EndInput        string // дата конца для поля input type="date"
}

func (h *Handler) toView(d ds.DignitarieList) DignitarieListView {
	view := DignitarieListView{
		DignitarieList: d,
		ImageSrc:       h.mediaURL(d.ImageURL, defaultImageURL),
		VideoSrc:       h.mediaURL(d.VideoURL, defaultVideoURL),
		LikesCount:     len(d.Likes), // количество лайков вычисляется по коллекции
		StartText:      "—",
		EndText:        "—",
	}
	if d.Description != nil {
		view.DescriptionText = *d.Description
	}
	if d.StartDate != nil {
		view.StartText = d.StartDate.Format("02.01.2006")
		view.StartInput = d.StartDate.Format(dateLayout)
	}
	if d.EndDate != nil {
		view.EndText = d.EndDate.Format("02.01.2006")
		view.EndInput = d.EndDate.Format(dateLayout)
	}
	return view
}

// GetDignitarieListCatalog — страница «плитка»: GET /dignitarie_lists/catalog?date=1570-01-01
func (h *Handler) GetDignitarieListCatalog(ctx *gin.Context) {
	dateQuery := ctx.Query("date")

	var onDate *time.Time
	if dateQuery != "" {
		parsed, err := time.Parse(dateLayout, dateQuery)
		if err != nil {
			logrus.Error(err)
			ctx.Redirect(http.StatusSeeOther, catalogURL)
			return
		}
		onDate = &parsed
	}

	dignitarieLists, err := h.Repository.GetDignitarieLists(onDate)
	if err != nil {
		logrus.Error(err)
		ctx.Redirect(http.StatusSeeOther, catalogURL)
		return
	}

	views := make([]DignitarieListView, 0, len(dignitarieLists))
	for _, d := range dignitarieLists {
		views = append(views, h.toView(d))
	}

	ctx.HTML(http.StatusOK, "dignitarie_lists_catalog.html", gin.H{
		"dignitarieLists": views,
		"date":            dateQuery,
	})
}

// GetDignitarieListFeed — страница «лента»:
// GET /dignitarie_lists/feed, ?id=3, ?id=3&next=true
func (h *Handler) GetDignitarieListFeed(ctx *gin.Context) {
	var (
		dignitarieList ds.DignitarieList
		err            error
	)

	idQuery := ctx.Query("id")
	if idQuery == "" {
		dignitarieList, err = h.Repository.GetNextDignitarieList(0)
	} else {
		id, convErr := strconv.ParseUint(idQuery, 10, 64)
		if convErr != nil {
			logrus.Error(convErr)
			ctx.Redirect(http.StatusSeeOther, catalogURL)
			return
		}
		if ctx.Query("next") == "true" {
			dignitarieList, err = h.Repository.GetNextDignitarieList(uint(id))
		} else {
			dignitarieList, err = h.Repository.GetDignitarieList(uint(id))
		}
	}

	if err != nil { // услуга не найдена или удалена — перенаправляем на плитку
		logrus.Error(err)
		ctx.Redirect(http.StatusSeeOther, catalogURL)
		return
	}

	ctx.HTML(http.StatusOK, "dignitarie_lists_feed.html", gin.H{
		"dignitarieList": h.toView(dignitarieList),
	})
}

// GetDignitarieListDraft — страница «добавление»: GET /dignitarie_lists/draft.
// Если черновика нет — форма создания (название, фото, видео, «Далее»),
// если есть — заполненная форма с кнопкой «Опубликовать».
func (h *Handler) GetDignitarieListDraft(ctx *gin.Context) {
	h.renderDraft(ctx, "")
}

func (h *Handler) renderDraft(ctx *gin.Context, formError string) {
	draft, err := h.Repository.GetDraftDignitarieList(currentUserID)
	if errors.Is(err, repository.ErrNotFound) {
		ctx.HTML(http.StatusOK, "dignitarie_lists_draft.html", gin.H{"error": formError})
		return
	}
	if err != nil {
		logrus.Error(err)
		ctx.Redirect(http.StatusSeeOther, catalogURL)
		return
	}

	ctx.HTML(http.StatusOK, "dignitarie_lists_draft.html", gin.H{
		"draft": h.toView(draft),
		"error": formError,
	})
}

// CreateDignitarieListDraft — кнопка «Далее»: POST /dignitarie_lists/draft.
// Файлы на сервер не передаются: форма отправляет только имена файлов,
// из которых формируются (неверные) url в Minio.
func (h *Handler) CreateDignitarieListDraft(ctx *gin.Context) {
	if _, err := h.Repository.GetDraftDignitarieList(currentUserID); err == nil {
		ctx.Redirect(http.StatusSeeOther, "/dignitarie_lists/draft")
		return
	}

	name := strings.TrimSpace(ctx.PostForm("name"))
	if name == "" {
		h.renderDraft(ctx, "Укажите название")
		return
	}

	imageURL := h.uploadedFileURL(ctx.PostForm("image"))
	videoURL := h.uploadedFileURL(ctx.PostForm("video"))

	if err := h.Repository.CreateDraftDignitarieList(currentUserID, name, imageURL, videoURL); err != nil {
		logrus.Error(err)
	}
	ctx.Redirect(http.StatusSeeOther, "/dignitarie_lists/draft")
}

func (h *Handler) uploadedFileURL(fileName string) string {
	fileName = path.Base(strings.ReplaceAll(fileName, "\\", "/"))
	if fileName == "" || fileName == "." || fileName == "/" {
		return ""
	}
	return h.minioURL + "/" + fileName
}

// PublishDignitarieListDraft — кнопка «Опубликовать»: POST /dignitarie_lists/draft/publish
func (h *Handler) PublishDignitarieListDraft(ctx *gin.Context) {
	description := strings.TrimSpace(ctx.PostForm("description"))
	startDate, errStart := time.Parse(dateLayout, ctx.PostForm("start_date"))
	endDate, errEnd := time.Parse(dateLayout, ctx.PostForm("end_date"))

	switch {
	case description == "":
		h.renderDraft(ctx, "Заполните краткое описание")
		return
	case errStart != nil || errEnd != nil:
		h.renderDraft(ctx, "Укажите дату начала и дату окончания пребывания в должности")
		return
	case startDate.After(endDate):
		h.renderDraft(ctx, "Дата начала не может быть позже даты окончания")
		return
	}

	if err := h.Repository.PublishDignitarieList(currentUserID, description, startDate, endDate); err != nil {
		logrus.Error(err)
		ctx.Redirect(http.StatusSeeOther, "/dignitarie_lists/draft")
		return
	}
	ctx.Redirect(http.StatusSeeOther, catalogURL)
}

// DeleteDignitarieList — логическое удаление: POST /dignitarie_lists/delete (поле id)
func (h *Handler) DeleteDignitarieList(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.PostForm("id"), 10, 64)
	if err != nil {
		logrus.Error(err)
		ctx.Redirect(http.StatusSeeOther, catalogURL)
		return
	}

	if err := h.Repository.DeleteDignitarieList(ctx.Request.Context(), uint(id)); err != nil {
		logrus.Error(err)
	}
	ctx.Redirect(http.StatusSeeOther, catalogURL)
}
