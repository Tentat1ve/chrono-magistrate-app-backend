-- Начальные данные. Выполнить в Adminer: «SQL-запрос» → вставить файл → «Выполнить»
-- (после go run ./cmd/migrate)

INSERT INTO users (id, login) VALUES
    (1, 'tentative'),
    (2, 'historian'),
    (3, 'chronicler'),
    (4, 'archivist'),
    (5, 'student');
SELECT setval('users_id_seq', (SELECT max(id) FROM users));

INSERT INTO dignitarie_lists (id, name, description, status, image_url, video_url,
                              start_date, end_date, created_at, published_at, creator_id) VALUES
    (1, 'Ибрагим-паша Паргалы',
     'Великий визирь Османской империи, друг детства и зять Сулеймана Великолепного. Руководил походами на Венгрию: при нём состоялась битва при Мохаче (1526) и первая осада Вены (1529). Казнён по приказу султана.',
     'опубликован', 'http://localhost:9000/dignitaries/ibrahim_pasha.jpg', 'http://localhost:9000/dignitaries/ibrahim_pasha.mp4',
     '1523-06-27', '1536-03-15', now(), now(), 1),
    (2, 'Мехмед-паша Соколлу',
     'Великий визирь Османской империи при трёх султанах: Сулеймане I, Селиме II и Мураде III. При нём взят Сигетвар (1566), заключён Адрианопольский мир с Габсбургами (1568) и проиграна битва при Лепанто (1571).',
     'опубликован', 'http://localhost:9000/dignitaries/sokollu_mehmed.jpg', 'http://localhost:9000/dignitaries/sokollu_mehmed.mp4',
     '1565-06-28', '1579-10-11', now(), now(), 1),
    (3, 'Марк Ульпий Траян',
     'Консул Римской империи 100 года вместе с Секстом Юлием Фронтином. Римский император 98–117 годов: при нём империя достигла наибольших размеров, были завоёваны Дакия и Парфия.',
     'опубликован', 'http://localhost:9000/dignitaries/trajan.jpg', 'http://localhost:9000/dignitaries/trajan.mp4',
     '0100-01-01', '0100-12-31', now(), now(), 1),
    (4, 'Марк Аврелий',
     'Консул Римской империи 140 года вместе с Антонином Пием. Будущий император (161–180), философ-стоик, автор сочинения «К самому себе».',
     'опубликован', 'http://localhost:9000/dignitaries/marcus_aurelius.jpg', 'http://localhost:9000/dignitaries/marcus_aurelius.mp4',
     '0140-01-01', '0140-12-31', now(), now(), 1),
    (5, 'Михаил Борисович Шеин',
     'Воевода Смоленска. Руководил обороной города от войск Сигизмунда III в 1609–1611 годах. Осада длилась 20 месяцев, город пал в июне 1611 года, а Шеин попал в польский плен.',
     'опубликован', 'http://localhost:9000/dignitaries/shein.jpg', 'http://localhost:9000/dignitaries/shein.mp4',
     '1609-09-16', '1611-06-03', now(), now(), 1),
    (6, 'Кара-Мустафа-паша',
     'Великий визирь Османской империи. Возглавил вторую осаду Вены в 1683 году. После поражения от войск Яна III Собеского казнён в Белграде по приказу султана Мехмеда IV.',
     'опубликован', 'http://localhost:9000/dignitaries/kara_mustafa.jpg', 'http://localhost:9000/dignitaries/kara_mustafa.mp4',
     '1676-11-01', '1683-12-25', now(), now(), 1),
    (7, 'Рустем-паша',
     'Великий визирь Османской империи, зять Сулеймана Великолепного, муж Михримах-султан. Известен финансовыми реформами и мечетью Рустема-паши в Стамбуле.',
     'удален', 'http://localhost:9000/dignitaries/rustem_pasha.jpg', 'http://localhost:9000/dignitaries/rustem_pasha.mp4',
     '1544-10-07', '1553-10-06', now(), now(), 1),
    (8, 'Публий Элий Адриан',
     'Консул Римской империи 118 года вместе с Гнеем Педанием Фуском Салинатором. Римский император 117–138 годов, построивший Адрианов вал в Британии.',
     'черновик', 'http://localhost:9000/dignitaries/hadrian.jpg', 'http://localhost:9000/dignitaries/hadrian.mp4',
     '0118-01-01', '0118-12-31', now(), NULL, 1);
SELECT setval('dignitarie_lists_id_seq', (SELECT max(id) FROM dignitarie_lists));

INSERT INTO dignitarie_list_likes (user_id, dignitarie_list_id) VALUES
    (1, 1), (2, 1), (3, 1),
    (2, 2), (4, 2),
    (1, 3), (2, 3), (3, 3), (4, 3), (5, 3),
    (3, 4),
    (1, 5), (5, 5),
    (2, 7);
