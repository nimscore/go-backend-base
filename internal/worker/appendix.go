package worker

// message, err := mailpkg.MessageFromJson(value)
// if err != nil {
// 	this.logger.Error("error unmarshalling message", zap.Error(err))
// 	continue
// }

// path := ""
// switch kind {
// case mailpkg.KIND_MAIL_CONFIRM:
// 	path = "mail_confirm.html"
// case mailpkg.KIND_MAIL_RECOVER:
// 	path = "mail_recover.html"
// }

// if path == "" {
// 	this.logger.Error("unknown message kind")
// 	continue
// }

// content, err := templatepkg.Render(fmt.Sprintf("template/%s", path), message.Arguments)
// if err != nil {
// 	this.logger.Error("error rendering mail", zap.Error(err))
// 	continue
// }

// err = this.mailClient.SendHTML(
// 	message.From,
// 	message.To,
// 	message.Subject,
// 	content,
// )
// if err != nil {
// 	this.logger.Error("error sending mail", zap.Error(err))
// 	continue
// }

// this.logger.Info(
// 	"mail sent to recipient",
// 	zap.String("kind", kind),
// 	zap.String("from", message.From),
// 	zap.String("to", message.To),
// )
