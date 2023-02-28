create table if not exists playlist
(
    id            serial primary key,
    description   varchar(255) NOT NULL,
    duration      int default 0,
    prev          int default 0,
    next          int default 0
);
create index ix_prev on playlist (prev);
create index ix_next on playlist (next);

insert into public.playlist (description, duration, prev, next)
values  ('Гимн шута', 10, 0, 2),
        ('Проклятый старый дом', 12, 1, 3),
        ('Пират', 14, 2, 4),
        ('Скотный двор', 16, 3, 5),
        ('Кто это всё придумал?', 18, 4, 0);