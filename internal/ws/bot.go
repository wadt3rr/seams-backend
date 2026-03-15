package ws

import "strings"

func GetBotReply(text string) string {

	msg := strings.ToLower(text)

	switch {

	case strings.Contains(msg, "привет"):
		return "Здравствуйте! Чем могу помочь?"

	case strings.Contains(msg, "пароль"):
		return "Вы можете восстановить пароль на странице восстановления."

	case strings.Contains(msg, "заказ"):
		return "Информацию о заказах можно посмотреть в разделе отслеживания заказов."

	default:
		return "Я не смог найти ответ. Попробуйте уточнить вопрос."
	}
}
