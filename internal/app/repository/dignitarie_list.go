package repository

import (
	"fmt"
	"sort"
	"time"
)

// Статусы услуги
const (
	StatusDraft     = "черновик"
	StatusPublished = "опубликован"
	StatusDeleted   = "удален"
)

// DignitarieList — услуга: сановник (визирь, консул, воевода) с датами пребывания в должности.
// Атомарные поля (1НФ): массивов с данными предметной области нет,
// кроме лайков — это ID пользователей, поставивших лайк.
type DignitarieList struct {
	ID          int
	Name        string    // наименование
	Description string    // краткое описание
	ImageKey    string    // ключ изображения в Minio
	VideoKey    string    // ключ видео в Minio
	StartDate   time.Time // поле по теме 1: дата вступления в должность
	EndDate     time.Time // поле по теме 2: дата оставления должности
	Status      string
	Likes       []int // ID пользователей, поставивших лайк
}

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func initialDignitarieLists() []DignitarieList {
	return []DignitarieList{
		{
			ID:          1,
			Name:        "Ибрагим-паша Паргалы",
			Description: "Великий визирь Османской империи, друг детства и зять Сулеймана Великолепного. Руководил походами на Венгрию: при нём состоялась битва при Мохаче (1526) и первая осада Вены (1529). Казнён по приказу султана.",
			ImageKey:    "ibrahim_pasha.jpg",
			VideoKey:    "ibrahim_pasha.mp4",
			StartDate:   date(1523, time.June, 27),
			EndDate:     date(1536, time.March, 15),
			Status:      StatusPublished,
			Likes:       []int{1, 2, 3},
		},
		{
			ID:          2,
			Name:        "Мехмед-паша Соколлу",
			Description: "Великий визирь Османской империи при трёх султанах: Сулеймане I, Селиме II и Мураде III. При нём взят Сигетвар (1566), заключён Адрианопольский мир с Габсбургами (1568) и проиграна битва при Лепанто (1571).",
			ImageKey:    "sokollu_mehmed.jpg",
			VideoKey:    "sokollu_mehmed.mp4",
			StartDate:   date(1565, time.June, 28),
			EndDate:     date(1579, time.October, 11),
			Status:      StatusPublished,
			Likes:       []int{2, 4},
		},
		{
			ID:          3,
			Name:        "Марк Ульпий Траян",
			Description: "Консул Римской империи 100 года вместе с Секстом Юлием Фронтином. Римский император 98–117 годов: при нём империя достигла наибольших размеров, были завоёваны Дакия и Парфия.",
			ImageKey:    "trajan.jpg",
			VideoKey:    "trajan.mp4",
			StartDate:   date(100, time.January, 1),
			EndDate:     date(100, time.December, 31),
			Status:      StatusPublished,
			Likes:       []int{1, 2, 3, 4, 5},
		},
		{
			ID:          4,
			Name:        "Марк Аврелий",
			Description: "Консул Римской империи 140 года вместе с Антонином Пием. Будущий император (161–180), философ-стоик, автор сочинения «К самому себе».",
			ImageKey:    "marcus_aurelius.jpg",
			VideoKey:    "marcus_aurelius.mp4",
			StartDate:   date(140, time.January, 1),
			EndDate:     date(140, time.December, 31),
			Status:      StatusPublished,
			Likes:       []int{3},
		},
		{
			ID:          5,
			Name:        "Михаил Борисович Шеин",
			Description: "Воевода Смоленска. Руководил обороной города от войск Сигизмунда III в 1609–1611 годах. Осада длилась 20 месяцев, город пал в июне 1611 года, а Шеин попал в польский плен.",
			ImageKey:    "shein.jpg",
			VideoKey:    "shein.mp4",
			StartDate:   date(1609, time.September, 16),
			EndDate:     date(1611, time.June, 3),
			Status:      StatusPublished,
			Likes:       []int{1, 5},
		},
		{
			ID:          6,
			Name:        "Кара-Мустафа-паша",
			Description: "Великий визирь Османской империи. Возглавил вторую осаду Вены в 1683 году. После поражения от войск Яна III Собеского казнён в Белграде по приказу султана Мехмеда IV.",
			ImageKey:    "kara_mustafa.jpg",
			VideoKey:    "kara_mustafa.mp4",
			StartDate:   date(1676, time.November, 1),
			EndDate:     date(1683, time.December, 25),
			Status:      StatusPublished,
			Likes:       []int{},
		},
		{
			ID:          7,
			Name:        "Рустем-паша",
			Description: "Великий визирь Османской империи, зять Сулеймана Великолепного, муж Михримах-султан. Известен финансовыми реформами и мечетью Рустема-паши в Стамбуле.",
			ImageKey:    "rustem_pasha.jpg",
			VideoKey:    "rustem_pasha.mp4",
			StartDate:   date(1544, time.October, 7),
			EndDate:     date(1553, time.October, 6),
			Status:      StatusDeleted,
			Likes:       []int{2},
		},
		{
			ID:          8,
			Name:        "Публий Элий Адриан",
			Description: "Консул Римской империи 118 года вместе с Гнеем Педанием Фуском Салинатором. Римский император 117–138 годов, построивший Адрианов вал в Британии.",
			ImageKey:    "hadrian.jpg",
			VideoKey:    "hadrian.mp4",
			StartDate:   date(118, time.January, 1),
			EndDate:     date(118, time.December, 31),
			Status:      StatusDraft,
			Likes:       []int{},
		},
	}
}

// GetDignitarieLists возвращает опубликованные услуги; если onDate != nil,
// то только те, кто находился в должности в указанную дату
func (r *Repository) GetDignitarieLists(onDate *time.Time) []DignitarieList {
	var result []DignitarieList
	for _, d := range r.dignitarieLists {
		if d.Status != StatusPublished {
			continue
		}
		if onDate != nil && (onDate.Before(d.StartDate) || onDate.After(d.EndDate)) {
			continue
		}
		result = append(result, d)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

// GetDignitarieList возвращает опубликованную услугу по ID
func (r *Repository) GetDignitarieList(id int) (DignitarieList, error) {
	for _, d := range r.dignitarieLists {
		if d.ID == id && d.Status == StatusPublished {
			return d, nil
		}
	}
	return DignitarieList{}, fmt.Errorf("услуга с id=%d не найдена", id)
}

// GetNextDignitarieList возвращает следующую опубликованную услугу после указанного ID
// (после последней — снова первую)
func (r *Repository) GetNextDignitarieList(id int) (DignitarieList, error) {
	published := r.GetDignitarieLists(nil)
	if len(published) == 0 {
		return DignitarieList{}, fmt.Errorf("нет опубликованных услуг")
	}
	for _, d := range published {
		if d.ID > id {
			return d, nil
		}
	}
	return published[0], nil
}

// GetDraftDignitarieList возвращает услугу в статусе черновик
func (r *Repository) GetDraftDignitarieList() (DignitarieList, error) {
	for _, d := range r.dignitarieLists {
		if d.Status == StatusDraft {
			return d, nil
		}
	}
	return DignitarieList{}, fmt.Errorf("черновик не найден")
}
