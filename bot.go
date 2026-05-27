package main

import (
	"log"
	"os"
	"strings"
	"unicode"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	token := resolveBotToken()
	if token == "" {
		log.Fatal("token do bot nao encontrado.\nDefina uma variavel de ambiente (token/TOKEN/BOT_TOKEN/TELEGRAM_BOT_TOKEN) ou crie um arquivo .env com TOKEN=SEU_TOKEN.\nPowerShell:\n  $env:token=\"SEU_TOKEN\"; go run .\nCMD:\n  set token=SEU_TOKEN && go run .")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("erro ao criar bot: %v", err)
	}

	if _, err := bot.Request(tgbotapi.DeleteWebhookConfig{DropPendingUpdates: true}); err != nil {
		log.Printf("aviso ao remover webhook: %v", err)
	}

	if err := registerBotCommands(bot); err != nil {
		log.Printf("aviso ao registrar comandos: %v", err)
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.GetUpdatesChan(updateConfig)

	log.Printf("bot iniciado como @%s (modo: long polling; confira se so ha UMA instancia rodando)", bot.Self.UserName)

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
		cmd := commandName(update.Message)
		log.Printf("mensagem recebida chat_id=%d texto=%q cmd=%q", update.Message.Chat.ID, text, cmd)

		var response string
		var markup interface{} = buildMainMenu()
		switch {
		case cmd == "start":
			response = "👋 Bem-vindo ao <b>Validador de CPF</b>!\n\nEscolha uma opcao no menu para comecar.\n\n" + fallbackMenuText()
		case isValidateRequest(text) || cmd == "validar":
			response = "🔎 Envie um CPF (com ou sem pontuacao) para eu validar."
		case isHelpRequest(text) || cmd == "ajuda":
			response = "🧭 Centro de ajuda\nEscolha uma opcao abaixo:\n\n" + fallbackMenuText()
			markup = buildInlineMenu()
		case isAboutRequest(text) || cmd == "sobre":
			response = "📘 Sobre este bot\nEscolha uma opcao abaixo:\n\n" + fallbackMenuText()
			markup = buildInlineMenu()
		case isCPF(text):
			response = "✅ CPF <b>valido</b>."
		default:
			response = "❌ CPF <b>invalido</b>.\nToque em <b>✅ Validar CPF</b> no menu e tente novamente.\n\n" + fallbackMenuText()
		}

		sendReply(bot, update.Message.Chat.ID, response, markup)
	}
}

func normalizeMenuText(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(value))
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
	case "validar", "validar cpf", "validacao", "validacao cpf":
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
		tgbotapi.BotCommand{Command: "start", Description: "Iniciar o bot"},
		tgbotapi.BotCommand{Command: "validar", Description: "Validar um CPF"},
		tgbotapi.BotCommand{Command: "ajuda", Description: "Central de ajuda"},
		tgbotapi.BotCommand{Command: "sobre", Description: "Sobre o bot"},
	)
	_, err := bot.Request(cfg)
	return err
}

func sendReply(bot *tgbotapi.BotAPI, chatID int64, text string, markup interface{}) {
	log.Printf("enviando resposta para chat_id=%d", chatID)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = markup
	if _, err := bot.Send(msg); err != nil {
		log.Printf("erro ao enviar mensagem (HTML): %v", err)
		plain := tgbotapi.NewMessage(chatID, text)
		plain.ReplyMarkup = markup
		if _, err := bot.Send(plain); err != nil {
			log.Printf("erro ao enviar mensagem (texto): %v", err)
		}
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
	keyboard.OneTimeKeyboard = false
	keyboard.Selective = false
	return keyboard
}

func buildInlineMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Como validar", "help_validate"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💡 Ver exemplo", "help_example"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🤖 Sobre o bot", "about_bot"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("↩️ Menu principal", "back_menu"),
		),
	)
}

func fallbackMenuText() string {
	return "• ✅ Validar CPF\n• ❓ Ajuda\n• ℹ️ Sobre"
}

func handleCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	var text string
	switch callback.Data {
	case "menu_validate":
		text = "🔎 Envie um CPF (com ou sem pontuacao) para eu validar."
	case "menu_help":
		text = "🧭 Centro de ajuda\nEscolha uma opcao abaixo:"
	case "menu_about":
		text = "📘 Sobre este bot\nEscolha uma opcao abaixo:"
	case "help_validate":
		text = "✅ Toque em <b>Validar CPF</b> e envie o numero.\nAceito com ou sem pontuacao."
	case "help_example":
		text = "📝 Exemplo de CPF para teste:\n<code>529.982.247-25</code>"
	case "about_bot":
		text = "🤖 Sou um bot de validacao de CPF com menu interativo para facilitar seu uso."
	case "back_menu":
		text = "🏠 Voce voltou ao menu principal."
	default:
		text = "⚠️ Opcao nao reconhecida."
	}

	alert := tgbotapi.NewCallback(callback.ID, "Opcao selecionada!")
	if _, err := bot.Request(alert); err != nil {
		log.Printf("erro ao responder callback: %v", err)
	}

	markup := interface{}(buildMainMenu())
	if callback.Data == "menu_help" || callback.Data == "menu_about" {
		markup = buildInlineMenu()
	}
	sendReply(bot, callback.Message.Chat.ID, text, markup)
}

func isCPF(cpf string) bool {
	cpf = onlyDigits(cpf)
	if len(cpf) != 11 {
		return false
	}

	// Rejeita CPFs com todos os digitos iguais (ex: 11111111111).
	if allDigitsEqual(cpf) {
		return false
	}

	firstCheck := cpfCheckDigit(cpf[:9], 10)
	secondCheck := cpfCheckDigit(cpf[:10], 11)

	return cpf[9] == byte(firstCheck+'0') && cpf[10] == byte(secondCheck+'0')
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