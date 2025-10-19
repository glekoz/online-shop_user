-- name: CreateUser :exec
INSERT INTO users(id, name, email, password)
VALUES ($1, $2, $3, $4);

-- name: PromoteModer :exec
INSERT INTO moders(id)
VALUES($1);

-- при назначении админом не забыть в транзакции дать права модератора
-- name: PromoteAdmin :exec
INSERT INTO admins(id, isCore)
VALUES($1, FALSE);

-- name: PromoteCoreAdmin :execrows
UPDATE admins
SET isCore = TRUE
WHERE id = $1;

-- name: GetUserByID :one
SELECT * 
FROM users
WHERE id = $1;

-- используется при логине (инфа добавляется в токен), поэтому 
-- нужна дополнительная информация о правах (модер, админ, isCore),
-- чтобы при каждом GET запросе не идти в БД
-- name: GetUserByEmail :one
SELECT users.id, users.name, users.password,
    CASE WHEN moders.id IS NOT NULL THEN TRUE ELSE FALSE END AS is_moder, 
    CASE WHEN admins.id IS NOT NULL THEN TRUE ELSE FALSE END AS is_admin,
    CASE WHEN admins.iscore IS NOT NULL THEN admins.iscore ELSE FALSE END AS is_core
FROM users
    LEFT JOIN moders ON users.id = moders.id
    LEFT JOIN admins ON users.id = admins.id
WHERE users.email = $1;

-- этот метод вызывается только администратором,
-- поэтому нужна полная инфоормация о правах (модератор, админ, isCore),
-- чтобы отобразить её в интерфейсе управления пользователями
-- name: GetUsersByEmail :many
SELECT users.id, users.name, users.email, users.email_confirmed,
	CASE WHEN moders.id IS NOT NULL THEN TRUE ELSE FALSE END AS is_moder, 
	CASE WHEN admins.id IS NOT NULL THEN TRUE ELSE FALSE END AS is_admin,
	CASE WHEN admins.iscore IS NOT NULL THEN admins.iscore ELSE FALSE END AS is_core
FROM users 
	LEFT JOIN moders ON users.id = moders.id
	LEFT JOIN admins ON users.id = admins.id
WHERE users.email LIKE $1;

-- name: GetModer :one
SELECT * 
FROM moders
WHERE id = $1;

-- name: GetAdmin :one
SELECT * 
FROM admins
WHERE id = $1;

-- name: ConfirmEmail :execrows
UPDATE users
SET email_confirmed = TRUE
WHERE id = $1;

-- впоследствии этот метод надо расширить на день рождения и адрес
-- name: ChangeName :execrows
UPDATE users
SET name = $1
WHERE id = $2;

-- асинхронно с подтверждением через почту (ссылка на изменение пароля так же отправляется на почту, и на странице по этой ссылке можно сменить пароль)
-- name: ChangePassword :execrows 
UPDATE users
SET password=$1
WHERE id = $2;

-- асинхронно и не обновлять, пока новая почта не будет подтверждена
-- name: ChangeEmail :execrows 
UPDATE users
SET email = $1
WHERE id = $2;

-- нужно проверять, чтобы было право администратора
-- name: DeleteUser :execrows
DELETE FROM users
WHERE id = $1;

-- нужно проверять, чтобы права на модерацию не убрали у админа
-- name: DeleteModer :execrows
DELETE FROM moders
WHERE id = $1;

-- нужно проверять, чтобы было isCore 
-- name: DeleteAdmin :execrows
DELETE FROM admins
WHERE id = $1;