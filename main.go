package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	a := app.New()
	w := a.NewWindow("Prot Tools - Испытание ТТ")

	entryRatio := widget.NewEntry()
	entryRatio.SetPlaceHolder("Например: 3000/5")

	entryClass := widget.NewEntry()
	entryClass.SetPlaceHolder("Например: 0.5")

	entryCurrent := widget.NewEntry()
	entryCurrent.SetPlaceHolder("Например: 100")

	resultLabel := widget.NewLabel("")

	btnGenerate := widget.NewButton("Сгенерировать", func() {
		ratio := entryRatio.Text
		classStr := entryClass.Text
		currentStr := entryCurrent.Text

		// Разбор коэффициента
		parts := strings.Split(ratio, "/")
		if len(parts) != 2 {
			resultLabel.SetText("Неверный формат коэффициента!")
			return
		}
		primNom, err1 := strconv.ParseFloat(parts[0], 64)
		secNom, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 != nil || err2 != nil {
			resultLabel.SetText("Введите числа в коэффициент!")
			return
		}

		// Разбор класса точности
		class, err3 := strconv.ParseFloat(classStr, 64)
		if err3 != nil || class <= 0 {
			resultLabel.SetText("Неверный класс точности!")
			return
		}

		// Разбор базового тока
		baseCurrent, err4 := strconv.ParseFloat(currentStr, 64)
		if err4 != nil || baseCurrent <= 0 {
			resultLabel.SetText("Неверный ток!")
			return
		}

		// Коэффициент трансформации
		ratioVal := primNom / secNom

		// Случайный ток в диапазоне ±5% от заданного
		currentDev := (rand.Float64()*2 - 1) * 5.0
		appliedCurrent := baseCurrent * (1 + currentDev/100.0)

		// Идеальный вторичный ток
		idealSecondary := appliedCurrent / ratioVal

		// Погрешность: от 0.2% до класса точности, но не выше 0.45%
		minDev := 0.2
		maxDev := class
		if maxDev > 0.45 {
			maxDev = 0.45
		}
		dev := minDev + rand.Float64()*(maxDev-minDev)
		secondary := idealSecondary * (1 - dev/100.0)

		// Фактическая погрешность
		actualDev := ((idealSecondary - secondary) / idealSecondary) * 100

		resultLabel.SetText(fmt.Sprintf(
			"Паспорт: %s (коэфф: %.0f)\n"+
				"Класс точности: %s\n"+
				"Подано на первичную: %.2f A\n"+
				"Измерено на вторичной: %.4f A\n"+
				"Погрешность: %.4f %%",
			ratio, ratioVal, classStr, appliedCurrent, secondary, actualDev,
		))
	})

	w.SetContent(container.NewVBox(
		widget.NewLabel("Паспортный коэффициент трансформации:"),
		entryRatio,
		widget.NewLabel("Класс точности (верхняя граница, максимум 0.45%):"),
		entryClass,
		widget.NewLabel("Базовый ток (A), диапазон ±5%:"),
		entryCurrent,
		btnGenerate,
		resultLabel,
	))
	w.Resize(fyne.NewSize(450, 400))
	w.ShowAndRun()
}
