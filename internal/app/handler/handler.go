package handler

import (
	"net/http"
	"strconv"

	"awesomeProject/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

// DignitaryView — данные сановника для шаблона: ссылки на Minio и число лайков
type DignitaryView struct {
	repository.Dignitary
	ImageURL   string
	VideoURL   string
	LikesCount int
}

func (h *Handler) toView(d repository.Dignitary) DignitaryView {
	return DignitaryView{
		Dignitary:  d,
		ImageURL:   h.Repository.MediaURL(d.ImageKey),
		VideoURL:   h.Repository.MediaURL(d.VideoKey),
		LikesCount: len(d.LikedBy), // количество лайков вычисляется по коллекции
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

	var views []DignitaryView
	for _, d := range h.Repository.GetDignitaries(year) {
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
		dignitary repository.Dignitary
		err       error
	)

	idStr := ctx.Param("id")
	if idStr == "" {
		dignitary, err = h.Repository.GetFirstDignitary()
	} else {
		id, convErr := strconv.Atoi(idStr)
		if convErr != nil {
			logrus.Error(convErr)
			ctx.String(http.StatusBadRequest, "некорректный id")
			return
		}
		if ctx.Query("next") == "true" {
			dignitary, err = h.Repository.GetNextDignitary(id)
		} else {
			dignitary, err = h.Repository.GetDignitary(id)
		}
	}

	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusNotFound, err.Error())
		return
	}

	ctx.HTML(http.StatusOK, "dignitary_feed.html", gin.H{
		"dignitary": h.toView(dignitary),
	})
}

// GetDignitaryDraft — страница «добавление»: GET /dignitary_draft
func (h *Handler) GetDignitaryDraft(ctx *gin.Context) {
	draft, err := h.Repository.GetDraftDignitary()
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusNotFound, err.Error())
		return
	}

	ctx.HTML(http.StatusOK, "dignitary_draft.html", gin.H{
		"dignitary": h.toView(draft),
	})
}
