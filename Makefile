service-start:
	@docker compose up
service-stop:
	@docker compose stop
server:
	@air
styles:
	@npx tailwindcss -i ./config/tailwind.css -o ./public/styles/styles.css --watch

.PHONY: service-start service-stop server styles