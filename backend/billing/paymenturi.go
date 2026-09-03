package billing

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"math/big"
	"strconv"
	"strings"

	"comics-galore/backend/nowpayments"

	qrcode "github.com/skip2/go-qrcode"
)

// XRPL base58 alphabet (differs from the Bitcoin alphabet).
const xrplBase58Alphabet = "rpshnaf39wBUDNEGHJKLM4PQRST7VWXYZ2bcdeCg65jkm8oFqi1tuvAxyz"

func base58Decode(s string) []byte {
	n := new(big.Int)
	base := big.NewInt(58)
	for i := 0; i < len(s); i++ {
		d := strings.IndexByte(xrplBase58Alphabet, s[i])
		if d < 0 {
			return nil
		}
		n.Mul(n, base)
		n.Add(n, big.NewInt(int64(d)))
	}
	b := n.Bytes()
	for i := 0; i < len(s) && s[i] == xrplBase58Alphabet[0]; i++ {
		b = append([]byte{0}, b...)
	}
	return b
}

func base58Encode(b []byte) string {
	n := new(big.Int).SetBytes(b)
	base := big.NewInt(58)
	mod := new(big.Int)
	var out []byte
	for n.Sign() > 0 {
		n.DivMod(n, base, mod)
		out = append(out, xrplBase58Alphabet[mod.Int64()])
	}
	for _, v := range b {
		if v != 0 {
			break
		}
		out = append(out, xrplBase58Alphabet[0])
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return string(out)
}

func sha256Twice(b []byte) []byte {
	h1 := sha256.Sum256(b)
	h2 := sha256.Sum256(h1[:])
	return h2[:]
}

// encodeXAddress encodes a classic Ripple address + destination tag into an
// X-address (mainnet), matching the XRPL reference implementation:
// [0x05, 0x44] + accountID(20) + tag-flag/bytes(9) + checksum(4), base58check.
func encodeXAddress(classicAddress string, tag uint64) (string, error) {
	decoded := base58Decode(classicAddress)
	if len(decoded) != 25 || decoded[0] != 0x00 {
		return "", strconv.ErrSyntax
	}
	if !bytes.Equal(sha256Twice(decoded[:21])[:4], decoded[21:25]) {
		return "", strconv.ErrSyntax
	}
	accountID := decoded[1:21]

	payload := []byte{0x05, 0x44}
	payload = append(payload, accountID...)
	if tag == 0 {
		payload = append(payload, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	} else {
		if tag > 0xFFFFFFFF {
			return "", strconv.ErrRange
		}
		payload = append(payload, 1,
			byte(tag&0xff), byte((tag>>8)&0xff), byte((tag>>16)&0xff), byte((tag>>24)&0xff),
			0, 0, 0, 0)
	}

	checked := append(payload, sha256Twice(payload)[:4]...)
	return base58Encode(checked), nil
}

func formatAmount(amount float64) string {
	if amount <= 0 {
		return ""
	}
	return strconv.FormatFloat(amount, 'f', -1, 64)
}

func amountParam(amount float64) string {
	if a := formatAmount(amount); a != "" {
		return "?amount=" + a
	}
	return ""
}

func weiValue(eth float64) string {
	f := new(big.Float).SetFloat64(eth)
	f.Mul(f, big.NewFloat(1e18))
	i, _ := f.Int(nil)
	return i.String()
}

// buildPaymentURI builds a scheme-aware payment URI so a wallet scanning the QR
// code recognises the network (and, where supported, pre-fills the amount).
func buildPaymentURI(network, payCurrency, address, payinExtraID string, amount float64) string {
	network = strings.ToLower(strings.TrimSpace(network))
	payCurrency = strings.ToLower(strings.TrimSpace(payCurrency))
	if address == "" {
		return ""
	}

	switch network {
	case "bitcoin":
		return "bitcoin:" + address + amountParam(amount)
	case "litecoin":
		return "litecoin:" + address + amountParam(amount)
	case "ethereum":
		if payCurrency == "eth" {
			return "ethereum:" + address + "?value=" + weiValue(amount)
		}
		return "ethereum:" + address
	case "solana":
		return "solana:" + address + amountParam(amount)
	case "tron":
		return "tron:" + address
	case "xrp":
		// X-address bundles the destination tag so the sender never has to type
		// one. X-addresses are self-identifying (mainnet prefix "X"), so no URI
		// scheme is needed.
		if payinExtraID != "" {
			if tag, err := strconv.ParseUint(payinExtraID, 10, 64); err == nil {
				if x, err := encodeXAddress(address, tag); err == nil {
					return x
				}
			}
		}
		return address
	case "stellar":
		u := "web+stellar:pay?destination=" + address
		if payinExtraID != "" {
			u += "&memo_type=id&memo=" + payinExtraID
		}
		if a := formatAmount(amount); a != "" {
			u += "&amount=" + a
		}
		return u
	default:
		return address
	}
}

// generateQRDataURL renders a QR code PNG (data URL) for the given content.
func generateQRDataURL(content string) (string, error) {
	png, err := qrcode.Encode(content, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png), nil
}

// buildDepositQR builds the payment URI + QR data URL for a deposit response.
func buildDepositQR(npResp *nowpayments.DepositResponse) (qrDataURL, paymentURI string) {
	uri := buildPaymentURI(npResp.Network, npResp.PayCurrency, npResp.PayAddress, npResp.PayinExtraID, npResp.PayAmount)
	qr, err := generateQRDataURL(uri)
	if err != nil {
		return "", uri
	}
	return qr, uri
}
