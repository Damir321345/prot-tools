package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type VACPoint struct {
	Current float64
	Voltage float64
}

func main() {
	rand.Seed(time.Now().UnixNano())
	a := app.New()
	w := a.NewWindow("Prot Tools")

	// ==================== Вкладка 1: Коэф. тр. ====================
	entryRatio := widget.NewEntry()
	entryRatio.SetPlaceHolder("Например: 3000/5")

	entryClass := widget.NewEntry()
	entryClass.SetPlaceHolder("Например: 0.5")

	entryCurrent := widget.NewEntry()
	entryCurrent.SetPlaceHolder("Например: 100")

	entryCount := widget.NewEntry()
	entryCount.SetPlaceHolder("Например: 3")
	entryCount.SetText("1")

	var allValues string
	resultLabel := widget.NewLabel("Нажмите «Сгенерировать»")
	btnCopy := widget.NewButton("Копировать всё", nil)
	btnCopy.Disable()

	btnGenerate := widget.NewButton("Сгенерировать", func() {
		ratio := entryRatio.Text
		classStr := entryClass.Text
		currentStr := entryCurrent.Text
		countStr := entryCount.Text

		parts := strings.Split(ratio, "/")
		if len(parts) != 2 {
			return
		}
		primNom, err1 := strconv.ParseFloat(parts[0], 64)
		secNom, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 != nil || err2 != nil {
			return
		}

		class, err3 := strconv.ParseFloat(classStr, 64)
		if err3 != nil || class <= 0 {
			return
		}

		baseCurrent, err4 := strconv.ParseFloat(currentStr, 64)
		if err4 != nil || baseCurrent <= 0 {
			return
		}

		count, err5 := strconv.Atoi(countStr)
		if err5 != nil || count <= 0 {
			count = 1
		}
		if count > 20 {
			count = 20
		}

		ratioVal := primNom / secNom
		ratioStr := fmt.Sprintf("%.0f", ratioVal)

		minDev := 0.2
		maxDev := class
		if maxDev < minDev {
			maxDev = minDev
		}

		var preview strings.Builder
		preview.WriteString("I перв (A)  |  I втор (A)  |  Коэфф\n")
		preview.WriteString("-----------------------------------\n")

		firstCurrentDev := (rand.Float64()*2 - 1) * 5.0
		firstApplied := baseCurrent * (1 + firstCurrentDev/100.0)
		firstDev := minDev + rand.Float64()*(maxDev-minDev)
		firstSecondary := (firstApplied / ratioVal) * (1 - firstDev/100.0)

		firstPrimaryStr := fmt.Sprintf("%.2f", firstApplied)
		firstSecondaryStr := fmt.Sprintf("%.4f", firstSecondary)

		preview.WriteString(fmt.Sprintf("%-10s  |  %-10s  |  %s\n", firstPrimaryStr, firstSecondaryStr, ratioStr))

		var valuesBuilder strings.Builder
		valuesBuilder.WriteString(firstPrimaryStr)
		valuesBuilder.WriteString("\t")
		valuesBuilder.WriteString(firstSecondaryStr)
		valuesBuilder.WriteString("\t")
		valuesBuilder.WriteString(ratioStr)

		for i := 1; i < count; i++ {
			currentDev := (rand.Float64()*2 - 1) * 5.0
			appliedCurrent := baseCurrent * (1 + currentDev/100.0)
			idealSecondary := appliedCurrent / ratioVal

			dev := minDev + rand.Float64()*(maxDev-minDev)
			secondary := idealSecondary * (1 - dev/100.0)

			primaryStr := fmt.Sprintf("%.2f", appliedCurrent)
			secondaryStr := fmt.Sprintf("%.4f", secondary)

			preview.WriteString(fmt.Sprintf("%-10s  |  %-10s  |  %s\n", primaryStr, secondaryStr, ratioStr))
			valuesBuilder.WriteString("\n")
			valuesBuilder.WriteString(primaryStr)
			valuesBuilder.WriteString("\t")
			valuesBuilder.WriteString(secondaryStr)
			valuesBuilder.WriteString("\t")
			valuesBuilder.WriteString(ratioStr)
		}

		allValues = valuesBuilder.String()
		resultLabel.SetText(preview.String())
		btnCopy.Enable()
	})

	btnCopy.OnTapped = func() {
		w.Clipboard().SetContent(allValues)
		btnCopy.SetText("Скопировано ✓")
		go func() {
			time.Sleep(1 * time.Second)
			btnCopy.SetText("Копировать всё")
		}()
	}

	ttTab := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Паспортный коэффициент:"),
			entryRatio,
			widget.NewLabel("Класс точности (%):"),
			entryClass,
			widget.NewLabel("Базовый ток (A):"),
			entryCurrent,
			widget.NewLabel("Количество точек:"),
			entryCount,
			btnGenerate,
			widget.NewSeparator(),
			btnCopy,
		),
		nil, nil, nil,
		container.NewScroll(resultLabel),
	)

	// ==================== Вкладка 2: ВАХ ====================
	vacEntry := widget.NewMultiLineEntry()
	vacEntry.SetPlaceHolder(`Вставьте таблицу из Word.
Первая строка — заголовок (пропускается).
Данные: Напряжение (В) и Ток (А) через табуляцию.`)

	vacResultLabel := widget.NewLabel("")
	vacBtnCopy := widget.NewButton("Копировать напряжения", nil)
	vacBtnCopy.Disable()
	var vacResultValues string

	vacBtnCalc := widget.NewButton("Рассчитать", func() {
		text := vacEntry.Text
		lines := strings.Split(text, "\n")
		var points []VACPoint

		for i, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			line = strings.Replace(line, ",", ".", -1)
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			voltage, err1 := strconv.ParseFloat(fields[0], 64)
			current, err2 := strconv.ParseFloat(fields[1], 64)
			if err1 != nil || err2 != nil {
				if i > 0 {
					fmt.Println("Пропущена строка:", line)
				}
				continue
			}
			points = append(points, VACPoint{Current: current, Voltage: voltage})
		}

		if len(points) < 2 {
			vacResultLabel.SetText(fmt.Sprintf("Нужно минимум 2 точки! Распознано: %d", len(points)))
			return
		}

		sort.Slice(points, func(i, j int) bool {
			return points[i].Current < points[j].Current
		})

		targetCurrents := []float64{0.1, 0.25, 0.5, 1.0}
		var result strings.Builder
		result.WriteString("Результаты:\n")
		result.WriteString("Ток (A)  |  Напряжение (V)\n")
		result.WriteString("------------------------\n")

		var valuesForCopy []string

		for _, targetI := range targetCurrents {
			voltage := interpolate(points, targetI)
			if math.IsNaN(voltage) {
				result.WriteString(fmt.Sprintf("%.2f     |  — (вне диапазона)\n", targetI))
				valuesForCopy = append(valuesForCopy, "—")
			} else {
				result.WriteString(fmt.Sprintf("%.2f     |  %.3f\n", targetI, voltage))
				valuesForCopy = append(valuesForCopy, fmt.Sprintf("%.3f", voltage))
			}
		}

		vacResultValues = strings.Join(valuesForCopy, "\t")
		vacResultLabel.SetText(result.String())
		vacBtnCopy.Enable()
	})

	vacBtnCopy.OnTapped = func() {
		w.Clipboard().SetContent(vacResultValues)
		vacBtnCopy.SetText("Скопировано ✓")
		go func() {
			time.Sleep(1 * time.Second)
			vacBtnCopy.SetText("Копировать напряжения")
		}()
	}

	vacTab := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Вставьте таблицу ВАХ (Напряжение [В] и Ток [А]):"),
			vacEntry,
			vacBtnCalc,
			widget.NewSeparator(),
			vacBtnCopy,
		),
		nil, nil, nil,
		container.NewScroll(vacResultLabel),
	)

	// ==================== Вкладка 3: О программе ====================
	aboutTab := container.NewVBox(
		widget.NewLabel(""),
		widget.NewLabel("Prot Tools v1.0"),
		widget.NewLabel(""),
		widget.NewLabel("Обновления:"),
		widget.NewLabel(""),

		&widget.Button{
			Text:       "🌐 prot-tools.website.yandexcloud.net",
			Importance: widget.LowImportance,
			OnTapped: func() {
				w.Clipboard().SetContent("prot-tools.website.yandexcloud.net")
			},
		},
	)

	// ==================== Вкладки ====================
	tabs := container.NewAppTabs(
		container.NewTabItem("Коэф. тр.", ttTab),
		container.NewTabItem("ВАХ", vacTab),
		container.NewTabItem("О программе", aboutTab),
	)

	w.SetContent(tabs)
	w.Resize(fyne.NewSize(600, 600))
	w.ShowAndRun()
}

func interpolate(points []VACPoint, targetI float64) float64 {
	if targetI < points[0].Current || targetI > points[len(points)-1].Current {
		return math.NaN()
	}

	for i := 0; i < len(points)-1; i++ {
		if targetI >= points[i].Current && targetI <= points[i+1].Current {
			u1, i1 := points[i].Voltage, points[i].Current
			u2, i2 := points[i+1].Voltage, points[i+1].Current
			return u1 + (u2-u1)*(targetI-i1)/(i2-i1)
		}
	}

	return math.NaN()
}
