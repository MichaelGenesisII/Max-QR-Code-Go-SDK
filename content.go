package getmaxqr

// Content is a typed QR payload for generate / bulk.
type Content struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

func URL(url string) Content {
	return Content{Type: "url", Data: map[string]any{"url": url}}
}

func Text(text string) Content {
	return Content{Type: "text", Data: map[string]any{"text": text}}
}

func Wifi(ssid string, password, security string, hidden *bool) Content {
	data := map[string]any{"ssid": ssid}
	if password != "" {
		data["password"] = password
	}
	if security != "" {
		data["security"] = security
	}
	if hidden != nil {
		data["hidden"] = *hidden
	}
	return Content{Type: "wifi", Data: data}
}

func Email(to, subject, body string) Content {
	data := map[string]any{}
	if to != "" {
		data["to"] = to
	}
	if subject != "" {
		data["subject"] = subject
	}
	if body != "" {
		data["body"] = body
	}
	return Content{Type: "email", Data: data}
}

func Tel(number string) Content {
	return Content{Type: "tel", Data: map[string]any{"number": number}}
}

func SMS(number, body string) Content {
	data := map[string]any{"number": number}
	if body != "" {
		data["body"] = body
	}
	return Content{Type: "sms", Data: data}
}

func Geo(lat, lng float64) Content {
	return Content{Type: "geo", Data: map[string]any{"lat": lat, "lng": lng}}
}

func Payment(link string) Content {
	return Content{Type: "payment", Data: map[string]any{"link": link}}
}

func VCard(data map[string]any) Content {
	if data == nil {
		data = map[string]any{}
	}
	return Content{Type: "vcard", Data: data}
}
