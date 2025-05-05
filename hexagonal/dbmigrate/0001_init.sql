CREATE SCHEMA IF NOT EXISTS sample;

CREATE TABLE IF NOT EXISTS sample.contact(
    id SERIAL PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    CONSTRAINT contact_first_name_last_name_key UNIQUE (first_name, last_name)
);
