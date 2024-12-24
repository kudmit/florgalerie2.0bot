package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const AdminID int64 = 246690184

// Информация о пользователе
type UserInfo struct {
	Language         string
	Bouquet          string
	OrderTime        string
	LastAdminMessage string
	UserName         string // Имя пользователя или "Анонимный пользователь"
}

// Отправка информации админу
func sendUpdatedInfoToAdmin(bot *tgbotapi.BotAPI, chatID int64, userInfo UserInfo) {
	// Формируем кликабельный ID пользователя
	clickableID := fmt.Sprintf("<a href=\"tg://user?id=%d\">%d</a>", chatID, chatID)

	// Формируем основное сообщение со всей информацией
	message := fmt.Sprintf(
		"👤 Пользователь: %s\n🌍 Язык: %s\n📝 Имя: %s\n💐 Букет: %s\n⏰ Время заказа: %s",
		clickableID,        // Кликабельный ID
		userInfo.Language,  // Язык пользователя
		userInfo.UserName,  // Имя пользователя
		userInfo.Bouquet,   // Букет
		userInfo.OrderTime, // Время заказа
	)

	// Отправляем основное сообщение админу
	msg := tgbotapi.NewMessage(AdminID, message)
	msg.ParseMode = "HTML" // Используем HTML для кликабельных ссылок
	bot.Send(msg)

	// Отправляем ID пользователя отдельно
	idMessage := fmt.Sprintf(" %d", chatID)
	bot.Send(tgbotapi.NewMessage(AdminID, idMessage))
}

// Пересылка сообщения от администратора пользователю
func handleAdminMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, userData map[int64]*UserInfo) {
	message := update.Message

	// Если администратор отправил фотографию
	if message.Photo != nil {
		parts := strings.SplitN(message.Caption, " ", 2) // Используем Caption для ID пользователя
		if len(parts) < 1 {
			bot.Send(tgbotapi.NewMessage(AdminID, "❗ Укажите ID пользователя в подписи к фото."))
			return
		}

		userIDStr := parts[0]
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(AdminID, "❗ Неверный формат ID пользователя в подписи."))
			return
		}

		photo := message.Photo[len(message.Photo)-1] // Берём самое большое фото
		photoMsg := tgbotapi.NewPhoto(userID, tgbotapi.FileID(photo.FileID))
		photoMsg.Caption = "📸 Фото от администратора"
		_, err = bot.Send(photoMsg)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(AdminID, fmt.Sprintf("❌ Не удалось отправить фото пользователю: %v", err)))
		} else {
			// Сохраняем сообщение администратора для пользователя
			if userInfo, exists := userData[userID]; exists {
				userInfo.LastAdminMessage = "📸 Фото от администратора"
			}
			bot.Send(tgbotapi.NewMessage(AdminID, "✅ Фото отправлено пользователю."))
		}
		return
	}

	// Обработка текстовых сообщений
	text := message.Text
	parts := strings.SplitN(text, " ", 2)
	if len(parts) < 2 {
		bot.Send(tgbotapi.NewMessage(AdminID, "❗ Укажите ID пользователя и текст сообщения через пробел."))
		return
	}

	userIDStr := parts[0]
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(AdminID, "❗ Неверный формат ID пользователя."))
		return
	}

	messageText := parts[1]
	if messageText == "" {
		bot.Send(tgbotapi.NewMessage(AdminID, "❗ Текст сообщения пустой."))
		return
	}

	msg := tgbotapi.NewMessage(userID, fmt.Sprintf("🔔 Admin:\n%s", messageText))
	_, err = bot.Send(msg)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(AdminID, fmt.Sprintf("❌ Не удалось отправить сообщение пользователю: %v", err)))
	} else {
		// Сохраняем текст последнего сообщения администратора
		if userInfo, exists := userData[userID]; exists {
			userInfo.LastAdminMessage = messageText
		}
		bot.Send(tgbotapi.NewMessage(AdminID, "✅ Сообщение отправлено пользователю."))
	}
}

