package main

import (
	"log"
	"os"
	"strings"
	"unicode"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func resolveBotToken() string {
	for _, key := range []string{"token", "TOKEN", "BOT_TOKEN", "TELEGRAM_BOT_TOKEN"} {
		if v, ok := os.LookupEnv(key); ok {
			v = strings.TrimSpace(v)
			if v != "" {
				return v
			}
		}
	}

	data, err := os.ReadFile(".env")
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		v = strings.Trim(v, `"'`)

		if k == "TOKEN" && v != "" {
			return v
		}
	}

	return ""
}

func main() {
	token := resolveBotToken()

	if token == "" {
		log.Fatal("token do bot nao encontrado")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("erro ao criar bot: %v", err)
	}

	if _, err := bot.Request(tgbotapi.DeleteWebhookConfig{
		DropPendingUpdates: true,
	}); err != nil {
		log.Printf("erro ao remover webhook: %v", err)
	}

	if err := registerBotCommands(bot); err != nil {
		log.Printf("erro ao registrar comandos: %v", err)
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	log.Printf("bot iniciado como @%s", bot.Self.UserName)

	for update := range updates {

		if update.CallbackQuery != nil {
			handleCallback(bot, update.CallbackQuery)
			continue
		}

		if update.Message == nil {
			continue
		}

		text := strings.TrimSpace(update.Message.Text)

		if text == "" {
			continue
		}

		cpfDigits := onlyDigits(text)

		if len(cpfDigits) == 11 {

			valid := isCPF(text)

			if valid {
				sendReply(
					bot,
					update.Message.Chat.ID,
					"✅ CPF <b>valido</b>.",
					buildMainMenu(),
				)
			} else {
				sendReply(
					bot,
					update.Message.Chat.ID,
					"❌ CPF <b>invalido</b>.",
					buildMainMenu(),
				)
			}

			continue
		}

		cmd := commandName(update.Message)

		var response string
		var markup interface{} = buildMainMenu()

		switch {

		case cmd == "start":
			response = "👋 Bem-vindo ao bot validador de CPF."

		case isValidateRequest(text) || cmd == "validar":
			response = "🔎 Envie um CPF para validar."

		case isHelpRequest(text) || cmd == "ajuda":
			response = "🧭 Centro de ajuda."
			markup = buildInlineMenu()

		case isAboutRequest(text) || cmd == "sobre":
			response = "📘 Bot de validação de CPF."
			markup = buildInlineMenu()

		default:
			response = "❌ Mensagem invalida."
		}

		sendReply(
			bot,
			update.Message.Chat.ID,
			response,
			markup,
		)
	}
}

func normalizeMenuText(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))

	if value == "" {
		return ""
	}

	var b strings.Builder

	space := false

	for _, r := range value {

		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			space = false
			continue
		}

		if !space {
			b.WriteByte(' ')
			space = true
		}
	}

	return strings.TrimSpace(b.String())
}

func isValidateRequest(text string) bool {
	switch normalizeMenuText(text) {

	case "validar", "validar cpf":
		return true

	default:
		return false
	}
}

func isHelpRequest(text string) bool {
	switch normalizeMenuText(text) {

	case "ajuda", "help":
		return true

	default:
		return false
	}
}

func isAboutRequest(text string) bool {
	switch normalizeMenuText(text) {

	case "sobre", "about":
		return true

	default:
		return false
	}
}

func commandName(msg *tgbotapi.Message) string {

	if msg == nil {
		return ""
	}

	if msg.IsCommand() {
		return strings.ToLower(strings.TrimSpace(msg.Command()))
	}

	text := strings.TrimSpace(msg.Text)

	if !strings.HasPrefix(text, "/") {
		return ""
	}

	part := strings.Fields(text)[0]

	part = strings.TrimPrefix(part, "/")

	if idx := strings.IndexByte(part, '@'); idx >= 0 {
		part = part[:idx]
	}

	return strings.ToLower(part)
}

func registerBotCommands(bot *tgbotapi.BotAPI) error {

	cfg := tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{
			Command:     "start",
			Description: "Iniciar o bot",
		},
		tgbotapi.BotCommand{
			Command:     "validar",
			Description: "Validar CPF",
		},
		tgbotapi.BotCommand{
			Command:     "ajuda",
			Description: "Ajuda",
		},
		tgbotapi.BotCommand{
			Command:     "sobre",
			Description: "Sobre",
		},
	)

	_, err := bot.Request(cfg)

	return err
}

func sendReply(
	bot *tgbotapi.BotAPI,
	chatID int64,
	text string,
	markup interface{},
) {

	msg := tgbotapi.NewMessage(chatID, text)

	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = markup

	_, err := bot.Send(msg)

	if err != nil {
		log.Printf("erro ao enviar mensagem: %v", err)
	}
}

func buildMainMenu() tgbotapi.ReplyKeyboardMarkup {

	keyboard := tgbotapi.NewReplyKeyboard(

		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✅ Validar CPF"),
		),

		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❓ Ajuda"),
			tgbotapi.NewKeyboardButton("ℹ️ Sobre"),
		),
	)

	keyboard.ResizeKeyboard = true

	return keyboard
}

func buildInlineMenu() tgbotapi.InlineKeyboardMarkup {

	return tgbotapi.NewInlineKeyboardMarkup(

		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"📋 Como validar",
				"help_validate",
			),
		),

		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🤖 Sobre",
				"about_bot",
			),
		),
	)
}

func handleCallback(
	bot *tgbotapi.BotAPI,
	callback *tgbotapi.CallbackQuery,
) {

	var text string

	switch callback.Data {

	case "help_validate":
		text = "Envie um CPF com ou sem pontuação."

	case "about_bot":
		text = "Bot validador de CPF."

	default:
		text = "Opcao invalida."
	}

	sendReply(
		bot,
		callback.Message.Chat.ID,
		text,
		buildMainMenu(),
	)
}

func isCPF(cpf string) bool {

	cpf = onlyDigits(cpf)

	if len(cpf) != 11 {
		return false
	}

	if allDigitsEqual(cpf) {
		return false
	}

	firstCheck := cpfCheckDigit(cpf[:9], 10)
	secondCheck := cpfCheckDigit(cpf[:10], 11)

	return cpf[9] == byte(firstCheck+'0') &&
		cpf[10] == byte(secondCheck+'0')
}

func onlyDigits(value string) string {

	var b strings.Builder

	for _, r := range value {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}

	return b.String()
}

func allDigitsEqual(cpf string) bool {

	first := cpf[0]

	for i := 1; i < len(cpf); i++ {
		if cpf[i] != first {
			return false
		}
	}

	return true
}

func cpfCheckDigit(base string, weightStart int) int {

	sum := 0
	weight := weightStart

	for i := 0; i < len(base); i++ {
		sum += int(base[i]-'0') * weight
		weight--
	}

	rest := sum % 11

	if rest < 2 {
		return 0
	}

	return 11 - rest
}