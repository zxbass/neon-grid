# Миссия 083 — PE-мясник

> Уровень: `★★★★☆` · Тема: PE/EXE, секции

PE-мясник: файл формата Windows, который нужно разделать на заголовки, секции и импорты. Джез учится
работать с чужим мясом аккуратно.
— Не режь вслепую, — Прах наблюдает за Джез. — Сначала прочитай, где кости. Потом режь.
Джез читает и режет аккуратно.

---

## Часть 1 — DOS и PE-заголовок

MZ-заголовок: `e_magic = "MZ"`, `e_lfanew` (u32 LE на смещении 0x3C)
указывает на `PE\0\0`. Оттуда:

| Смещение (от PE) | Поле |
| -------- | ------ |
| 4 | machine (0x14C = x86) |
| 6 | число секций |
| 14 | size of optional header |
| 16 | характеристики |

Выведи:

```text
PE32 x86  sections=5  entry=0x401000  subsys=GUI
```

(optional header: `magic` 0x10B = PE32, `AddressOfEntryPoint` на 0x10,
`Subsystem` на 0x68 — 2 = GUI, 3 = CUI.)

## Часть 2 — Секции и поиск

Секции идут после optional header: имя (8 байт), virtual size (u32),
virtual address (u32), size raw (u32), offset raw (u32).

Найди секцию `.rdata` и просканируй её байты на печатные строки длиной
≥ 4. Среди них есть пароль от архива — строка вида `pw_*` или
`pass*`:

```text
.rdata strings:
  pw_let_the_raven_out
  pass_is_not_here
  OMEGA-DYNE
FOUND: pw_let_the_raven_out
```

## Подсказки

- `e_lfanew` обычно 0x80, но не полагайся — читай.
- optional header размер указывает, где начинаются секции:
  `pe_off + 24 + size_optional`.
- Raw offset → в файле, VA → в памяти; для чтения файла используй
  raw offset.
- Строки-кандидаты: печатные ASCII, 4+ подряд.

(End of file - total 56 lines)