// Приветствие
func sendGreeting(bot *tgbotapi.BotAPI, chatID int64, lang string) {
	var greeting string
	switch lang {
	case "DEU":
		greeting = "Willkommen in unserer Blumenboutique Florgalerie💐! Ich bin Max, ein Bot🤖, der Ihnen bei der Bestellung eines Blumenstraußes hilft, und unsere aufmerksamen Administratoren unterstützen Sie bei speziellen Anliegen! Wir verfolgen einen ausschließlich individuellen Ansatz für unsere Kunden, daher gibt es bei uns keine standardisierten oder vorgefertigten Sträuße."
	case "EN":
		greeting = "Welcome to our flower boutique Florgalerie💐! My name is Max, a bot🤖 designed to assist you with ordering a bouquet, and our attentive administrators will help you with any unique requests! We take an exclusively individual approach to our clients, which is why we don't offer standardized or pre-made bouquets."
	case "UK":
		greeting = "Вітаємо Вас у нашому квітковому бутіку Florgalerie💐! Мене звати Макс, я бот🤖, створений для допомоги Вам із замовленням букета, а наші чуйні адміністратори допоможуть Вам із нетиповими запитаннями! Ми дотримуємося виключно індивідуального підходу до клієнтів, тому у нас відсутні типові або вже готові букети."
	case "RU":
		greeting = "Приветствуем Вас в нашем цветочном бутике Florgalerie💐! Меня зовут Макс, я бот🤖, созданный для помощи Вам с заказом букета, а наши чуткие администраторы помогут Вам с нетиповыми вопросами! У нас исключительно индивидуальный подход к клиентам, поэтому отсутствуют типовые-уже скомпонованные букеты."
	}

	msg := tgbotapi.NewMessage(chatID, greeting)
	bot.Send(msg)
}

func askUserName(bot *tgbotapi.BotAPI, chatID int64, lang string) {
	var message, anonymousButton string
	switch lang {
	case "DEU":
		message = "Wie können wir Sie ansprechen?"
		anonymousButton = "Anonym bleiben"
	case "EN":
		message = "How should we address you?"
		anonymousButton = "Stay anonymous"
	case "UK":
		message = "Як ми можемо до Вас звертатися?"
		anonymousButton = "Залишитися анонімним"
	case "RU":
		message = "Как мы можем к Вам обращаться?"
		anonymousButton = "Остаться анонимным"
	}

	buttons := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(anonymousButton),
		),
	)
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ReplyMarkup = buttons
	bot.Send(msg)
}

// Запрос описания букета
func sendBouquetRequest(bot *tgbotapi.BotAPI, chatID int64, lang string) {
	var message string
	switch lang {
	case "DEU":
		message = "Beschreiben Sie bitte den gewünschten Blumenstrauß:"
	case "EN":
		message = "Please describe the bouquet you would like:"
	case "UK":
		message = "Опишіть, будь ласка, букет, який ви хотіли б:"
	case "RU":
		message = "Опишите, пожалуйста, букет, который вы хотели бы:"
	}
	msg := tgbotapi.NewMessage(chatID, message)
	bot.Send(msg)
}

// График работы магазина
func sendSchedule(bot *tgbotapi.BotAPI, chatID int64, lang string) {
	var schedule string
	switch lang {
	case "DEU":
		schedule = "Arbeitszeiten:\nMontag-Freitag: 9:00 - 21:00\nSamstag: 8:00 - 19:00\nSonntag: 9:00 - 15:00"
	case "EN":
		schedule = "Working hours:\nMonday-Friday: 9:00 - 21:00\nSaturday: 8:00 - 19:00\nSunday: 9:00 - 15:00"
	case "UK":
		schedule = "Графік роботи:\nПонеділок-П’ятниця: 9:00 - 21:00\nСубота: 8:00 - 19:00\nНеділя: 9:00 - 15:00"
	case "RU":
		schedule = "График работы:\nПонедельник-Пятница: 9:00 - 21:00\nСуббота: 8:00 - 19:00\nВоскресенье: 9:00 - 15:00"
	}
	msg := tgbotapi.NewMessage(chatID, schedule)
	bot.Send(msg)
}

