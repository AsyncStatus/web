set-app-prod-secrets-from-env:
	cat ./web-app/.env.production | fly secrets import -a as-app

set-app-dev-secrets-from-env:
	cat ./web-app/.env.development | fly secrets import -a dev-as-app

set-landing-prod-secrets-from-env:
	cat ./web-landing/.env.production | fly secrets import -a as-landing

set-landing-dev-secrets-from-env:
	cat ./web-landing/.env.development | fly secrets import -a dev-as-landing
