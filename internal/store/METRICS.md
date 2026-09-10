# Cálculos y períodos

Esta API calcula resultados en memoria. No cambia el JSON ni la TUI.

- Día, semana (lunes a domingo), mes y año son períodos de calendario.
- La zona horaria es la de la fecha pasada a NewPeriod; la TUI debe usar hora local.
- Los intervalos incluyen el inicio y excluyen el fin.
- Shift avanza o retrocede períodos completos sin usar duraciones de 24 horas.
- FocusTime suma solo la parte de cada sesión que coincide con el período,
  hasta el instante now. Sessions cuenta las sesiones con duración positiva
  dentro del intervalo; una sesión que cruza medianoche puede contar en ambos
  días, pero su tiempo no se duplica. Tasks contiene el desglose por tarea.
- CompletedTodos cuenta fechas de finalización dentro del período, hasta now.
- HabitOpportunities suma días elegibles por hábito: desde el día de creación
  hasta hoy inclusive, intersectados con el período. No incluye días futuros.
- HabitCompletions cuenta marcas únicas en esos días. Las marcas son fechas
  civiles locales del JSON; no se convierten como instantes UTC.
- HabitPercent es completados / oportunidades × 100; devuelve cero cuando
  no hay oportunidades. La interfaz puede mostrar “sin datos” en ese caso.

Los resultados reflejan los registros actuales. Reabrir un pendiente elimina
su fecha de finalización; borrar pendientes, hábitos o sesiones elimina su
contribución histórica. No existe un registro de eventos para reconstruirla.
Hoy participa en el porcentaje aunque el día todavía no haya terminado.
