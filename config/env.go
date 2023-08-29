package config

// based on https://docs.gitlab.com/ee/ci/variables/predefined_variables.html

type EnvVariables struct {
	JwtPublicKey        string `env:"JWT_PUBLIC_KEY"`
	WebhookSecret       string `env:"WEBHOOK_SECRET"`
	DatabaseUrl         string `env:"DATABASE_URL"`
	GitHubAppPrivateKey string `env:"GITHUB_APP_PRIVATE_KEY"`
}
