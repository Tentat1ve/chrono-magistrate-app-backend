package handler

import (
	"net/http"
	"strconv"
	"time"

	"awesomeProject/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const dateLayout = "2006-01-02" // формат поля input type="date"

const catalogURL = "/dignitarie_lists/catalog"

// DignitarieListView — данные услуги для шаблона: ссылки на Minio и число лайков
type DignitarieListView struct {
	repository.DignitarieList
	ImageURL   string
	VideoURL   string
	LikesCount int
	StartInput string // дата начала для поля input type="date"
	EndInput   string // дата конца для поля input type="date"
}

func (h *Handler) toView(d repository.DignitarieList) DignitarieListView {
	return DignitarieListView{
		DignitarieList: d,
		ImageURL:       h.Repository.MediaURL(d.ImageKey),
		VideoURL:       h.Repository.MediaURL(d.VideoKey),
		LikesCount:     len(d.Likes), // количество лайков вычисляется по коллекции
		StartInput:     d.StartDate.Format(dateLayout),
		EndInput:       d.EndDate.Format(dateLayout),
	}
}

// GetDignitarieListCatalog — страница «плитка»: GET /dignitarie_lists/catalog?date=1570-01-01
// Фильтр на сервере: сановники, находившиеся в должности в выбранную дату.
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

	views := make([]DignitarieListView, 0)
	for _, d := range h.Repository.GetDignitarieLists(onDate) {
		views = append(views, h.toView(d))
	}

	ctx.HTML(http.StatusOK, "dignitarie_lists_catalog.html", gin.H{
		"dignitarieLists": views,
		"date":            dateQuery,
	})
}

// GetDignitarieListFeed — страница «лента»:
// GET /dignitarie_lists/feed — первая услуга (для панели вкладок id не указывается),
// GET /dignitarie_lists/feed?id=3 — услуга по ID,
// GET /dignitarie_lists/feed?id=3&next=true — следующая после ID.
func (h *Handler) GetDignitarieListFeed(ctx *gin.Context) {
	var (
		dignitarieList repository.DignitarieList
		err            error
	)

	idQuery := ctx.Query("id")
	if idQuery == "" {
		dignitarieList, err = h.Repository.GetNextDignitarieList(0)
	} else {
		id, convErr := strconv.Atoi(idQuery)
		if convErr != nil {
			logrus.Error(convErr)
			ctx.Redirect(http.StatusSeeOther, catalogURL)
			return
		}
		if ctx.Query("next") == "true" {
			dignitarieList, err = h.Repository.GetNextDignitarieList(id)
		} else {
			dignitarieList, err = h.Repository.GetDignitarieList(id)
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

// GetDignitarieListDraft — страница «добавление»: GET /dignitarie_lists/draft
func (h *Handler) GetDignitarieListDraft(ctx *gin.Context) {
	draft, err := h.Repository.GetDraftDignitarieList()
	if err != nil {
		logrus.Error(err)
		ctx.Redirect(http.StatusSeeOther, catalogURL)
		return
	}

	ctx.HTML(http.StatusOK, "dignitarie_lists_draft.html", gin.H{
		"dignitarieList": h.toView(draft),
	})
}
