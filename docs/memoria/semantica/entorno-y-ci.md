# Entorno y CI

Verificado: 2026-10-29 contra `43f3393`.

## Hechos

- La base de datos es PostgreSQL 17.6 (`postgres:17.6-alpine`), en el servicio `db` de Docker Compose, publicado en el puerto `5432` — `compose.yaml:2-9`
- Hay dos bases de datos, `todo_development` y `todo_test`, con usuario `postgres` y host `DB_HOST` (por defecto `127.0.0.1`) — `database.yml:1-16`
- La CI ejecuta, con `GO_ENV=test`: `gofmt -l .` (debe salir vacío), `go vet ./...`, `soda migrate -e test` (soda `v6.4.1`) y `go test -p 1 ./...` — `.github/workflows/ci.yaml:28`, `:39`, `:47`, `:51-55`
- Los tests de `actions` y `models` son métodos de suites de testify (`Test_ActionSuite`, `Test_ModelSuite`): `-run` filtra la suite y `-testify.m` filtra el método — `actions/actions_test.go:14`, `models/models_test.go:14`
- La app responde en `http://localhost:3000/healthz` — `actions/app.go:58`, `compose.yaml:29`

## Particularidades de Windows

- Con `core.autocrlf=true`, `gofmt -l .` lista casi todos los `.go` por los finales de línea CRLF, aunque estén bien formateados. Lo que ve la CI es el contenido del commit (`git show HEAD:<archivo> | gofmt -l`) — [episodio](../episodica/2026-10-26-fecha-limite-tasks.md)
- Los 6 tests de `todo/adws` fallan en Windows y pasan en Linux: el `gh` falso es un script bash sin extensión (`adws/trigger_test.go:55`) que se añade al principio del `PATH` (`adws/trigger_test.go:90`), pero se resuelve antes el `gh.exe` real. Es un fallo previo e independiente de los cambios de la aplicación
- `.claude/worktrees/` contiene copias de otras ramas y también sale en `gofmt -l .`; `git ls-files` lo deja fuera — `.claude/skills/validacion/SKILL.md`, sección 3

El procedimiento de validación completo, con estos atajos para Windows, está en la skill [`/validacion`](../../../.claude/skills/validacion/SKILL.md).

## Origen

- [2026-10-26 · Fecha límite opcional en las tareas](../episodica/2026-10-26-fecha-limite-tasks.md)