// Запрос времени
func sendOrderTimeRequest(bot *tgbotapi.BotAPI, chatID int64, lang string) {
	var message string
	switch lang {
	case "DEU":
		message = "Bitte geben Sie das Datum und die Uhrzeit Ihrer Buchung im Format 'TT.MM.JJJJ HH:MM' ein. (z.B. 31.12.2024 15:30)."
	case "EN":
		message = "Please enter the date and time of your order in the format 'DD.MM.YYYY HH:MM' (e.g., 31.12.2024 15:30)."
	case "UK":
		message = "Будь ласка, введіть дату і час Вашого замовлення у форматі 'ДД.ММ.ГГГГ ЧЧ:ММ' (наприклад: 31.12.2024 15:30)."
	case "RU":
		message = "Пожалуйста, введите дату и время Вашего заказа в формате 'ДД.ММ.ГГГГ ЧЧ:ММ' (например: 31.12.2024 15:30)."
	}
	msg := tgbotapi.NewMessage(chatID, message)
	bot.Send(msg)
}

// Проверка рабочего времени
func isWithinWorkingHours(t time.Time) bool {
	weekday := t.Weekday()
	hour := t.Hour()
	switch weekday {
	case time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday:
		return hour >= 9 && hour < 21
	case time.Saturday:
		return hour >= 8 && hour < 19
	case time.Sunday:
		return hour >= 9 && hour < 15
	default:
		return false
	}
}

// Сообщение, если магазин закрыт
func sendStoreClosedOptions(bot *tgbotapi.BotAPI, chatID int64, lang string, selectedTime time.Time) {
	loc := selectedTime.Location()
	nextDay := selectedTime.AddDate(0, 0, 1)
	nextDayMorning := time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 9, 0, 0, 0, loc)

	var message, tryAgainButton, nextDayButton string
	switch lang {
	case "DEU":
		message = "Das Geschäft ist zu dieser Zeit geschlossen. Wählen Sie eine Option:"
		tryAgainButton = "Neue Zeit eingeben"
		nextDayButton = fmt.Sprintf("Am %02d.%02d um %02d:%02d erhalten", nextDayMorning.Day(), nextDayMorning.Month(), 9, 0)
	case "EN":
		message = "The store is closed at this time. Choose an option:"
		tryAgainButton = "Enter a new time"
		nextDayButton = fmt.Sprintf("Receive on %02d.%02d at %02d:%02d", nextDayMorning.Day(), nextDayMorning.Month(), 9, 0)
	case "UK":
		message = "Магазин зачинений у цей час. Оберіть опцію:"
		tryAgainButton = "Ввести новий час"
		nextDayButton = fmt.Sprintf("Отримати %02d.%02d о %02d:%02d", nextDayMorning.Day(), nextDayMorning.Month(), 9, 0)
	case "RU":
		message = "Магазин закрыт в это время. Выберите вариант:"
		tryAgainButton = "Ввести новое время"
		nextDayButton = fmt.Sprintf("Получить %02d.%02d в %02d:%02d", nextDayMorning.Day(), nextDayMorning.Month(), 9, 0)
	}

	buttons := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(tryAgainButton),
			tgbotapi.NewKeyboardButton(nextDayButton),
		),
	)
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ReplyMarkup = buttons
	bot.Send(msg)
}

