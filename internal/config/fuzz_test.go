package config_test

import (
	"testing"

	"github.com/h13/gtm-users/internal/config"
)

func FuzzParse(f *testing.F) {
	f.Add([]byte(`
account_id: "123"
mode: additive
users:
  - email: a@b.com
    account_access: user
`))
	f.Add([]byte(`
account_id: "999"
mode: authoritative
users:
  - email: test@example.com
    account_access: admin
    container_access:
      - container_id: "GTM-AAAA1111"
        permission: publish
`))
	f.Add([]byte(`{}`))
	f.Add([]byte(``))

	f.Fuzz(func(t *testing.T, data []byte) {
		cfg, err := config.Parse(data)
		if err != nil {
			return
		}
		_ = config.Validate(cfg)
	})
}
