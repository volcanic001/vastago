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
- `1` a `5`: inicio, sesiones, pendientes, habitos y metricas.
- `tab` y `shift+tab`: cambiar de pantalla.

En Sesiones:

- `j/k` o flechas arriba/abajo: recorrer todo el historial.
- `enter` o `e`: editar tarea, nota, inicio y fin.
- `tab` y `shift+tab`: elegir campo; `enter` avanza y guarda desde el ultimo.
- Fechas en hora local, formato `AAAA-MM-DD HH:MM:SS`.
- `ctrl+u`: limpiar el campo; `esc`: cancelar sin guardar.
- `d`: solicitar borrado de la sesion seleccionada, incluso si esta activa.
- `y`: confirmar borrado; `n` o `esc`: cancelar.

Las sesiones activas conservan su estado al editar: se terminan con `x`.
Los horarios sin editar conservan su precision original.

Atajos generales:
- `r`: recargar los datos.
- `?`: ampliar la ayuda.
- `q`: salir.

## Metricas en la TUI

Pulsa `5` para abrir las metricas (semana actual por defecto).

- `d`, `w`, `m`, `y`: dia, semana, mes o año.
- `[` y `]`: periodo anterior o siguiente.
- `t`: volver al periodo actual; sigue el reloj automaticamente.
- `j/k` o flechas arriba/abajo: desplazar resultados.
- `tab` o `1–5`: cambiar de pantalla; `?`: ayuda.

Se muestran tiempo enfocado, sesiones, pendientes completados y cumplimiento
de habitos. La comparativa contrasta el periodo elegido con el anterior equivalente mediante lineas de actividad. Las barras representan la proporcion del tiempo total por tarea.
Las semanas comienzan el lunes. Hoy participa en el cumplimiento de habitos;
sin oportunidades se muestra «sin datos». Los periodos futuros muestran cero.

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
