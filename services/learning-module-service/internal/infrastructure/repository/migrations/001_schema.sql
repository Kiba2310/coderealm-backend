CREATE TABLE IF NOT EXISTS courses (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    description TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS chapters (
    id         TEXT PRIMARY KEY,
    course_id  TEXT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    order_num  INT  NOT NULL,
    title      TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS lessons (
    id            TEXT PRIMARY KEY,
    chapter_id    TEXT NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    order_num     INT  NOT NULL,
    title         TEXT NOT NULL,
    description   TEXT NOT NULL,
    type          TEXT NOT NULL,
    language      TEXT,
    starter_code  TEXT,
    solution_code TEXT
);

CREATE TABLE IF NOT EXISTS answer_options (
    id         TEXT    PRIMARY KEY,
    lesson_id  TEXT    NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    text       TEXT    NOT NULL,
    is_correct BOOLEAN NOT NULL DEFAULT false
);
