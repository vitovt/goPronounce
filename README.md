# goPronounce - Audio-Pronounce Trainer
* **goPronounce** is a lightweight desktop trainer for sharpening your foreign-language pronunciation.
* **goPronounce** — це невеличкий настільний застосунок-тренажер, який допомагає покращувати вимову іноземними мовами через метод «слухай → повторюй → порівнюй».
* **goPronounce** ist ein schlanker Desktop-Trainer, mit dem du deine Aussprache in Fremdsprachen verbessern kannst. Lade eine Referenz-Audiodatei, wähle einen kurzen Abschnitt, höre dem Muttersprachler zu, nimm deine Version auf und vergleiche beide Aufnahmen, bis deine Aussprache passt.
![ksnip_20250619-140649](https://github.com/user-attachments/assets/b86ede2d-5664-4491-ac7f-9f5d00eef1a7)

**Jump to:** [English](#english) • **Перейти до:** [Українська](#українська) • **Springe zu:** [Deutsch](#deutsch)

---

## English

### goPronounce – Audio-Pronounce Trainer

**goPronounce** is a lightweight desktop trainer for sharpening foreign-language pronunciation.

#### Key features

* **Load reference audio** — open any *wav/mp3/ogg/flac* file with clear native speech.
* **Select a fragment** — set start and end (sliders or `MM:SS` fields) to focus on one–two sentences.
* **Listen to the original** — catch every nuance of pronunciation.
* **Record your take** — press **Record**, speak, then stop.
* **Instant comparison** — play the original and your recording as many times as needed until they sound alike.

#### Quick user guide

1. Click **Select File** and choose the reference audio.
2. Wait for the duration to appear, then set the desired range.
3. Press **▶ Play Reference** and listen to the fragment.
4. Press **🎤 Record**, say the phrase, then **⏹ Stop**.
5. Compare with **Play Reference** and **Play Recording**.
6. Repeat recording until you are happy with your pronunciation.

#### Technologies

Written in Go using the Fyne GUI framework. Recording and playback rely on external tools (ffmpeg, ffplay, etc.).

#### Requirements

| Component | Version / note                                                |
| --------- | ------------------------------------------------------------- |
| Go        | ≥ 1.22                                                        |
| Fyne      | ≥ 2.6                                                         |
| FFmpeg    | Needs `ffprobe`, `ffplay`, `ffmpeg` (or `afplay` / `arecord`) |
| OS        | Linux, Windows, macOS                                         |

#### How to run

**Option 1:** download ready-made binaries from the *Releases* section on GitHub.

**Option 2:** build manually

```bash
git clone https://github.com/vitovt/goPronounce.git
cd goPronounce
make build
```

> **Note:** The author has tested only on Linux so far. Contributions and suggestions are welcome.

#### Licence

Released under the **MIT [LICENSE](LICENSE)**.

---

## Українська

### goPronounce – Audio-Pronounce Trainer

**goPronounce** — легкий настільний тренажер для покращення вимови іноземними мовами.

#### Основні можливості

* **Завантаження референс-аудіо** — відкрийте будь-який файл *wav/mp3/ogg/flac* з чистою вимовою носія.
* **Вибір фрагмента** — задайте початок і кінець (повзунками або полями `MM:SS`), щоб працювати лише з 1–2 реченнями.
* **Прослуховування оригіналу** — переконайтесь, що чуєте всі нюанси вимови.
* **Запис власного варіанта** — натисніть **Record**, вимовте фразу, зупиніть запис.
* **Миттєве порівняння** — слухайте оригінал і свій запис стільки разів, скільки потрібно, доки результат не стане максимально схожим.

#### Коротка інструкція користувача

1. Натисніть **Select File** і оберіть референс-аудіо.
2. Дочекайтесь відображення тривалості, встановіть потрібний діапазон.
3. Натисніть **▶ Play Reference** і прослухайте фрагмент.
4. Натисніть **🎤 Record**, промовте фразу, потім **⏹ Stop**.
5. Порівняйте звучання кнопками **Play Reference** та **Play Recording**.
6. Повторюйте запис, доки не будете задоволені своєю вимовою.

#### Технології

Написаний мовою Go з використанням GUI-фреймворку Fyne. Для запису та відтворення використовує зовнішні утиліти (ffmpeg, ffplay тощо).

#### Вимоги

| Компонент | Версія / примітка                                                 |
| --------- | ----------------------------------------------------------------- |
| Go        | ≥ 1.22                                                            |
| Fyne      | ≥ 2.6                                                             |
| FFmpeg    | Потрібні `ffprobe`, `ffplay`, `ffmpeg` (або `afplay` / `arecord`) |
| ОС        | Linux, Windows, macOS                                             |

#### Як запустити

**Варіант 1:** завантажити готові бінарні файли з розділу *Releases* на GitHub.

**Варіант 2:** скомпілювати вручну

```bash
git clone https://github.com/vitovt/goPronounce.git
cd goPronounce
make build
```

> **Увага:** автор тестував лише під Linux. Доробки та побажання вітаються.

#### Ліцензія

Цей проєкт поширюється на умовах **MIT [LICENSE](LICENSE)**.

---

## Deutsch

### goPronounce – Audio-Pronounce Trainer

**goPronounce** ist ein schlanker Desktop-Trainer zur Verbesserung der Aussprache in Fremdsprachen.

#### Hauptfunktionen

* **Referenz-Audio laden** — öffne jede *wav/mp3/ogg/flac*-Datei mit klarer Muttersprachler-Aussprache.
* **Ausschnitt wählen** — Start und Ende (Schieberegler oder `MM:SS`-Felder) festlegen, um ein bis zwei Sätze zu üben.
* **Original anhören** — alle Nuancen der Aussprache wahrnehmen.
* **Eigene Version aufnehmen** — **Record** drücken, Phrase sprechen, Aufnahme stoppen.
* **Direkter Vergleich** — Original und Aufnahme beliebig oft abspielen, bis beide nahezu identisch klingen.

#### Kurzanleitung

1. **Select File** anklicken und Referenz-Audio auswählen.
2. Warten, bis die Gesamtlänge erscheint, dann gewünschten Bereich einstellen.
3. **▶ Play Reference** drücken und aufmerksam zuhören.
4. **🎤 Record** drücken, Satz sprechen, dann **⏹ Stop**.
5. Mit **Play Reference** und **Play Recording** vergleichen.
6. Aufnahme wiederholen, bis die Aussprache überzeugt.

#### Technologien

Geschrieben in Go mit dem Fyne-GUI-Framework. Für Aufnahme und Wiedergabe werden externe Tools (ffmpeg, ffplay usw.) verwendet.

#### Anforderungen

| Komponente     | Version / Hinweis                                                  |
| -------------- | ------------------------------------------------------------------ |
| Go             | ≥ 1.22                                                             |
| Fyne           | ≥ 2.6                                                              |
| FFmpeg         | Benötigt `ffprobe`, `ffplay`, `ffmpeg` (oder `afplay` / `arecord`) |
| Betriebssystem | Linux, Windows, macOS                                              |

#### Ausführen

**Variante 1:** Vorgefertigte Binärdateien aus dem *Releases*-Bereich von GitHub herunterladen.

**Variante 2:** selbst kompilieren

```bash
git clone https://github.com/vitovt/goPronounce.git
cd goPronounce
make build
```

> **Achtung:** Der Autor hat bislang nur unter Linux getestet. Beiträge und Feedback sind willkommen!

#### Lizenz

Veröffentlicht unter der **MIT-[LICENSE](LICENSE)**.