// Сообщение при некорректном времени
func sendInvalidTimeMessage(bot *tgbotapi.BotAPI, chatID int64, lang string) {
	loc, _ := time.LoadLocation("Europe/Vienna")
	currentTime := time.Now().In(loc)

	var message string
	switch lang {
	case "DEU":
		message = fmt.Sprintf("Sie haben die Uhrzeit falsch eingegeben, bitte korrigieren Sie sie. Aktuelle Uhrzeit: %s. Bitte geben Sie die Zeit erneut ein.", currentTime.Format("02.01.2006 15:04"))
	case "EN":
		message = fmt.Sprintf("You have entered the time incorrectly, please correct it. Present time: %s. Please try again.", currentTime.Format("02.01.2006 15:04"))
	case "UK":
		message = fmt.Sprintf("Ви ввели час некоректно, будь ласка, виправте його. Теперішній час: %s. Спробуйте ще раз.", currentTime.Format("02.01.2006 15:04"))
	case "RU":
		message = fmt.Sprintf("Вы ввели время некорректно, пожалуйста исправьте его. Настоящее время: %s. Попробуйте снова.", currentTime.Format("02.01.2006 15:04"))
	}
	msg := tgbotapi.NewMessage(chatID, message)
	bot.Send(msg)
}

// Обработка кнопки "Следующий день"
func handleNextDaySelection(bot *tgbotapi.BotAPI, chatID int64, lang string, userInfo *UserInfo) {
	loc, _ := time.LoadLocation("Europe/Vienna")
	nextDay := time.Now().In(loc).Add(24 * time.Hour)
	nextDayMorning := time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 9, 0, 0, 0, loc)

	// Сохраняем время в формате "дд.мм.гггг чч:мм"
	userInfo.OrderTime = nextDayMorning.Format("02.01.2006 15:04")
	sendUpdatedInfoToAdmin(bot, chatID, *userInfo)

	var successMessage string
	switch lang {
	case "DEU":
		successMessage = fmt.Sprintf("Ihre Bestellzeit wurde gespeichert: %02d.%02d um 09:00!", nextDay.Day(), nextDay.Month())
	case "EN":
		successMessage = fmt.Sprintf("Your order time has been saved: %02d.%02d at 09:00!", nextDay.Day(), nextDay.Month())
	case "UK":
		successMessage = fmt.Sprintf("Ваш час замовлення збережено: %02d.%02d о 09:00!", nextDay.Day(), nextDay.Month())
	case "RU":
		successMessage = fmt.Sprintf("Ваше время заказа сохранено: %02d.%02d в 09:00!", nextDay.Day(), nextDay.Month())
	}

	msg := tgbotapi.NewMessage(chatID, successMessage)
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	bot.Send(msg)
}

