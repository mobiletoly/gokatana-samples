package repo

const selectContactByIdSql =
/*language=sql*/ `
SELECT
   id, first_name, last_name
FROM sample.contact
WHERE id = @id
LIMIT 1
`

const selectAllContactsSql =
/*language=sql*/ `
SELECT
   id, first_name, last_name
FROM sample.contact
ORDER BY id
`

const insertContactSql =
/*language=sql*/ `
INSERT INTO sample.contact (first_name, last_name)
VALUES (@first_name, @last_name)
RETURNING id
`
