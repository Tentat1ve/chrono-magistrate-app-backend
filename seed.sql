-- Начальные данные. Выполнить в Adminer: «SQL-запрос» → вставить файл → «Выполнить»
-- (после go run ./cmd/migrate)

INSERT INTO users (id, login, full_name, created_at) VALUES
    (1, 'tentative', 'Тентатив (создатель)', now()),
    (2, 'historian', 'Историк', now()),
    (3, 'chronicler', 'Летописец', now()),
    (4, 'archivist', 'Архивариус', now()),
    (5, 'student', 'Студент', now());
SELECT setval('users_id_seq', (SELECT max(id) FROM users));

INSERT INTO dignitaries (id, name, office, description, status, image_url, video_url,
                         office_start, office_end, created_at, published_at, creator_id) VALUES
    (1, 'Ибрагим-паша Паргалы', 'Великий визирь Османской империи',
     'Друг детства и зять Сулеймана Великолепного. Руководил походами на Венгрию, при нём состоялась битва при Мохаче (1526) и первая осада Вены (1529). Казнён по приказу султана в 1536 году.',
     'опубликован', 'http://localhost:9000/dignitaries/ibrahim_pasha.jpg', 'http://localhost:9000/dignitaries/ibrahim_pasha.mp4',
     1523, 1536, now(), now(), 1),
    (2, 'Мехмед-паша Соколлу', 'Великий визирь Османской империи',
     'Служил трём султанам: Сулейману I, Селиму II и Мураду III. При нём взят Сигетвар (1566), заключён Адрианопольский мир с Габсбургами (1568) и проиграна битва при Лепанто (1571).',
     'опубликован', 'http://localhost:9000/dignitaries/sokollu_mehmed.jpg', 'http://localhost:9000/dignitaries/sokollu_mehmed.mp4',
     1565, 1579, now(), now(), 1),
    (3, 'Марк Туллий Цицерон', 'Консул Римской республики',
     'Оратор и политик. В год его консульства был раскрыт заговор Катилины, а сам Цицерон получил титул «Отец Отечества». Коллега по консульству — Гай Антоний Гибрида.',
     'опубликован', 'http://localhost:9000/dignitaries/cicero.jpg', 'http://localhost:9000/dignitaries/cicero.mp4',
     -63, -63, now(), now(), 1),
    (4, 'Гай Юлий Цезарь', 'Консул Римской республики',
     'Первое консульство Цезаря вместе с Марком Кальпурнием Бибулом. В этот год был принят аграрный закон, а Цезарь получил наместничество в Галлии.',
     'опубликован', 'http://localhost:9000/dignitaries/caesar.jpg', 'http://localhost:9000/dignitaries/caesar.mp4',
     -59, -59, now(), now(), 1),
    (5, 'Михаил Борисович Шеин', 'Воевода Смоленска',
     'Руководил обороной Смоленска от войск Сигизмунда III в 1609–1611 годах. Осада длилась 20 месяцев, город пал в июне 1611 года, а Шеин попал в польский плен.',
     'опубликован', 'http://localhost:9000/dignitaries/shein.jpg', 'http://localhost:9000/dignitaries/shein.mp4',
     1607, 1611, now(), now(), 1),
    (6, 'Кара-Мустафа-паша', 'Великий визирь Османской империи',
     'Возглавил вторую осаду Вены в 1683 году. После поражения от войск Яна III Собеского был казнён в Белграде по приказу Мехмеда IV.',
     'опубликован', 'http://localhost:9000/dignitaries/kara_mustafa.jpg', 'http://localhost:9000/dignitaries/kara_mustafa.mp4',
     1676, 1683, now(), now(), 1),
    (7, 'Рустем-паша', 'Великий визирь Османской империи',
     'Зять Сулеймана Великолепного, муж Михримах-султан. Известен финансовыми реформами и мечетью Рустема-паши в Стамбуле.',
     'удален', 'http://localhost:9000/dignitaries/rustem_pasha.jpg', 'http://localhost:9000/dignitaries/rustem_pasha.mp4',
     1544, 1553, now(), now(), 1),
    (8, 'Гней Помпей Великий', 'Консул Римской республики',
     'Первое консульство Помпея совместно с Марком Лицинием Крассом. Восстановлены полномочия народных трибунов, урезанные Суллой.',
     'черновик', 'http://localhost:9000/dignitaries/pompey.jpg', 'http://localhost:9000/dignitaries/pompey.mp4',
     -70, -70, now(), NULL, 1);
SELECT setval('dignitaries_id_seq', (SELECT max(id) FROM dignitaries));

INSERT INTO dignitary_likes (user_id, dignitary_id) VALUES
    (1, 1), (2, 1), (3, 1),
    (2, 2), (4, 2),
    (1, 3), (2, 3), (3, 3), (4, 3), (5, 3),
    (3, 4),
    (1, 5), (5, 5),
    (2, 7);