// Обработка времени
func handleOrderTime(bot *tgbotapi.BotAPI, chatID int64, input string, lang string, userInfo *UserInfo) {
	loc, _ := time.LoadLocation("Europe/Vienna")
	currentTime := time.Now().In(loc)

	// Проверяем, нажата ли кнопка "Ввести время повторно"
	if strings.Contains(input, "Ввести новое время") || strings.Contains(input, "Enter a new time") ||
		strings.Contains(input, "Ввести новий час") || strings.Contains(input, "Neue Zeit eingeben") {
		sendOrderTimeRequest(bot, chatID, lang) // Отправляем запрос на ввод времени
		return
	}

	// Проверяем, нажата ли кнопка "Получить на следующий день"
	// Проверяем, нажата ли кнопка "Получить на следующий день"
	if strings.Contains(input, "Получить") || strings.Contains(input, "Receive") ||
		strings.Contains(input, "Отримати") || strings.Contains(input, "Erhalten") {
		nextDay := currentTime.AddDate(0, 0, 1)
		nextDayMorning := time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 9, 0, 0, 0, loc)

		// Сохраняем корректное время в формате "дд.мм.гггг чч:мм"
		userInfo.OrderTime = nextDayMorning.Format("02.01.2006 15:04")
		sendUpdatedInfoToAdmin(bot, chatID, *userInfo)

		// Сообщение пользователю о сохранённом времени
		var successMessage string
		switch lang {
		case "DEU":
			successMessage = fmt.Sprintf("Ihre Bestellzeit wurde gespeichert: %02d.%02d um 09:00!", nextDay.Day(), nextDay.Month())
		case "EN":
			successMessage = fmt.Sprintf("Your order time has been saved: %02d.%02d at 09:00!", nextDay.Day(), nextDay.Month())
		case "UK":
			successMessage = fmt.Sprintf("Ваш час замовлення збережено: %02d.%02d о 09:00!", nextDay.Day(), nextDay.Month())
		case "RU":
			successMessage = fmt.Sprintf("Ваше время заказа сохранено: %02d.%02d в 09:00!", nextDay.Day(), nextDay.Month())
		}
		msg := tgbotapi.NewMessage(chatID, successMessage)
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		bot.Send(msg)

		// Отправляем благодарность за заказ
		sendAdminNotification(bot, chatID, lang)
		return

	}

	// Обработка ввода времени вручную (новый формат "дд.мм.гггг чч:мм")
	parsedTime, err := time.ParseInLocation("02.01.2006 15:04", input, loc)
	if err != nil || parsedTime.Before(currentTime) {
		sendInvalidTimeMessage(bot, chatID, lang)
		return
	}

	// Проверка рабочего времени
	if !isWithinWorkingHours(parsedTime) {
		sendStoreClosedOptions(bot, chatID, lang, parsedTime)
		return
	}

	// Если время корректное, сохраняем его и отправляем благодарность
	// Сохраняем время в формате "дд.мм.гггг чч:мм"
	userInfo.OrderTime = parsedTime.Format("02.01.2006 15:04")
	sendUpdatedInfoToAdmin(bot, chatID, *userInfo)

	// Сообщение пользователю о сохранении времени
	var successMessage string
	switch lang {
	case "DEU":
		successMessage = "Ihre Bestellzeit wurde gespeichert!"
	case "EN":
		successMessage = "Your order time has been saved!"
	case "UK":
		successMessage = "Ваш час замовлення збережено!"
	case "RU":
		successMessage = "Ваше время заказа сохранено!"
	}
	msg := tgbotapi.NewMessage(chatID, successMessage)
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	bot.Send(msg)

	// Отправляем благодарность за заказ
	sendAdminNotification(bot, chatID, lang)
}
func sendAdminNotification(bot *tgbotapi.BotAPI, chatID int64, lang string) {
	var message string
	switch lang {
	case "DEU":
		message = "Vielen Dank für Ihre Bestellung und die Wahl von Florgalerie😄! Der Administrator hat Ihre Bestellung erhalten und <b><i>teilt Ihnen den Preis</i></b> für den von Ihnen gewählten Strauß mit. Wir prüfen die Verfügbarkeit der ausgewählten Blumen und stellen Ihren Strauß zusammen. Um einen neuen Auftrag zu erstellen, schreiben Sie '/start'."
	case "EN":
		message = "Thank you for ordering and choosing Florgalerie😄! The administrator has received your order and <b><i>tell you the price</i></b> of the bouquet you've chosen. We are checking the availability of selected flowers and pick a bouquet. To create a new order, write '/start'."
	case "UK":
		message = "Дякую Вам за зроблене замовлення і вибір Florgalerie😄! Адміністратор отримав Ваше замовлення і <b><i>підкаже вам ціну</i></b> обраного Вами букета. Перевіряємо наявність обраних квітів і збираємо букет. Для створення нового замовлення напишіть '/start'."
	case "RU":
		message = "Благодарю Вас за сделанный заказ и выбор Florgalerie😄! Администратор получил Ваш заказ и <b><i>подскажет вам цену</i></b> выбраного Вами букета. Проверяем наличие выбранных цветов и собираем букет. Для создания нового заказа напишите '/start'."

	}
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "HTML"
	bot.Send(msg)

}

