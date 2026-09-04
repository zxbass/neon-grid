# 193 — MQTT-шлюз

> Уровень: `★★★★★` · Тема: MQTT · IoT

MQTT-шлюз в умном квартале: устройства шепчутся темами и подписками. Джез слушает разговор
холодильников, датчиков и камер.
— Даже кофеварка говорит, — Дон смеётся. — Вопрос — кто её слушает.
Джез слушает кофеварку.

---

## Данные

- `data/193/mqtt.bin`

## Часть 1

Верни:

```text
CONNECT: flags=0x02 client=zen-agent keepalive=60
PUBLISH: topic=neon/alerts qos=0
SUBSCRIBE: topic=neon/#
```

## Часть 2

Верни `PAYLOAD: ICE_BREACH_LEVEL_3` — полезная нагрузка PUBLISH.
