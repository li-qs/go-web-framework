package utils

import "encoding/base64"

var base64URLEncoder = base64.URLEncoding.WithPadding(base64.NoPadding)

func Base64URLEncode(username string) string {
	return base64URLEncoder.EncodeToString([]byte(username))
}

func Base64URLDecode(encoded string) (string, error) {
	data, err := base64URLEncoder.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
