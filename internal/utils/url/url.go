package url_utils

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	qrcode "github.com/skip2/go-qrcode"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateSlug(len int) (string, error) {
	slug := make([]byte, len)

	for i := range slug {
		a, err := rand.Int(rand.Reader, big.NewInt(62))

		if err != nil {
			return "", err
		}

		slug[i] = alphabet[a.Int64()]
	}

	return string(slug), nil
}

func ValidateUrl(rawUrl string) error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	var validateUrl struct {
		Url string `validate:"required,max=2048,url"`
	}

	validateUrl.Url = rawUrl

	err := validate.Struct(validateUrl)

	if err != nil {
		return err
	}

	u, err := url.ParseRequestURI(rawUrl)

	if err != nil {
		return err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("invalid url")
	}

	if strings.HasPrefix(rawUrl, "localhost:8080/") {
		return errors.New("invalid url")
	}

	return nil
}

func ValidateSlug(slug string) error {
	if len(slug) != 7 {
		return errors.New("invalid slug length")
	}

	for i := range slug {
		if !strings.Contains(alphabet, string(slug[i])) {
			return errors.New("invalid slug format: " + string(slug[i]))
		}
	}

	return nil
}

func GetIP(r *http.Request) string {
	hostPort := r.RemoteAddr
	host, _, err := net.SplitHostPort(hostPort)

	if err != nil {
		return hostPort
	}

	return host
}

func GenerateQrCode(shortUrl string) ([]byte, error) {
	var data []byte

	data, err := qrcode.Encode(shortUrl, qrcode.Medium, 256)

	return data, err
}

func BuildClicksArgs(clicks map[string]int64) (string, []any) {
	query := "UPDATE urls SET clicks = urls.clicks + data.added FROM (VALUES "
	args := []any{}
	argCounter := 0
	argStr := ""

	for k, v := range clicks {
		argStr = fmt.Sprintf("($%d, $%d::bigint),", argCounter+1, argCounter+2)
		query += argStr

		args = append(args, k, v)

		argCounter += 2
	}

	query = strings.TrimSuffix(query, ",")
	query += ") AS data(slug, added) WHERE urls.slug = data.slug"

	return query, args
}

func ConvertToInt64(clicks map[string]string) (map[string]int64, error) {
	mp := make(map[string]int64)

	for k, v := range clicks {
		n, err := (strconv.Atoi(v))

		if err != nil {
			return nil, err
		}

		num := int64(n)

		mp[k] = num
	}

	return mp, nil
}
