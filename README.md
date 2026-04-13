## kontarawa_agent (Go CLI)

Локальный агент с долговременной памятью на файлах.

### Идея
- **Модель** работает локально (Ollama / позже можно llama.cpp server).
- **Память** живёт на диске (`memory/`) и подмешивается в контекст через retrieval.
- **Обучение тобой** = команда `learn` сохраняет уроки (плохой/хороший ответ + почему).

### Структура памяти
- `memory/profile.md` — твои постоянные правила и стиль.
- `memory/knowledge/` — заметки/решения/сниппеты (markdown).
- `memory/lessons/` — уроки из правок (markdown).

### Переменные окружения
Файл `env.example` лежит в репозитории **только как пример** и автоматически не используется.
Пример `.env`, если память лежит в отдельном приватном репозитории:
- `MEMORY_DIR`: путь до папки памяти (относительный или абсолютный). По умолчанию `memory`.
- `OLLAMA_HOST`: адрес Ollama. По умолчанию `http://localhost:11434`.

### Быстрый старт (Ubuntu)
1) Ollama:
```bash
curl -fsSL https://ollama.com/install.sh | sh
ollama serve
ollama pull qwen2.5-coder:7b-instruct
```

2) Собрать CLI:
```bash
go build -o kontarawa ./cmd/kontarawa
```

3) Использование:
```bash
./kontarawa doctor
./kontarawa ask "Напиши функцию на Go для ..."
./kontarawa learn --prompt "..." --bad "..." --good "..." --why "..."
```

