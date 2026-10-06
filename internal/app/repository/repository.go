package repository

import (
	"fmt"
	"sort"
)

// Статусы услуги-сановника
const (
	StatusDraft     = "черновик"
	StatusPublished = "опубликован"
	StatusDeleted   = "удален"
)

// Dignitary — услуга: сановник (визирь, консул, воевода) с годами пребывания в должности.
// Годы до н.э. хранятся отрицательными числами (-63 = 63 г. до н.э.).
type Dignitary struct {
	ID          int
	Name        string // наименование
	Office      string // должность и государство
	Description string // краткое описание
	ImageKey    string // ключ изображения в Minio
	VideoKey    string // ключ видео в Minio
	OfficeStart int    // поле по теме 1: год вступления в должность
	OfficeEnd   int    // поле по теме 2: год оставления должности
	Status      string
	LikedBy     []int // ID пользователей, поставивших лайк
}

type Repository struct {
	minioURL    string
	dignitaries []Dignitary
}

func NewRepository(minioURL string) *Repository {
	return &Repository{
		minioURL: minioURL,
		dignitaries: []Dignitary{
			{
				ID:          1,
				Name:        "Ибрагим-паша Паргалы",
				Office:      "Великий визирь Османской империи",
				Description: "Друг детства и зять Сулеймана Великолепного. Руководил походами на Венгрию, при нём состоялась битва при Мохаче (1526) и первая осада Вены (1529). Казнён по приказу султана в 1536 году.",
				ImageKey:    "ibrahim_pasha.jpg",
				VideoKey:    "ibrahim_pasha.mp4",
				OfficeStart: 1523,
				OfficeEnd:   1536,
				Status:      StatusPublished,
				LikedBy:     []int{1, 2, 3},
			},
			{
				ID:          2,
				Name:        "Мехмед-паша Соколлу",
				Office:      "Великий визирь Османской империи",
				Description: "Служил трём султанам: Сулейману I, Селиму II и Мураду III. При нём взят Сигетвар (1566), заключён Адрианопольский мир с Габсбургами (1568) и проиграна битва при Лепанто (1571).",
				ImageKey:    "sokollu_mehmed.jpg",
				VideoKey:    "sokollu_mehmed.mp4",
				OfficeStart: 1565,
				OfficeEnd:   1579,
				Status:      StatusPublished,
				LikedBy:     []int{2, 4},
			},
			{
				ID:          3,
				Name:        "Марк Туллий Цицерон",
				Office:      "Консул Римской республики",
				Description: "Оратор и политик. В год его консульства был раскрыт заговор Катилины, а сам Цицерон получил титул «Отец Отечества». Коллега по консульству — Гай Антоний Гибрида.",
				ImageKey:    "cicero.jpg",
				VideoKey:    "cicero.mp4",
				OfficeStart: -63,
				OfficeEnd:   -63,
				Status:      StatusPublished,
				LikedBy:     []int{1, 2, 3, 4, 5},
			},
			{
				ID:          4,
				Name:        "Гай Юлий Цезарь",
				Office:      "Консул Римской республики",
				Description: "Первое консульство Цезаря вместе с Марком Кальпурнием Бибулом. В этот год был принят аграрный закон, а Цезарь получил наместничество в Галлии.",
				ImageKey:    "caesar.jpg",
				VideoKey:    "caesar.mp4",
				OfficeStart: -59,
				OfficeEnd:   -59,
				Status:      StatusPublished,
				LikedBy:     []int{3},
			},
			{
				ID:          5,
				Name:        "Михаил Борисович Шеин",
				Office:      "Воевода Смоленска",
				Description: "Руководил обороной Смоленска от войск Сигизмунда III в 1609–1611 годах. Осада длилась 20 месяцев, город пал в июне 1611 года, а Шеин попал в польский плен.",
				ImageKey:    "shein.jpg",
				VideoKey:    "shein.mp4",
				OfficeStart: 1607,
				OfficeEnd:   1611,
				Status:      StatusPublished,
				LikedBy:     []int{1, 5},
			},
			{
				ID:          6,
				Name:        "Кара-Мустафа-паша",
				Office:      "Великий визирь Османской империи",
				Description: "Возглавил вторую осаду Вены в 1683 году. После поражения от войск Яна III Собеского был казнён в Белграде по приказу Мехмеда IV.",
				ImageKey:    "kara_mustafa.jpg",
				VideoKey:    "kara_mustafa.mp4",
				OfficeStart: 1676,
				OfficeEnd:   1683,
				Status:      StatusPublished,
				LikedBy:     []int{},
			},
			{
				ID:          7,
				Name:        "Рустем-паша",
				Office:      "Великий визирь Османской империи",
				Description: "Зять Сулеймана Великолепного, муж Михримах-султан. Известен финансовыми реформами и мечетью Рустема-паши в Стамбуле.",
				ImageKey:    "rustem_pasha.jpg",
				VideoKey:    "rustem_pasha.mp4",
				OfficeStart: 1544,
				OfficeEnd:   1553,
				Status:      StatusDeleted,
				LikedBy:     []int{2},
			},
			{
				ID:          8,
				Name:        "Гней Помпей Великий",
				Office:      "Консул Римской республики",
				Description: "Первое консульство Помпея совместно с Марком Лицинием Крассом. Восстановлены полномочия народных трибунов, урезанные Суллой.",
				ImageKey:    "pompey.jpg",
				VideoKey:    "pompey.mp4",
				OfficeStart: -70,
				OfficeEnd:   -70,
				Status:      StatusDraft,
				LikedBy:     []int{},
			},
		},
	}
}

// MediaURL возвращает ссылку на объект в Minio по его ключу
func (r *Repository) MediaURL(key string) string {
	return r.minioURL + "/" + key
}

// GetDignitaries возвращает опубликованных сановников; если year указан,
// то только тех, кто находился в должности в этом году
func (r *Repository) GetDignitaries(year *int) []Dignitary {
	var result []Dignitary
	for _, d := range r.dignitaries {
		if d.Status != StatusPublished {
			continue
		}
		if year != nil && (*year < d.OfficeStart || *year > d.OfficeEnd) {
			continue
		}
		result = append(result, d)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

// GetDignitary возвращает опубликованного сановника по ID
func (r *Repository) GetDignitary(id int) (Dignitary, error) {
	for _, d := range r.dignitaries {
		if d.ID == id && d.Status == StatusPublished {
			return d, nil
		}
	}
	return Dignitary{}, fmt.Errorf("сановник с id=%d не найден", id)
}

// GetNextDignitary возвращает следующего опубликованного сановника после указанного ID
// (после последнего — снова первого)
func (r *Repository) GetNextDignitary(id int) (Dignitary, error) {
	published := r.GetDignitaries(nil)
	if len(published) == 0 {
		return Dignitary{}, fmt.Errorf("нет опубликованных сановников")
	}
	for _, d := range published {
		if d.ID > id {
			return d, nil
		}
	}
	return published[0], nil
}

// GetFirstDignitary возвращает первого опубликованного сановника (лента без ID)
func (r *Repository) GetFirstDignitary() (Dignitary, error) {
	return r.GetNextDignitary(0)
}

// GetDraftDignitary возвращает сановника в статусе черновик
func (r *Repository) GetDraftDignitary() (Dignitary, error) {
	for _, d := range r.dignitaries {
		if d.Status == StatusDraft {
			return d, nil
		}
	}
	return Dignitary{}, fmt.Errorf("черновик не найден")
}
