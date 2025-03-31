# Instagram/YouTube Media Downloader Bot

A Telegram bot that downloads and sends Instagram/YouTube media (photos/videos) through inline queries or direct messages.

## Features
- Download Instagram and YouTube media
- Supports both photos and videos
- Inline query interface for easy sharing
- Rate limiting to prevent abuse
- Automatic cleanup of downloaded files

## Setup

1. Clone the repository:
```bash
git clone https://github.com/yourusername/insta-scraper.git
cd insta-scraper
```

2. Install dependencies:
```bash
pip install -r requirements.txt
```

3. Create `secrets.json` (see below)

## Configuration

Create `secrets.json` based on the template:
```json
{
    "bot_token": "YOUR_TELEGRAM_BOT_TOKEN",
    "chat_id": "YOUR_TELEGRAM_CHAT_ID"
}
```

## Known Issues
- Instagram photos may be cropped in preview (Instagram API limitation)
- Some private/unavailable content may not download
- Large videos may timeout (>20MB)

## Usage

1. Start the bot:
```bash
python scraper.py
```

2. In Telegram:
- Send direct message with Instagram/YouTube link
- Or use inline mode: `@YourBotName instagram.com/p/...`

## Secrets Template

`secrets.template.json`:
```json
{
    "bot_token": "YOUR_BOT_TOKEN_FROM_BOTFATHER",
    "chat_id": "YOUR_CHAT_ID_FOR_UPLOADS"
}
```

> **Warning**: Never commit your actual `secrets.json` to version control!

## Troubleshooting

**Cropped Instagram Photos**  
This is currently a limitation of Instagram's API. Full-resolution images may require additional scraping methods.

**Errors**  
Check `scraper.log` for detailed error messages.

## Requirements

See `requirements.txt` for complete dependency list.
