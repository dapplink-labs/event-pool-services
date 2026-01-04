package totpGenerate

import (
	"bytes"
	"encoding/base64"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"image/png"
	"log"
)

func GenerateTOTP(account string) (key *otp.Key, imgReturnBase64 string, err error) {
	key, err = totp.Generate(totp.GenerateOpts{
		Issuer:      "MyApp",
		AccountName: account,
	})
	if err != nil {
		log.Fatal(err)
	}

	img, err := key.Image(256, 256)
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		log.Fatal(err)
	}

	base64Img := base64.StdEncoding.EncodeToString(buf.Bytes())
	return key, base64Img, nil
}

func ValidateTOTP(secret string, userCode string) bool {
	return totp.Validate(userCode, secret)
}
