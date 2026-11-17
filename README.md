# Max QR Code — Go SDK

Official **Go** client for the [Max QR Code](https://www.getmaxqr.com) API.

- Static QR **generate** (PNG / SVG / PDF) + multipart logo upload  
- **Dynamic** short links (`go.getmaxqr.com/r/:slug`)  
- **Bulk** CSV/rows → poll → ZIP download  

Module: `github.com/MichaelGenesisII/Max-QR-Code-Go-SDK`  
API keys (`mq_live_…`) are **Business** plan. Docs: [getmaxqr.com/docs](https://www.getmaxqr.com/docs)

## Install

```bash
go get github.com/MichaelGenesisII/Max-QR-Code-Go-SDK@v0.1.0
```

Requires **Go 1.21+**.

## Quickstart — generate a PNG

```go
package main

import (
	"context"
	"encoding/base64"
	"os"

	getmaxqr "github.com/MichaelGenesisII/Max-QR-Code-Go-SDK"
)

func main() {
	client := getmaxqr.New(getmaxqr.Options{APIKey: os.Getenv("MAXQR_API_KEY")})
	out, err := client.Generate(context.Background(), getmaxqr.URL("https://example.com"), getmaxqr.GenerateOptions{
		Format: "png",
		Style: map[string]any{
			"foreground": "#111111",
			"background": "#ffffff",
			"size":       1024,
		},
	})
	if err != nil {
		panic(err)
	}
	qr := out.(getmaxqr.GenerateJSONResult)
	png, _ := base64.StdEncoding.DecodeString(qr.DataBase64)
	_ = os.WriteFile("code.png", png, 0o644)
}
```

## Generate with a center logo

Call the API host directly for large logos (not via a serverless proxy):

```go
logo, _ := os.ReadFile("./logo.png")
scale := 0.22
out, err := client.GenerateWithLogo(context.Background(), getmaxqr.URL("https://example.com"), getmaxqr.LogoInput{
	Data: logo, Filename: "logo.png", Mime: "image/png",
}, getmaxqr.GenerateOptions{Format: "png", LogoScale: &scale})
```

## Dynamic codes

```go
ws := os.Getenv("MAXQR_WORKSPACE_ID")
result, err := client.Dynamic.Create(context.Background(), ws, getmaxqr.CreateDynamicInput{
	Name:           "Table tent",
	DestinationURL: "https://example.com/menu",
})
code := result["code"].(map[string]any)
fmt.Println(code["redirect_url"]) // https://go.getmaxqr.com/r/…
```

## Bulk job → ZIP

```go
created, _ := client.Bulk.Create(context.Background(), ws, getmaxqr.CreateBulkInput{
	ContentType: "url",
	Format:      "png",
	Rows: []map[string]any{
		{"url": "https://example.com/a", "name": "A"},
		{"url": "https://example.com/b", "name": "B"},
	},
})
jobID := created["job"].(map[string]any)["id"].(string)
_, _ = client.Bulk.Wait(context.Background(), ws, jobID, getmaxqr.WaitOptions{})
zipBytes, _ := client.Bulk.DownloadZip(context.Background(), ws, jobID)
_ = os.WriteFile("codes.zip", zipBytes, 0o644)
```

## Content helpers

```go
getmaxqr.URL("https://example.com")
getmaxqr.Text("Hello")
getmaxqr.Wifi("Cafe", "secret", "WPA", nil)
getmaxqr.Email("hi@example.com", "Hi", "")
getmaxqr.Tel("+15550100")
getmaxqr.SMS("+15550100", "Hello")
getmaxqr.Geo(1.5, -2.25)
getmaxqr.Payment("https://buy.stripe.com/…")
getmaxqr.VCard(map[string]any{"first_name": "Ada", "last_name": "Lovelace"})
```

## Errors

```go
_, err := client.GenerateOptions(context.Background())
if apiErr, ok := err.(*getmaxqr.Error); ok {
	// rate_limited | daily_quota | plan_forbidden | insufficient_scope | …
	fmt.Println(apiErr.Code, apiErr.Status, apiErr.RetryAfterSeconds, apiErr.RequestID)
}
```

## License

MIT
