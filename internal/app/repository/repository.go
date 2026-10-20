package repository

// Repository хранит коллекцию услуг «списки сановников» (без БД, ЛР1)
type Repository struct {
	minioURL        string
	dignitarieLists []DignitarieList
}

func NewRepository(minioURL string) *Repository {
	return &Repository{
		minioURL:        minioURL,
		dignitarieLists: initialDignitarieLists(),
	}
}

// MediaURL возвращает ссылку на объект в Minio по его ключу
func (r *Repository) MediaURL(key string) string {
	return r.minioURL + "/" + key
}
