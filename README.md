# Insta-Scraper Telegram Bot

A Go-based Telegram bot that provides direct download links for Instagram videos (currently supports videos only). Live bot: [@igramloadbot](https://t.me/igramloadbot)

## Features

- Direct video URL generation from Instagram posts/reels
- Inline query support 
- Fast and lightweight
- Requires no client-side installation

## Setup Instructions

1. **Clone the repository**
```bash
git clone https://github.com/<your-username>/insta-scraper.git
cd insta-scraper
```

2. **Configure secrets**
Create `config/secrets.json` with:
```json
{
  "bot_token": "YOUR_TELEGRAM_BOT_TOKEN",
  "chat_id": YOUR_CHAT_ID
}
```

3. **Build and run**
```bash
go mod tidy
go build -o insta-scraper
./insta-scraper
```

## Features & Limitations

✓ Video downloads from:
- Instagram Reels
- Public posts
- Stories (public accounts only)

✗ Currently doesn't support:
- Photo posts 
- Private accounts
- IGTV videos

## Deployment

The bot can be deployed using:
- Docker (see Dockerfile)
- Direct binary execution
- Cloud platforms (AWS, GCP, Azure)

## Contributing

Pull requests are welcome! Please ensure:
- Proper error handling 
- Tests for new features
- Backwards compatibility

## License

MIT
