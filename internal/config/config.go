package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	FacebookGroupID string
	FacebookCUser   string
	FacebookXS      string
}

func (c *Config) GroupURL() string {
	return "https://www.facebook.com/groups/" + c.FacebookGroupID
}

func (c *Config) Cookie() string {
	return fmt.Sprintf("c_user=%s; xs=%s;", c.FacebookCUser, c.FacebookXS)
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	cfg := &Config{
		FacebookGroupID: strings.TrimSpace(os.Getenv("FACEBOOK_GROUPS_ID")),
		FacebookCUser:   strings.TrimSpace(os.Getenv("FACEBOOK_CUSER")),
		FacebookXS:      strings.TrimSpace(os.Getenv("FACEBOOK_XS")),
	}

	if cfg.FacebookGroupID == "" {
		return nil, fmt.Errorf("FACEBOOK_GROUPS_ID is required but not set")
	}
	if cfg.FacebookCUser == "" {
		return nil, fmt.Errorf("FACEBOOK_CUSER is required but not set")
	}
	if cfg.FacebookXS == "" {
		return nil, fmt.Errorf("FACEBOOK_XS is required but not set")
	}

	return cfg, nil
}