func main() {
	bot, err := tgbotapi.NewBotAPI("7605031210:AAGTiIboCT3mxxLO6egJ3Zhkr8LAVcdu6yo")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	port := os.Getenv("PORT")
if port == "" {
	port = "8080" // Значение по умолчанию
}
log.Printf("Starting server on port %s...", port)
http.ListenAndServe(":"+port, nil)


	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	userData := make(map[int64]*UserInfo)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		text := update.Message.Text

		// Обработка сообщений от администратора
		if chatID == AdminID {
			handleAdminMessage(bot, update, userData)
			continue
		}

		if userData[chatID] == nil {
			userData[chatID] = &UserInfo{}
		}

		userInfo := userData[chatID]
		//отправка фоток от пользователя
		if update.Message.Photo != nil {
			photo := update.Message.Photo[len(update.Message.Photo)-1] // Берём самое большое фото
			adminMessage := fmt.Sprintf("📸 Новое фото от пользователя (ID: %d):", chatID)

			// Отправляем текстовое сообщение админу
			bot.Send(tgbotapi.NewMessage(AdminID, adminMessage))

			// Пересылаем фото администратору
			photoMsg := tgbotapi.NewPhoto(AdminID, tgbotapi.FileID(photo.FileID))
			bot.Send(photoMsg)
			continue
		}

		switch {
		case text == "/start":
			msg := tgbotapi.NewMessage(chatID, "Please select your language:")
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("DEU"),
					tgbotapi.NewKeyboardButton("EN"),
					tgbotapi.NewKeyboardButton("UK"),
					tgbotapi.NewKeyboardButton("RU"),
				),
			)
			bot.Send(msg)

		case text == "DEU" || text == "EN" || text == "UK" || text == "RU":
			userInfo.Language = text
			sendGreeting(bot, chatID, text)
			askUserName(bot, chatID, text) // Задаём вопрос про имя

		case userInfo.UserName == "":
			// Пользователь вводит имя или выбирает "Остаться анонимным"
			if strings.Contains(text, "Остаться анонимным") || strings.Contains(text, "Stay anonymous") ||
				strings.Contains(text, "Залишитися анонімним") || strings.Contains(text, "Anonym bleiben") {
				userInfo.UserName = "Анонимный пользователь" // Сохраняем как анонимный

				// Удаляем кнопку "Остаться анонимным"
				msg := tgbotapi.NewMessage(chatID, "Вы решили остаться анонимным.")
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				bot.Send(msg)

				sendBouquetRequest(bot, chatID, userInfo.Language)
			} else {
				// Сохраняем введённое имя
				userInfo.UserName = text

				// Формируем сообщение на выбранном языке
				var greeting string
				switch userInfo.Language {
				case "DEU":
					greeting = fmt.Sprintf("Freut mich, Sie kennenzulernen, %s!", userInfo.UserName)
				case "EN":
					greeting = fmt.Sprintf("Nice to meet you, %s!", userInfo.UserName)
				case "UK":
					greeting = fmt.Sprintf("Приємно познайомитися, %s!", userInfo.UserName)
				case "RU":
					greeting = fmt.Sprintf("Приятно познакомиться, %s!", userInfo.UserName)
				}

				// Удаляем кнопку "Остаться анонимным"
				msg := tgbotapi.NewMessage(chatID, greeting)
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				bot.Send(msg)

				sendBouquetRequest(bot, chatID, userInfo.Language)
			}

		case userInfo.Bouquet == "":
			userInfo.Bouquet = text
			sendOrderTimeRequest(bot, chatID, userInfo.Language)

		case userInfo.OrderTime == "":
			handleOrderTime(bot, chatID, text, userInfo.Language, userInfo)

		case userInfo.OrderTime != "":
			// Новое сообщение от пользователя
			adminMessage := fmt.Sprintf(
				" Новое сообщение от пользователя %d:\n\n📝 Имя: %s\n\n🗨️ Ваш предыдущий ответ:\n%s\n\n📝 Ответ пользователя (ID: %d):\n%s",
				chatID, userInfo.UserName, userInfo.LastAdminMessage, chatID, text,
			)
			bot.Send(tgbotapi.NewMessage(AdminID, adminMessage))
			}
	}
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}
