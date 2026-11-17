package getmaxqr

import "testing"

func TestContentHelpers(t *testing.T) {
	u := URL("https://example.com")
	if u.Type != "url" || u.Data["url"] != "https://example.com" {
		t.Fatalf("URL helper: %+v", u)
	}
	tel := Tel("+15550100")
	if tel.Type != "tel" || tel.Data["number"] != "+15550100" {
		t.Fatalf("Tel helper: %+v", tel)
	}
	pay := Payment("https://buy.stripe.com/x")
	if pay.Type != "payment" || pay.Data["link"] != "https://buy.stripe.com/x" {
		t.Fatalf("Payment helper: %+v", pay)
	}
}
