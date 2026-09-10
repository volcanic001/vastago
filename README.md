# Vastago

Tiempo enfocado, crecimiento constante.

`vastago` es un rastreador de tiempo minimalista para la terminal, construido
con Bubble Tea y Lip Gloss. Esta inspirado en la claridad de `hours`, pero es un
proyecto independiente y pone especial atencion en pantallas estrechas como
Termux en Android.

## Que incluye el primer MVP

- TUI con panel activo, resumen de hoy y de los ultimos siete dias.
- Historial reciente y estadisticas agrupadas por tarea.
- Tres densidades automaticas: amplia, compacta y minima.
- Inicio y cierre de sesiones desde la TUI o desde comandos cortos.
- Almacenamiento local en JSON, sin servidor ni cuenta.
- Compatibilidad objetivo: Termux, Linux y macOS.

La aplicacion no cambia el zoom o tamano de fuente del emulador de terminal.
Lee el ancho y alto disponibles y reorganiza su contenido para aprovecharlos.

## Instalar

Se requiere Go 1.25 o posterior.

```sh
go install github.com/volcanic001/vastago@latest
```

Mientras el repositorio aun no se haya publicado, desde esta carpeta:

```sh
go install .
```

En Termux:

```sh
pkg update
pkg install golang git
go install github.com/volcanic001/vastago@latest
```

Asegura que el directorio de binarios de Go este en `PATH`:

```sh
export PATH="$PATH:$HOME/go/bin"
```

## Uso

Abre la interfaz:

```sh
vastago
```

Comandos rapidos:

```sh
vastago start "Trabajo profundo"
vastago status
vastago stop --note "Capitulo terminado"
vastago log --days 14
vastago stats --days 30
```

Teclas en la TUI:

- `n`: iniciar una sesion.
- `x`: terminar la sesion activa.
- `tab`: alternar inicio e historial.
- `r`: recargar los datos.
- `?`: ampliar la ayuda.
- `q`: salir.

## Datos

Por defecto se guardan en:

```text
~/.local/share/vastago/store.json
```

Puedes cambiar la ruta con `VASTAGO_DATA`, `XDG_DATA_HOME` o `--data`:

```sh
vastago --data ./mis-horas.json stats
```

## Desarrollo

```sh
go test ./...
go run .
```

Al publicar una nueva etiqueta `v*`, GitHub Actions solicita su indexacion al
proxy de Go. Puede haber un breve retraso mientras `@latest` se propaga por la
red de proxies.

## Licencia

MIT
